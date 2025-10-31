# Worker Pool

The `Pool` type provides a simple fan-out/fan-in pattern with a fixed number of workers. It's ideal for processing jobs concurrently while maintaining control over resource usage.

## Overview

A worker pool distributes jobs from an input channel to a fixed number of workers, each processing jobs concurrently. Results are collected into a single output channel.

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/logimos/concurrent"
)

func main() {
    ctx := context.Background()
    
    jobs := make(chan int)
    pool := concurrent.NewPool(3, func(ctx context.Context, v int) (string, error) {
        // Simulate work
        time.Sleep(100 * time.Millisecond)
        return fmt.Sprintf("processed-%d", v), nil
    })
    
    results := pool.Run(ctx, jobs)
    
    // Send jobs
    go func() {
        for i := 0; i < 10; i++ {
            jobs <- i
        }
        close(jobs)
    }()
    
    // Collect results
    for r := range results {
        fmt.Println(r)
    }
}
```

## API Reference

### `NewPool[T, R](n int, fn func(context.Context, T) (R, error)) *Pool[T, R]`

Creates a new pool with `n` workers and a processing function.

**Parameters:**
- `n`: Number of workers (must be > 0, defaults to 1 if <= 0)
- `fn`: Processing function that takes a context and input value, returns result and error

**Returns:** A new `Pool` instance

### `Run(ctx context.Context, jobs <-chan T) <-chan R`

Executes jobs until the context is canceled or the jobs channel is closed.

**Important:** The caller MUST consume the results channel until it is closed to prevent goroutine leaks.

**Parameters:**
- `ctx`: Context for cancellation
- `jobs`: Input channel of jobs to process

**Returns:** Output channel of results

## Behavior

- **Error Handling**: If `fn` returns an error, that job's result is dropped. Use a wrapper function if you need to propagate per-item errors.
- **Cancellation**: The pool respects context cancellation. When `ctx` is canceled, workers stop accepting new jobs and complete in-flight operations.
- **Channel Closing**: When the input channel is closed, workers finish processing remaining jobs and the results channel is closed automatically.

## Best Practices

1. **Always consume results**: Make sure to read from the results channel until it's closed
2. **Close input channel**: Close the jobs channel when done sending to signal completion
3. **Use context**: Always pass a context with appropriate timeout or cancellation
4. **Error handling**: Wrap the processing function if you need per-item error handling

## Example: Error Handling

```go
type Result struct {
    Value string
    Error error
}

jobs := make(chan int)
pool := concurrent.NewPool(3, func(ctx context.Context, v int) (Result, error) {
    result, err := processJob(v)
    return Result{Value: result, Error: err}, nil
})

results := pool.Run(ctx, jobs)

for r := range results {
    if r.Error != nil {
        fmt.Printf("Error processing job: %v\n", r.Error)
    } else {
        fmt.Printf("Success: %s\n", r.Value)
    }
}
```

