# GitHub Actions CI/CD Pipeline

This directory contains the CI/CD pipeline configuration for the concurrent library.

Documwntation can be found [here](https://logimos.github.io/concurrent/)

## Workflow Overview

The CI pipeline includes the following jobs:

### 1. **Test Suite** (`test`)
- Runs on multiple Go versions (1.23, 1.24.3)
- Executes all tests with race detection
- Generates coverage reports
- Uploads coverage to Codecov

### 2. **Performance Benchmarks** (`performance`)
- Runs comprehensive performance benchmarks
- Parses and analyzes benchmark results
- Checks performance thresholds
- Uploads benchmark results as artifacts

### 3. **Memory Benchmarks** (`memory`)
- Runs memory profiling benchmarks
- Analyzes memory usage patterns
- Checks memory thresholds
- Uploads memory analysis results

### 4. **Security Scan** (`security`)
- Runs gosec security scanner
- Checks for vulnerabilities with govulncheck
- Ensures code security standards

### 5. **Build and Examples** (`build`)
- Builds the project for multiple platforms
- Tests all example programs
- Ensures cross-platform compatibility

### 6. **Performance Regression Detection** (`regression`)
- Compares current performance with previous runs
- Detects performance regressions
- Generates performance reports
- Only runs on push/schedule events

### 7. **Quality Gate** (`quality-gate`)
- Final validation of all pipeline stages
- Ensures all quality metrics are met
- Provides comprehensive status report

## Configuration

### Environment Variables
- `GO_VERSION`: Go version to use (default: 1.24.3)
- `BENCHMARK_THRESHOLD_NS`: Performance threshold in nanoseconds (default: 100000)
- `MEMORY_THRESHOLD_MB`: Memory threshold in MB (default: 50)
- `COVERAGE_THRESHOLD`: Coverage threshold percentage (default: 80)

### Triggers
- **Push**: Runs on pushes to main/develop branches
- **Pull Request**: Runs on PRs to main/develop branches
- **Schedule**: Daily at 2 AM UTC for regression detection

## Scripts

The pipeline uses several custom scripts in the `scripts/` directory:

- `parse_benchmarks.go`: Parses benchmark output into JSON
- `check_thresholds.go`: Validates performance thresholds
- `check_memory.go`: Validates memory usage thresholds
- `compare_benchmarks.go`: Compares performance between runs
- `generate_report.go`: Generates performance reports
- `test_ci.sh`: Local CI testing script

## Local Testing

To test the CI pipeline locally:

```bash
# Run the full CI pipeline locally
./scripts/test_ci.sh

# Or run individual components
make test-all
make coverage
go test -bench=. -benchmem
```

## Artifacts

The pipeline generates several artifacts:

- **benchmark-results**: Raw benchmark data and summaries
- **memory-results**: Memory profiling data and analysis
- **performance-report**: Generated performance reports

## Quality Gates

The pipeline enforces the following quality gates:

1. ✅ All tests must pass
2. ✅ No race conditions detected
3. ✅ Coverage above threshold (80%)
4. ✅ Performance within thresholds
5. ✅ Memory usage within limits
6. ✅ No security vulnerabilities
7. ✅ All examples build and run
8. ✅ No performance regressions

## Monitoring

The pipeline provides detailed monitoring:

- Real-time test results
- Performance trend analysis
- Memory usage tracking
- Security vulnerability reports
- Coverage trend analysis

## Troubleshooting

### Common Issues

1. **Performance Thresholds**: If benchmarks fail, check if the threshold is too strict
2. **Memory Limits**: If memory checks fail, analyze the memory profile
3. **Race Conditions**: Use `-race` flag to detect concurrency issues
4. **Coverage**: Ensure all new code is properly tested

### Debugging

```bash
# Run specific tests
go test -v -run TestSpecific

# Run benchmarks with detailed output
go test -bench=. -benchmem -v

# Analyze memory usage
go tool pprof mem.prof

# Check race conditions
go test -race -v
```

## Performance Baselines

Current performance baselines (as of latest run):

- **Pool**: ~50,000 ns/op
- **MapConcurrent**: ~49,000 ns/op
- **FanOut**: ~43,000 ns/op
- **FanIn**: ~37,000 ns/op
- **Pipeline**: ~90,000 ns/op
- **RateLimiter**: ~8 ns/op
- **CircuitBreaker**: ~93 ns/op

These baselines are used for regression detection and threshold validation.
