package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

var reTime, reClOrderId, reAccount, reSymbol, reOrderType *regexp.Regexp

func init() {
	reTime = regexp.MustCompile(`^D\d{4} (\d{2}/\d{2}/\d{4} \d{2}:\d{2}:\d{2}\.\d{6})`)
	reClOrderId = regexp.MustCompile(`\|11=([^|]+)\|`)
	reAccount = regexp.MustCompile(`\|1=([^|]+)\|`)
	reSymbol = regexp.MustCompile(`\|55=([^|]+)\|`)
	reOrderType = regexp.MustCompile(`\|35=([^|]+)\|`)
}

type Order struct {
	LogTime   string
	OrderType string
	ClOrderId string
	Account   string
	Symbol    string
}

func isOrder(line string) bool {
	return strings.Contains(line, "8=FIX") && strings.Contains(line, "11=") && strings.Contains(line, "55=") && strings.Contains(line, "recv") && strings.Contains(line, "|49=HRT")
}

func parseLine(line string) (Order, error) {
	returnTimeMatches := reTime.FindStringSubmatch(line)
	clOrderIdMatches := reClOrderId.FindStringSubmatch(line)
	accountMatches := reAccount.FindStringSubmatch(line)
	symbolMatches := reSymbol.FindStringSubmatch(line)
	orderTypeMatches := reOrderType.FindStringSubmatch(line)

	order := Order{}

	if len(returnTimeMatches) > 1 {
		order.LogTime = returnTimeMatches[1]
	} else {
		return Order{}, fmt.Errorf("LogTime not found")
	}

	if len(clOrderIdMatches) > 1 {
		order.ClOrderId = clOrderIdMatches[1]
	} else {
		return Order{}, fmt.Errorf("clOrderId not found")
	}

	if len(accountMatches) > 1 {
		order.Account = accountMatches[1]
	} else {
		return Order{}, fmt.Errorf("account not found")
	}

	if len(symbolMatches) > 1 {
		order.Symbol = symbolMatches[1]
	} else {
		return Order{}, fmt.Errorf("symbol not found")
	}

	if len(orderTypeMatches) > 1 {
		order.OrderType = orderTypeMatches[1]
	} else {
		return Order{}, fmt.Errorf("order type not found")
	}

	return order, nil
}

func getOrders(filename string) (map[string]Order, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	orders := make(map[string]Order)
	scanner := bufio.NewScanner(file)

	count := 0
	for scanner.Scan() {
		// 替换每一行中的'\x01'为'|'
		line := strings.Replace(scanner.Text(), "\x01", "|", -1)
		if isOrder(line) {
			count += 1
			order, err := parseLine(line)
			if err != nil {
				fmt.Printf("parse error: %v\n", err)
				continue // 解析错误时跳过该行
			}
			orders[order.ClOrderId] = order
		}
	}

	fmt.Println("Order Count: ", count)

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %v", err)
	}

	return orders, nil
}

func exportToJsonl(orders map[string]Order, jsonlFilename string) error {
	// 创建文件
	file, err := os.Create(jsonlFilename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	// 将map转换为slice以便排序
	ordersSlice := make([]Order, 0, len(orders))
	for _, order := range orders {
		ordersSlice = append(ordersSlice, order)
	}

	// 使用sort.Slice对ordersSlice进行排序
	sort.Slice(ordersSlice, func(i, j int) bool {
		t1, err1 := time.Parse("01/02/2006 15:04:05.000000", ordersSlice[i].LogTime)
		t2, err2 := time.Parse("01/02/2006 15:04:05.000000", ordersSlice[j].LogTime)
		if err1 != nil || err2 != nil {
			fmt.Printf("Error parsing time: %v, %v\n", err1, err2)
			return false
		}
		return t1.Before(t2)
	})

	// 按顺序将排序后的orders写入文件
	for _, order := range ordersSlice {
		// 根据OrderType修改其值
		switch order.OrderType {
		case "D":
			order.OrderType = "New"
		case "F":
			order.OrderType = "Cancel"
		}

		jsonBytes, err := json.Marshal(order)
		if err != nil {
			return fmt.Errorf("error marshalling to JSON: %v", err)
		}
		_, err = file.Write(jsonBytes)
		if err != nil {
			return fmt.Errorf("error writing to file: %v", err)
		}
		_, err = file.WriteString("\n")
		if err != nil {
			return fmt.Errorf("error writing newline to file: %v", err)
		}
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: <program> <logFilePath> <outputJsonlPath> \nVersion: 0.0.1")
		return
	}
	logFilePath := os.Args[1]
	outputJsonlPath := os.Args[2]

	orders, err := getOrders(logFilePath)
	if err != nil {
		fmt.Printf("Error getting orders: %v\n", err)
		return
	}

	if exportToJsonl(orders, outputJsonlPath) != nil {
		fmt.Printf("Error filling send time: %v\n", err)
		return
	}

	fmt.Println("Orders exported successfully to", outputJsonlPath)
}
