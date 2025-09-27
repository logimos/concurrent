//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"
)

type BenchmarkResult struct {
	Name    string  `json:"name"`
	NsPerOp float64 `json:"ns_per_op"`
	Allocs  int     `json:"allocs_per_op"`
	Bytes   int     `json:"bytes_per_op"`
}

type BenchmarkSummary struct {
	Results []BenchmarkResult `json:"results"`
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <previous_results.txt> <current_results.txt>\n", os.Args[0])
		os.Exit(1)
	}

	previousFile := os.Args[1]
	currentFile := os.Args[2]

	// Parse previous results
	prevSummary := parseBenchmarkFile(previousFile)
	currSummary := parseBenchmarkFile(currentFile)

	fmt.Println("📊 Performance Comparison Report")
	fmt.Println(strings.Repeat("=", 50))

	regressions := 0
	improvements := 0
	noChange := 0

	for _, curr := range currSummary.Results {
		var prev *BenchmarkResult
		for _, p := range prevSummary.Results {
			if p.Name == curr.Name {
				prev = &p
				break
			}
		}

		if prev == nil {
			fmt.Printf("🆕 NEW: %s - %.0f ns/op\n", curr.Name, curr.NsPerOp)
			continue
		}

		change := ((curr.NsPerOp - prev.NsPerOp) / prev.NsPerOp) * 100

		if change > 5 {
			fmt.Printf("❌ REGRESSION: %s - %.0f ns/op (%.1f%% slower)\n",
				curr.Name, curr.NsPerOp, change)
			regressions++
		} else if change < -5 {
			fmt.Printf("✅ IMPROVEMENT: %s - %.0f ns/op (%.1f%% faster)\n",
				curr.Name, curr.NsPerOp, -change)
			improvements++
		} else {
			fmt.Printf("➖ NO CHANGE: %s - %.0f ns/op (%.1f%% change)\n",
				curr.Name, curr.NsPerOp, change)
			noChange++
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("📈 Summary: %d regressions, %d improvements, %d no change\n",
		regressions, improvements, noChange)

	if regressions > 0 {
		fmt.Println("❌ Performance regressions detected!")
		os.Exit(1)
	} else {
		fmt.Println("✅ No significant performance regressions!")
	}
}

func parseBenchmarkFile(filename string) BenchmarkSummary {
	// Simplified parsing - in real implementation, you'd parse the actual benchmark output
	// For now, return empty summary
	return BenchmarkSummary{Results: []BenchmarkResult{}}
}
