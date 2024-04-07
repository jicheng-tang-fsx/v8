package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var reMsgEx, reMsgCorrect, reMsgFill, reMsgFix *regexp.Regexp
var reTime, reClOrderId *regexp.Regexp

func init() {
	reMsgEx, _ = regexp.Compile(`\|35=8\|`)
	reMsgCorrect, _ = regexp.Compile(`\|20=2\|`)
	reMsgFill, _ = regexp.Compile(`\|39=2\|`)
	reMsgFix, _ = regexp.Compile(`8=FIX`)
	reTime = regexp.MustCompile(`^D\d{4} (\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}\.\d{6})`)
	reClOrderId = regexp.MustCompile(`\|11=([^|]+)\|`)
}

type JnetConfirmedOrder struct {
	ClOrderId  string
	RecvTime   string
	ReturnTime string
	CostTime   string
}

func isJnetConfirmed(line string) bool {
	if reMsgEx.MatchString(line) && reMsgCorrect.MatchString(line) && reMsgFill.MatchString(line) && reMsgFix.MatchString(line) {
		return true
	}
	return false
}

func parseLine(line string) (JnetConfirmedOrder, error) {
	returnTimeMatches := reTime.FindStringSubmatch(line)
	clOrderIdMatches := reClOrderId.FindStringSubmatch(line)

	order := JnetConfirmedOrder{}

	if len(returnTimeMatches) > 1 {
		order.ReturnTime = returnTimeMatches[1]
	} else {
		return JnetConfirmedOrder{}, fmt.Errorf("ReturnTime not found")
	}
	if len(clOrderIdMatches) > 1 {
		order.ClOrderId = clOrderIdMatches[1]
	} else {
		return JnetConfirmedOrder{}, fmt.Errorf("ClOrderId not found")
	}

	return order, nil
}

func getAllJnetConfirmedOrder(filename string) ([]JnetConfirmedOrder, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var orders []JnetConfirmedOrder
	scanner := bufio.NewScanner(file)

	count := 0
	for scanner.Scan() {
		// 替换每一行中的'\x01'为'|'
		line := strings.Replace(scanner.Text(), "\x01", "|", -1)
		if isJnetConfirmed(line) {
			count += 1
			order, err := parseLine(line)
			if err != nil {
				fmt.Printf("parse error: %v\n", err)
				continue // 解析错误时跳过该行
			}
			orders = append(orders, order)
		}
	}

	fmt.Println("JNET Correction Order Count: ", count)

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return orders, nil
}

func exportToJsonl(orders []JnetConfirmedOrder, jsonlFilename string) error {
	file, err := os.Create(jsonlFilename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	for _, order := range orders {
		jsonBytes, err := json.Marshal(order)
		if err != nil {
			return fmt.Errorf("error marshalling to JSON: %v", err)
		}
		_, err = file.Write(jsonBytes)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		// 写入换行符以满足jsonl格式要求
		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing newline to file: %v", err)
		}
	}

	return nil
}

func fillSendTime(orders []JnetConfirmedOrder, filename string) error {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Replace(scanner.Text(), "\x01", "|", -1)
		// 检查每个订单
		for i, order := range orders {
			if order.RecvTime == "" && strings.Contains(line, fmt.Sprintf("|11=%s|", order.ClOrderId)) && strings.Contains(line, "|35=D|") && strings.Contains(line, "8=FIX") {
				// 匹配到订单ID，尝试提取时间
				timeMatches := reTime.FindStringSubmatch(line)
				if len(timeMatches) > 1 {
					// 更新RecvTime
					orders[i].RecvTime = timeMatches[1]
				}
				break
			}
		}
	}

	// 检查扫描过程中的错误
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %v", err)
	}

	return nil
}

func fillCostTime(orders []JnetConfirmedOrder) error {
	// 定义时间字符串的解析格式
	const layout = "01/02/2006 15:04:05.000000" // 注意Go中月份和日的位置是固定的

	for i, order := range orders {
		// 解析RecvTime
		recvTime, err := time.Parse(layout, order.RecvTime)
		if err != nil {
			return fmt.Errorf("error parsing RecvTime for order %s: %v", order.ClOrderId, err)
		}

		// 解析ReturnTime
		returnTime, err := time.Parse(layout, order.ReturnTime)
		if err != nil {
			return fmt.Errorf("error parsing ReturnTime for order %s: %v", order.ClOrderId, err)
		}

		// 计算差值（以秒为单位）
		duration := returnTime.Sub(recvTime).Seconds()

		// 更新CostTime
		orders[i].CostTime = fmt.Sprintf("%.6f", duration) // 保留六位小数，与原时间字符串格式一致
	}

	return nil
}

func exportCsv(orders []JnetConfirmedOrder, csvFilename string) error {
	file, err := os.Create(csvFilename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"ClientOrderID", "CostMillisecond"}); err != nil {
		return fmt.Errorf("error writing header to CSV file: %v", err)
	}

	for _, order := range orders {
		costTimeSeconds, err := strconv.ParseFloat(order.CostTime, 64)
		if err != nil {
			return fmt.Errorf("error parsing CostTime to float: %v", err)
		}

		// 将秒转换为毫秒并格式化为字符串
		costTimeMilliseconds := fmt.Sprintf("%.3f", costTimeSeconds*1000)

		// 写入一行CSV数据
		if err := writer.Write([]string{order.ClOrderId, costTimeMilliseconds}); err != nil {
			return fmt.Errorf("error writing record to CSV file: %v", err)
		}
	}

	// 确保所有的缓存数据都被写入文件
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("error flushing data to CSV file: %v", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <program> <logFilePath> <outputCsvPath>")
		return
	}
	logFilePath := os.Args[1]
	outputCsvPath := os.Args[2]

	orders, err := getAllJnetConfirmedOrder(logFilePath)
	if err != nil {
		fmt.Printf("Error getting orders: %v\n", err)
		return
	}

	if fillSendTime(orders, logFilePath) != nil {
		fmt.Printf("Error filling send time: %v\n", err)
		return
	}

	if fillCostTime(orders) != nil {
		fmt.Printf("Error filling cost time: %v\n", err)
		return
	}

	if exportCsv(orders, outputCsvPath) != nil {
		fmt.Printf("Error exporting to CSV: %v\n", err)
		return
	}

	fmt.Println("Orders exported successfully to", outputCsvPath)
}
