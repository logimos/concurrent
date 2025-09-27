//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BenchmarkData struct {
	Name    string
	NsPerOp float64
	Allocs  int
	Bytes   int
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <benchmark_results.txt>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]
	benchmarks := parseBenchmarkFile(filename)

	report := generateReport(benchmarks)
	fmt.Print(report)
}

func parseBenchmarkFile(filename string) []BenchmarkData {
	file, err := os.Open(filename)
	if err != nil {
		return []BenchmarkData{}
	}
	defer file.Close()

	var benchmarks []BenchmarkData
	scanner := bufio.NewScanner(file)
	pattern := regexp.MustCompile(`^Benchmark(\w+)-(\d+)\s+(\d+)\s+([\d.]+)\s+ns/op\s+([\d.]+)\s+allocs/op\s+([\d.]+)\s+B/op`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := pattern.FindStringSubmatch(line); matches != nil {
			name := matches[1]
			nsPerOp, _ := strconv.ParseFloat(matches[4], 64)
			allocs, _ := strconv.Atoi(matches[5])
			bytes, _ := strconv.Atoi(matches[6])

			benchmarks = append(benchmarks, BenchmarkData{
				Name:    name,
				NsPerOp: nsPerOp,
				Allocs:  allocs,
				Bytes:   bytes,
			})
		}
	}

	return benchmarks
}

func generateReport(benchmarks []BenchmarkData) string {
	var report strings.Builder

	report.WriteString("# Performance Report\n\n")
	report.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	report.WriteString("## Benchmark Results\n\n")
	report.WriteString("| Benchmark | ns/op | allocs/op | B/op |\n")
	report.WriteString("|-----------|-------|-----------|------|\n")

	for _, b := range benchmarks {
		report.WriteString(fmt.Sprintf("| %s | %.0f | %d | %d |\n",
			b.Name, b.NsPerOp, b.Allocs, b.Bytes))
	}

	report.WriteString("\n## Summary\n\n")

	if len(benchmarks) > 0 {
		var totalNs, maxNs, minNs float64
		minNs = benchmarks[0].NsPerOp
		maxNs = benchmarks[0].NsPerOp

		for _, b := range benchmarks {
			totalNs += b.NsPerOp
			if b.NsPerOp > maxNs {
				maxNs = b.NsPerOp
			}
			if b.NsPerOp < minNs {
				minNs = b.NsPerOp
			}
		}

		avgNs := totalNs / float64(len(benchmarks))

		report.WriteString(fmt.Sprintf("- **Total Benchmarks**: %d\n", len(benchmarks)))
		report.WriteString(fmt.Sprintf("- **Average Performance**: %.0f ns/op\n", avgNs))
		report.WriteString(fmt.Sprintf("- **Best Performance**: %.0f ns/op\n", minNs))
		report.WriteString(fmt.Sprintf("- **Worst Performance**: %.0f ns/op\n", maxNs))
	}

	return report.String()
}
