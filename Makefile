# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOBENCH=$(GOCMD) test -bench=.
GORACE=$(GOCMD) test -race
GOCOVER=$(GOCMD) test -cover
GOCOVERPROFILE=$(GOCMD) test -coverprofile=coverage.out
GOCOVERHTML=$(GOCMD) tool cover -html=coverage.out

# Project parameters
BINARY_NAME=concurrent
BINARY_UNIX=$(BINARY_NAME)_unix
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

# Default target
.PHONY: all
all: clean deps test build

# Build the project (verify compilation)
.PHONY: build
build:
	$(GOBUILD) -v ./...

# Build for Linux (verify compilation)
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_HTML)

# Run tests
.PHONY: test
test:
	$(GOTEST) -v ./...

# Run tests with race detection
.PHONY: test-race
test-race:
	$(GORACE) -v ./...

# Run benchmarks
.PHONY: bench
bench:
	$(GOBENCH) ./...

# Run tests with coverage
.PHONY: coverage
coverage:
	$(GOCOVERPROFILE)
	$(GOCOVERHTML) -o $(COVERAGE_HTML)
	@echo "Coverage report generated: $(COVERAGE_HTML)"

# Show coverage in terminal
.PHONY: coverage-show
coverage-show:
	$(GOCOVER) ./...

# Run all tests including race detection and coverage
.PHONY: test-all
test-all: test-race coverage

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) ./pipeline.go ./pool.go ./mapreduce.go ./fan.go ./rate.go ./retry.go ./config.go ./concurrent_test.go ./fan_test.go ./pipeline_test.go ./rate_test.go ./retry_test.go

# Run go vet
.PHONY: vet
vet:
	$(GOVET) ./pipeline.go ./pool.go ./mapreduce.go ./fan.go ./rate.go ./retry.go ./config.go

# Lint and format check
.PHONY: lint
lint: fmt vet

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run examples
.PHONY: examples
examples:
	@echo "Running pool example..."
	$(GOCMD) run examples/pool/main.go
	@echo "\nRunning mapreduce example..."
	$(GOCMD) run examples/mapreduce/main.go
	@echo "\nRunning pipeline example..."
	$(GOCMD) run examples/pipeline.go

# Run specific example
.PHONY: example-pool
example-pool:
	$(GOCMD) run examples/pool/main.go

.PHONY: example-mapreduce
example-mapreduce:
	$(GOCMD) run examples/mapreduce/main.go

.PHONY: example-pipeline
example-pipeline:
	$(GOCMD) run examples/pipeline.go

# Development workflow
.PHONY: dev
dev: deps fmt vet test

# CI/CD workflow
.PHONY: ci
ci: deps fmt vet test-race coverage

# Performance testing
.PHONY: perf
perf: bench

# Memory profiling
.PHONY: memprofile
memprofile:
	$(GOTEST) -memprofile=mem.prof -bench=. ./...

# CPU profiling
.PHONY: cpuprofile
cpuprofile:
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...

# Security scan (requires gosec)
.PHONY: security
security:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate mocks (if using mockgen)
.PHONY: mocks
mocks:
	@if command -v mockgen >/dev/null 2>&1; then \
		mockgen -source=pool.go -destination=mocks/pool_mock.go; \
	else \
		echo "mockgen not installed. Install with: go install github.com/golang/mock/mockgen@latest"; \
	fi

# Documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	$(GOCMD) doc -all ./...

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean, install deps, test, and build"
	@echo "  build        - Build the project"
	@echo "  build-linux  - Build for Linux"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests"
	@echo "  test-race    - Run tests with race detection"
	@echo "  test-all     - Run all tests (race + coverage)"
	@echo "  bench        - Run benchmarks"
	@echo "  coverage     - Run tests with coverage and generate HTML report"
	@echo "  coverage-show- Show coverage in terminal"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  lint         - Format and vet"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  examples     - Run all examples"
	@echo "  example-pool - Run pool example"
	@echo "  example-mapreduce - Run mapreduce example"
	@echo "  example-pipeline - Run pipeline example"
	@echo "  dev          - Development workflow (deps, fmt, vet, test)"
	@echo "  ci           - CI/CD workflow (deps, fmt, vet, test-race, coverage)"
	@echo "  perf         - Performance testing"
	@echo "  memprofile   - Generate memory profile"
	@echo "  cpuprofile   - Generate CPU profile"
	@echo "  security     - Run security scan (requires gosec)"
	@echo "  mocks        - Generate mocks (requires mockgen)"
	@echo "  docs         - Generate documentation"
	@echo "  help         - Show this help message"
