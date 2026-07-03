package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"time"
)

func main() {
	iterations := 50
	fmt.Printf("Running Benchmark sampling %dx\n", iterations)

	startTime := time.Now()

	cmd := exec.Command("go", "test", "-bench=.", "-benchmem", "-count="+strconv.Itoa(iterations), "./...")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed running go test: %v\n", err)
		fmt.Printf("Error:\n%s\n", stderr.String())
		return
	}

	totalDuration := time.Since(startTime)

	outputStr := out.String()

	targets := []string{
		"BenchmarkGenerateNIK_BatchLoop",
		"BenchmarkGenerateRandomNIK_BatchLoop",
		"BenchmarkGenerateNIK_SyncPool",
		"BenchmarkGenerateRandomNIK_SyncPool",
		"BenchmarkNikParser_GetDetails",
		"BenchmarkParser_Province",
		"BenchmarkParser_RegencyCity",
		"BenchmarkParser_District",
		"BenchmarkParser_PostalCode",
		"BenchmarkParser_Gender",
		"BenchmarkParser_BirthDate",
		"BenchmarkParser_getSubstring",
	}

	fmt.Println("\n================ AVG SAMPLING RESULT ================")
	
	for _, target := range targets {
		re := regexp.MustCompile(target + `-\d+\s+\d+\s+([\d.]+)\s+ns/op`)
		matches := re.FindAllStringSubmatch(outputStr, -1)

		if len(matches) == 0 {
			fmt.Printf("%-40s : Data not found\n", target)
			continue
		}

		var totalNs float64
		count := 0
		for _, match := range matches {
			if len(match) > 1 {
				val, err := strconv.ParseFloat(match[1], 64)
				if err == nil {
					totalNs += val
					count++
				}
			}
		}

		avg := totalNs / float64(count)
		fmt.Printf("%-40s : %.2f ns/op (%d samples)\n", target, avg, count)
	}
	fmt.Println("==========================================================")
	fmt.Printf("Total Sampling Time Execution: %s\n", totalDuration.Truncate(time.Millisecond))
	fmt.Println("==========================================================")
}