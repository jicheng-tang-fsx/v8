package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

var reTime, reClOrderId, reMatchOrderID *regexp.Regexp

func init() {
	reTime = regexp.MustCompile(`^D\d{4} (\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}\.\d{6})`)
	reClOrderId = regexp.MustCompile(`\|11=([^|]+)\|`)
	reMatchOrderID = regexp.MustCompile(`\|198=([^|]+)\|`)
}

type JnetConfirmedOrder struct {
	ClOrderId       string
	RecvClientTime  string
	SendMatchTime   string
	RecvMatchTime   string
	FinalReturnTime string

	OmsCostTime1  string
	MatchCostTime string
	OmsCostTime2  string

	TotalCostTime string
}

func isJnetConfirmed(line string) bool {
	if strings.Contains(line, "|35=8|") && strings.Contains(line, "|20=2|") && strings.Contains(line, "|39=2|") && strings.Contains(line, "8=FIX") {
		return true
	}
	return false
}

func parseLine(line string) (JnetConfirmedOrder, error) {
	returnTimeMatches := reTime.FindStringSubmatch(line)
	clOrderIdMatches := reClOrderId.FindStringSubmatch(line)

	order := JnetConfirmedOrder{}

	if len(returnTimeMatches) > 1 {
		order.FinalReturnTime = returnTimeMatches[1]
	} else {
		return JnetConfirmedOrder{}, fmt.Errorf("FinalReturnTime not found")
	}
	if len(clOrderIdMatches) > 1 {
		order.ClOrderId = clOrderIdMatches[1]
	} else {
		return JnetConfirmedOrder{}, fmt.Errorf("ClOrderId not found")
	}

	return order, nil
}

func getAllJnetConfirmedOrder(filename string) (map[string]JnetConfirmedOrder, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	orders := make(map[string]JnetConfirmedOrder)
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
			orders[order.ClOrderId] = order
		}
	}

	fmt.Println("JNET Correction Order Count: ", count)

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return orders, nil
}

func exportToJsonl(orders map[string]JnetConfirmedOrder, jsonlFilename string) error {
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

func fillSendTime(orders map[string]JnetConfirmedOrder, filename string) error {
	// 打开文件
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Replace(scanner.Text(), "\x01", "|", -1)
		if strings.Contains(line, "|35=D|") && strings.Contains(line, "8=FIX") {
			clOrderIdMatches := reClOrderId.FindStringSubmatch(line)
			recvTimeMatches := reTime.FindStringSubmatch(line)
			if len(clOrderIdMatches) > 1 && len(recvTimeMatches) > 1 {
				order, exists := orders[clOrderIdMatches[1]]
				if exists && order.RecvClientTime == "" {
					// 修改结构体字段
					order.RecvClientTime = recvTimeMatches[1]
					orders[clOrderIdMatches[1]] = order
				}
			}
		}

		if strings.Contains(line, "|35=D|") && strings.Contains(line, "|49=router_branch|") && strings.Contains(line, "|56=exch_sim|") {
			matchOrderIDResults := reMatchOrderID.FindStringSubmatch(line)
			timeMatches := reTime.FindStringSubmatch(line)
			if len(matchOrderIDResults) > 1 && len(timeMatches) > 1 {
				order, exists := orders[matchOrderIDResults[1]]
				if exists && order.SendMatchTime == "" {
					// 修改结构体字段
					order.SendMatchTime = timeMatches[1]
					orders[matchOrderIDResults[1]] = order
				}
			}
		}

		if strings.Contains(line, "|150=G|") && strings.Contains(line, "|49=exch_sim|") && strings.Contains(line, "|56=router_branch|") {
			matchOrderIDResults := reMatchOrderID.FindStringSubmatch(line)
			timeMatches := reTime.FindStringSubmatch(line)
			if len(matchOrderIDResults) > 1 && len(timeMatches) > 1 {
				order, exists := orders[matchOrderIDResults[1]]
				if exists && order.RecvMatchTime == "" {
					// 修改结构体字段
					order.RecvMatchTime = timeMatches[1]
					orders[matchOrderIDResults[1]] = order
				}
			}
		}

	}

	// 检查扫描过程中的错误
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file: %v", err)
	}

	return nil
}

func fillCostTime(orders map[string]JnetConfirmedOrder) error {
	// 定义时间字符串的解析格式
	const layout = "01/02/2006 15:04:05.000000" // 注意Go中月份和日的位置是固定的

	for i, order := range orders {
		// 解析RecvClientTime
		recvClientTime, err := time.Parse(layout, order.RecvClientTime)
		if err != nil {
			fmt.Printf("error parsing RecvClientTime for order %s: %v", order.ClOrderId, err)
			return fmt.Errorf("error parsing RecvClientTime for order %s: %v", order.ClOrderId, err)
		}

		// 解析SendMatchTime
		sendMatchTime, err := time.Parse(layout, order.SendMatchTime)
		if err != nil {
			fmt.Printf("error parsing SendMatchTime for order %s: %v", order.ClOrderId, err)
			return fmt.Errorf("error parsing SendMatchTime for order %s: %v", order.ClOrderId, err)
		}

		// 解析RecvMatchTime
		recvMatchTime, err := time.Parse(layout, order.RecvMatchTime)
		if err != nil {
			fmt.Printf("error parsing RecvMatchTime for order %s: %v", order.ClOrderId, err)
			return fmt.Errorf("error parsing RecvMatchTime for order %s: %v", order.ClOrderId, err)
		}

		// 解析FinalReturnTime
		finalReturnTime, err := time.Parse(layout, order.FinalReturnTime)
		if err != nil {
			fmt.Printf("error parsing FinalReturnTime for order %s: %v", order.ClOrderId, err)
			return fmt.Errorf("error parsing FinalReturnTime for order %s: %v", order.ClOrderId, err)
		}

		// 计算OmsCostTime1
		order.OmsCostTime1 = fmt.Sprintf("%.6f", sendMatchTime.Sub(recvClientTime).Seconds())
		// 计算MatchCostTime
		order.MatchCostTime = fmt.Sprintf("%.6f", recvMatchTime.Sub(sendMatchTime).Seconds())
		// 计算OmsCostTime2
		order.OmsCostTime2 = fmt.Sprintf("%.6f", finalReturnTime.Sub(recvMatchTime).Seconds())
		// 更新TotalCostTime
		order.TotalCostTime = fmt.Sprintf("%.6f", finalReturnTime.Sub(recvClientTime).Seconds())

		// 更新map中的订单
		orders[i] = order
	}

	return nil
}

func exportCsv(orders map[string]JnetConfirmedOrder, csvFilename string) error {
	file, err := os.Create(csvFilename)
	if err != nil {
		return fmt.Errorf("error creating CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write([]string{"ClientOrderID", "OmsCostTime1", "MatchCostTime", "OmsCostTime2", "TotalCostTime"}); err != nil {
		return fmt.Errorf("error writing header to CSV file: %v", err)
	}

	// 创建切片用于排序
	orderSlice := make([]JnetConfirmedOrder, 0, len(orders))
	for _, order := range orders {
		orderSlice = append(orderSlice, order)
	}

	// 根据RecvTime排序订单
	sort.Slice(orderSlice, func(i, j int) bool {
		return orderSlice[i].RecvClientTime < orderSlice[j].RecvClientTime
	})

	// 假设writer是已经被初始化的csv.Writer
	for _, order := range orderSlice {
		// 准备要写入CSV的记录
		record := []string{order.ClOrderId}

		// 需要转换和格式化的字段
		timeFields := []string{order.OmsCostTime1, order.MatchCostTime, order.OmsCostTime2, order.TotalCostTime}

		// 遍历每个时间字段进行处理
		for _, field := range timeFields {
			costTimeSeconds, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return fmt.Errorf("error parsing time field to float: %v", err)
			}
			// 将秒转换为毫秒并格式化为字符串
			costTimeMilliseconds := fmt.Sprintf("%.3f", costTimeSeconds*1000)
			// 将处理后的时间添加到记录中
			record = append(record, costTimeMilliseconds)
		}

		// 写入一行CSV数据
		if err := writer.Write(record); err != nil {
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

	// if exportToJsonl(orders, "t1.jsonl") != nil {
	// 	fmt.Printf("Error exportToJsonl: %v\n", err)
	// 	return
	// }

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
