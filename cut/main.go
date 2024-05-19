package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	startTimeStr := flag.String("start", "", "Start time in format MM/DD/YYYY HH:MM:SS")
	endTimeStr := flag.String("end", "", "End time in format MM/DD/YYYY HH:MM:SS")
	logFilePath := flag.String("file", "", "Path to the log file")
	flag.Parse()

	if *startTimeStr == "" || *endTimeStr == "" || *logFilePath == "" {
		fmt.Println("Usage: go run main.go -start=\"MM/DD/YYYY HH:MM:SS\" -end=\"MM/DD/YYYY HH:MM:SS\" -file=\"path/to/logfile.log\"")
		return
	}

	startTime := *startTimeStr
	endTime := *endTimeStr
	file, err := os.Open(*logFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var capture bool

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, startTime) {
			capture = true
		}
		if capture {
			fmt.Println(line)
		}
		if strings.Contains(line, endTime) {
			capture = false
			break
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
