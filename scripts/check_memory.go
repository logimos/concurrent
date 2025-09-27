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
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <memory_analysis.txt> <threshold_mb>\n", os.Args[0])
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

	scanner := bufio.NewScanner(file)
	var totalMemory float64
	failed := false

	// Regex to match memory usage lines
	memoryPattern := regexp.MustCompile(`(\d+\.?\d*)\s+(MB|KB|B)`)

	fmt.Printf("🔍 Checking memory usage (threshold: %.0f MB)\n", threshold)
	fmt.Println(strings.Repeat("=", 60))

	for scanner.Scan() {
		line := scanner.Text()
		if matches := memoryPattern.FindStringSubmatch(line); matches != nil {
			value, _ := strconv.ParseFloat(matches[1], 64)
			unit := matches[2]

			// Convert to MB
			var mbValue float64
			switch unit {
			case "B":
				mbValue = value / (1024 * 1024)
			case "KB":
				mbValue = value / 1024
			case "MB":
				mbValue = value
			}

			totalMemory += mbValue

			if mbValue > threshold {
				fmt.Printf("❌ HIGH: %s (%.2f MB)\n", line, mbValue)
				failed = true
			} else {
				fmt.Printf("✅ OK: %s (%.2f MB)\n", line, mbValue)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📊 Total memory usage: %.2f MB\n", totalMemory)

	if totalMemory > threshold {
		fmt.Printf("❌ Total memory (%.2f MB) exceeds threshold (%.0f MB)\n", totalMemory, threshold)
		failed = true
	} else {
		fmt.Printf("✅ Total memory (%.2f MB) within threshold (%.0f MB)\n", totalMemory, threshold)
	}

	if failed {
		fmt.Println("\n❌ Memory thresholds not met!")
		os.Exit(1)
	} else {
		fmt.Println("\n✅ All memory thresholds met!")
	}
}
