package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/ballspins/gonik"
)

func main() {
	if err := gonik.InitDatabase(); err != nil {
		panic(err)
	}

	iterations := 1000000
	fmt.Printf("Running benchmark for %d iterations...\n", iterations)

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)

	start := time.Now()
	for i := 0; i < iterations; i++ {
		parser := gonik.New("3578201503990001")
		_ = parser.District()
		_ = parser.BirthDate()
	}
	elapsed := time.Since(start)

	runtime.ReadMemStats(&after)
	peakMemMB := float64(after.HeapAlloc) / 1024 / 1024
	allocDiffMB := float64(after.TotalAlloc-before.TotalAlloc) / 1024 / 1024

	fmt.Println("------------------------------------------")
	fmt.Printf("Go Execution Time: %.6f seconds\n", elapsed.Seconds())
	fmt.Printf("Go Allocated Memory: %.4f MB\n", allocDiffMB)
	fmt.Printf("Go Peak Heap Alloc: %.4f MB\n", peakMemMB)
	fmt.Println("------------------------------------------")
}
