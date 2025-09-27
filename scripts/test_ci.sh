#!/bin/bash

# Test script to simulate GitHub Actions CI pipeline locally
set -e

echo "🚀 Starting CI Pipeline Test"
echo "=============================="

# Test 1: Dependencies
echo "📦 Installing dependencies..."
make deps

# Test 2: Linting
echo "🔍 Running linter..."
make lint

# Test 3: Tests
echo "🧪 Running tests..."
make test

# Test 4: Race detection
echo "🏃 Running race detection tests..."
make test-race

# Test 5: Coverage
echo "📊 Running coverage tests..."
make coverage

# Test 6: Performance benchmarks
echo "⚡ Running performance benchmarks..."
go test -bench=. -benchmem -count=3 > benchmark_results.txt

# Test 7: Parse benchmarks
echo "📈 Parsing benchmark results..."
go run scripts/parse_benchmarks.go benchmark_results.txt > benchmark_summary.json

# Test 8: Check performance thresholds
echo "🎯 Checking performance thresholds..."
go run scripts/check_thresholds.go benchmark_summary.json 100000

# Test 9: Memory analysis
echo "🧠 Running memory analysis..."
go test -bench=. -benchmem -memprofile=mem.prof -count=3 > /dev/null 2>&1
go tool pprof -text -alloc_space mem.prof > memory_analysis.txt

# Test 10: Check memory thresholds
echo "💾 Checking memory thresholds..."
go run scripts/check_memory.go memory_analysis.txt 50

# Test 11: Build examples
echo "🔨 Building examples..."
make examples

# Test 12: Generate report
echo "📋 Generating performance report..."
go run scripts/generate_report.go benchmark_results.txt > performance_report.md

echo ""
echo "✅ All CI tests passed!"
echo "📊 Performance report generated: performance_report.md"
echo "📈 Benchmark summary: benchmark_summary.json"
echo "🧠 Memory analysis: memory_analysis.txt"
