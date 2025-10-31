# concurrent

**Tiny, practical Go concurrency helpers**

A lightweight Go library providing essential concurrency patterns and utilities for building robust concurrent applications.

## Features

### 🎯 **Worker Pool**
Simple fan-out/fan-in pattern with `context` cancellation support. Process jobs concurrently with a fixed number of workers.

### 🔄 **Pipeline**
Composable channel stages (`Map`, `Filter`, `Batch`) with clean cancellation. Build data processing pipelines that are easy to reason about.

### ⚡ **MapConcurrent**
Bounded parallelism over a slice. Process collections concurrently while maintaining order and respecting cancellation.

### 🌊 **Fan Out/In**
Distribute work across multiple workers and merge results efficiently. Includes round-robin distribution strategies.

### 🚦 **Rate Limiting**
Token bucket rate limiting with burst support. Control the rate of operations to prevent overwhelming downstream systems.

### 🔁 **Retry & Circuit Breaker**
Exponential backoff retry logic with configurable policies. Circuit breaker pattern for handling cascading failures.

## Quick Example

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/logimos/concurrent"
)

func main() {
    ctx := context.Background()
    
    // Process 100 numbers concurrently with 10 workers
    data := make([]int, 100)
    for i := range data {
        data[i] = i
    }
    
    results, err := concurrent.MapConcurrent(ctx, data, 10, func(ctx context.Context, n int) (int, error) {
        return n * n, nil
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(results)
}
```

## Why concurrent?

- **Simple API**: Easy to understand and use
- **Context-aware**: Proper cancellation support throughout
- **Type-safe**: Leverages Go generics for type safety
- **Lightweight**: Minimal dependencies, fast compilation
- **Production-ready**: Well-tested and battle-tested patterns

## Installation

```bash
go get github.com/logimos/concurrent
```

## Documentation

- [Getting Started](getting-started/installation.md) - Installation and setup
- [Features](features/pool.md) - Detailed feature documentation
- [Examples](examples/index.md) - Code examples and patterns
- [API Reference](api/index.md) - Complete API documentation

## License

See [LICENSE](../LICENSE) file for details.

