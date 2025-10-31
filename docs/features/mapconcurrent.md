# MapConcurrent

`MapConcurrent` provides bounded parallelism over a slice, processing elements concurrently while maintaining order and respecting cancellation.

## Overview

`MapConcurrent` applies a function to each element of a slice with at most `n` concurrent operations. Results are returned in the original order, and the first error aborts early.

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/logimos/concurrent"
)

func main() {
    ctx := context.Background()
    
    data := []int{1, 2, 3, 4, 5}
    
    results, err := concurrent.MapConcurrent(ctx, data, 3, func(ctx context.Context, n int) (int, error) {
        return n * n, nil
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(results) // [1 4 9 16 25]
}
```

## Key Features

- **Order Preservation**: Results are returned in the same order as input
- **Bounded Concurrency**: Limits the number of concurrent operations
- **Early Abort**: First error stops processing
- **Cancellation Support**: Respects context cancellation
- **Graceful Shutdown**: Waits for in-flight operations to complete

## API Reference

### `MapConcurrent[T, R](ctx context.Context, in []T, n int, fn func(context.Context, T) (R, error)) ([]R, error)`

Applies `fn` to each element of `in` with at most `n` concurrent tasks.

**Parameters:**
- `ctx`: Context for cancellation
- `in`: Input slice
- `n`: Maximum number of concurrent operations (must be > 0, defaults to 1)
- `fn`: Function to apply to each element

**Returns:**
- `[]R`: Results in the same order as input
- `error`: First error encountered, or nil

## Behavior

### Concurrency Control

The `n` parameter controls the maximum number of goroutines processing items simultaneously. If `n` is greater than the slice length, only as many goroutines as needed are started.

### Error Handling

If any function call returns an error:
- Processing stops immediately
- In-flight operations complete
- The error is returned

### Cancellation

When the context is canceled:
- No new operations start
- In-flight operations complete
- `ctx.Err()` is returned

### Empty Input

If the input slice is empty, an empty result slice and `nil` error are returned immediately.

## Examples

### Processing URLs

```go
type Result struct {
    URL    string
    Status int
}

urls := []string{"https://example.com", "https://google.com", "https://github.com"}

results, err := concurrent.MapConcurrent(ctx, urls, 5, func(ctx context.Context, url string) (Result, error) {
    resp, err := http.Get(url)
    if err != nil {
        return Result{}, err
    }
    defer resp.Body.Close()
    return Result{URL: url, Status: resp.StatusCode}, nil
})

if err != nil {
    log.Fatal(err)
}

for _, r := range results {
    fmt.Printf("%s: %d\n", r.URL, r.Status)
}
```

### With Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

results, err := concurrent.MapConcurrent(ctx, data, 10, processFunc)
if err != nil {
    if err == context.DeadlineExceeded {
        fmt.Println("Operation timed out")
    } else {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Error Propagation

```go
results, err := concurrent.MapConcurrent(ctx, data, 3, func(ctx context.Context, n int) (int, error) {
    if n < 0 {
        return 0, fmt.Errorf("negative number: %d", n)
    }
    return n * n, nil
})

if err != nil {
    // Handle error - processing stopped at first error
    fmt.Printf("Failed: %v\n", err)
    return
}
```

## Best Practices

1. **Choose appropriate concurrency**: `n` should balance throughput and resource usage
2. **Handle errors**: Always check the returned error
3. **Use context**: Pass contexts with timeouts for long-running operations
4. **Idempotent functions**: Ensure `fn` is safe to retry if needed
5. **Resource cleanup**: Make sure `fn` properly cleans up resources (connections, files, etc.)

## Comparison with Pool

- **MapConcurrent**: Best for processing a known slice of items, preserves order, early abort on error
- **Pool**: Best for processing an unknown stream of jobs from a channel, continues on errors

Choose `MapConcurrent` when you have a fixed set of items to process. Choose `Pool` when you have a stream of jobs.

