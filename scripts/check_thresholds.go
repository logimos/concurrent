//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type BenchmarkSummary struct {
	Results []BenchmarkResult `json:"results"`
	Summary struct {
		TotalBenchmarks int     `json:"total_benchmarks"`
		AvgNsPerOp      float64 `json:"avg_ns_per_op"`
		MaxNsPerOp      float64 `json:"max_ns_per_op"`
		MinNsPerOp      float64 `json:"min_ns_per_op"`
	} `json:"summary"`
}

type BenchmarkResult struct {
	Name    string  `json:"name"`
	NsPerOp float64 `json:"ns_per_op"`
	Allocs  int     `json:"allocs_per_op"`
	Bytes   int     `json:"bytes_per_op"`
	Runs    int     `json:"runs"`
	AvgTime float64 `json:"avg_time_ns"`
	MinTime float64 `json:"min_time_ns"`
	MaxTime float64 `json:"max_time_ns"`
	StdDev  float64 `json:"std_dev_ns"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <benchmark_summary.json> <threshold_ns>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	thresholdStr := os.Args[2]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing threshold: %v\n", err)
		os.Exit(1)
	}

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var summary BenchmarkSummary
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&summary); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("🔍 Checking performance thresholds (threshold: %.0f ns/op)\n", threshold)
	fmt.Println(strings.Repeat("=", 60))

	failed := false
	passed := 0
	total := len(summary.Results)

	for _, result := range summary.Results {
		status := "✅ PASS"
		if result.NsPerOp > threshold {
			status = "❌ FAIL"
			failed = true
		} else {
			passed++
		}

		fmt.Printf("%s %-30s %8.0f ns/op %3d allocs %4d B/op\n",
			status, result.Name, result.NsPerOp, result.Allocs, result.Bytes)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📊 Summary: %d/%d benchmarks passed (%.1f%%)\n",
		passed, total, float64(passed)/float64(total)*100)

	if summary.Summary.AvgNsPerOp > threshold {
		fmt.Printf("❌ Average performance (%.0f ns/op) exceeds threshold (%.0f ns/op)\n",
			summary.Summary.AvgNsPerOp, threshold)
		failed = true
	} else {
		fmt.Printf("✅ Average performance (%.0f ns/op) within threshold (%.0f ns/op)\n",
			summary.Summary.AvgNsPerOp, threshold)
	}

	if summary.Summary.MaxNsPerOp > threshold {
		fmt.Printf("❌ Worst performance (%.0f ns/op) exceeds threshold (%.0f ns/op)\n",
			summary.Summary.MaxNsPerOp, threshold)
		failed = true
	} else {
		fmt.Printf("✅ Worst performance (%.0f ns/op) within threshold (%.0f ns/op)\n",
			summary.Summary.MaxNsPerOp, threshold)
	}

	if failed {
		fmt.Println("\n❌ Performance thresholds not met!")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ All performance thresholds met!")
	}
}
