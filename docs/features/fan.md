# Fan Out/In

Fan-out/fan-in patterns distribute work across multiple workers and merge results efficiently. The `concurrent` package provides several variants for different distribution strategies.

## Overview

- **FanOut**: Distributes work from a single input channel to multiple workers
- **FanIn**: Merges multiple input channels into a single output channel
- **FanOutFanIn**: Combines both patterns for parallel processing
- **RoundRobin**: Distributes work in round-robin fashion

## FanOut

Distributes work from a single input channel to multiple workers, each processing items concurrently.

### Example

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/logimos/concurrent"
)

func main() {
    ctx := context.Background()
    
    input := make(chan int)
    
    // Process with 5 workers
    output := concurrent.FanOut(ctx, input, 5, func(ctx context.Context, n int) (int, error) {
        return n * n, nil
    })
    
    // Send work
    go func() {
        for i := 1; i <= 10; i++ {
            input <- i
        }
        close(input)
    }()
    
    // Collect results
    for result := range output {
        fmt.Println(result)
    }
}
```

### API

```go
func FanOut[T any, R any](ctx context.Context, input <-chan T, workers int, fn func(context.Context, T) (R, error)) <-chan R
```

**Parameters:**
- `ctx`: Context for cancellation
- `input`: Input channel of work items
- `workers`: Number of worker goroutines
- `fn`: Processing function

**Returns:** Output channel of results

## FanIn

Merges multiple input channels into a single output channel.

### Example

```go
ch1 := make(chan int)
ch2 := make(chan int)
ch3 := make(chan int)

output := concurrent.FanIn(ctx, ch1, ch2, ch3)

// Send to channels
go func() { ch1 <- 1; close(ch1) }()
go func() { ch2 <- 2; close(ch2) }()
go func() { ch3 <- 3; close(ch3) }()

// Collect merged results
for result := range output {
    fmt.Println(result) // Order may vary
}
```

### API

```go
func FanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T
```

**Parameters:**
- `ctx`: Context for cancellation
- `inputs`: Variable number of input channels

**Returns:** Merged output channel

## FanOutFanIn

Combines fan-out and fan-in for parallel processing with result merging.

### Example

```go
input := make(chan string)

// Process with 3 workers and merge results
output := concurrent.FanOutFanIn(ctx, input, 3, func(ctx context.Context, s string) (string, error) {
    return strings.ToUpper(s), nil
})

go func() {
    input <- "hello"
    input <- "world"
    close(input)
}()

for result := range output {
    fmt.Println(result)
}
```

### API

```go
func FanOutFanIn[T any, R any](ctx context.Context, input <-chan T, workers int, fn func(context.Context, T) (R, error)) <-chan R
```

## RoundRobin

Distributes work in round-robin fashion to multiple workers.

### Example

```go
input := make(chan int)

// Distribute evenly across 4 workers
output := concurrent.RoundRobin(ctx, input, 4, func(ctx context.Context, n int) (int, error) {
    return n * 2, nil
})

go func() {
    for i := 1; i <= 8; i++ {
        input <- i
    }
    close(input)
}()

for result := range output {
    fmt.Println(result)
}
```

### API

```go
func RoundRobin[T any, R any](ctx context.Context, input <-chan T, workers int, fn func(context.Context, T) (R, error)) <-chan R
```

**Note:** RoundRobin ensures work is distributed evenly across workers, unlike FanOut which distributes work as workers become available.

## Use Cases

### FanOut
- When you want concurrent processing but don't care about work distribution order
- Processing independent tasks from a single source

### FanIn
- Merging results from multiple processing pipelines
- Combining outputs from different data sources

### FanOutFanIn
- Parallel processing with automatic result merging
- Simplifies the pattern when you need both fan-out and fan-in

### RoundRobin
- When you need even work distribution
- Load balancing across workers

## Best Practices

1. **Close input channels**: Always close input channels when done sending
2. **Consume output**: Read from output channels until closed
3. **Handle errors**: Errors in processing functions are silently dropped - wrap if needed
4. **Use context**: Pass contexts with appropriate timeouts
5. **Worker count**: Choose worker count based on CPU cores and I/O characteristics

## Error Handling

By default, errors from processing functions are dropped. To handle errors:

```go
type Result struct {
    Value int
    Error error
}

output := concurrent.FanOut(ctx, input, 5, func(ctx context.Context, n int) (Result, error) {
    result, err := process(n)
    return Result{Value: result, Error: err}, nil
})

for r := range output {
    if r.Error != nil {
        fmt.Printf("Error: %v\n", r.Error)
    } else {
        fmt.Printf("Success: %d\n", r.Value)
    }
}
```

