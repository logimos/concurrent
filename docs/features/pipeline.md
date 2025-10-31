# Pipeline

The `Pipeline` type provides composable channel stages for building data processing pipelines. It supports operations like `Map`, `Filter`, and `Batch` with proper cancellation handling.

## Overview

A pipeline chains together multiple stages, where each stage transforms data flowing through channels. Stages are composable and can be combined to build complex data processing workflows.

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
    
    input := make(chan int)
    pipeline := concurrent.NewPipeline[int](ctx)
    
    // Add stages
    pipeline.AddStage(concurrent.Map(func(n int) int {
        return n * 2
    }))
    
    pipeline.AddStage(concurrent.Filter(func(n int) bool {
        return n > 10
    }))
    
    output := pipeline.Run(input)
    
    // Send data
    go func() {
        for i := 1; i <= 10; i++ {
            input <- i
        }
        close(input)
    }()
    
    // Process results
    for result := range output {
        fmt.Println(result) // 12, 14, 16, 18, 20
    }
    
    pipeline.Close()
}
```

## Pipeline Builder

You can also use the fluent builder pattern:

```go
pipeline := concurrent.NewPipelineBuilder[int](ctx).
    AddStage(concurrent.Map(func(n int) int { return n * 2 })).
    AddStage(concurrent.Filter(func(n int) bool { return n > 10 })).
    Build()

output := pipeline.Run(input)
```

## Available Stages

### Map

Applies a function to each item:

```go
pipeline.AddStage(concurrent.Map(func(n int) int {
    return n * 2
}))
```

### Filter

Keeps only items where the predicate returns true:

```go
pipeline.AddStage(concurrent.Filter(func(n int) bool {
    return n > 0
}))
```

### Batch

Groups items into slices of a specified size:

```go
pipeline.AddStage(concurrent.Batch[int](10))
```

**Note:** The output type changes from `T` to `[]T` when using `Batch`.

### Unbatch

Splits batches back into individual items:

```go
pipeline.AddStage(concurrent.Unbatch[int]())
```

**Note:** The input type must be `[]T` and output type becomes `T`.

### Tee

Splits the input to multiple output channels:

```go
output1 := make(chan int)
output2 := make(chan int)

pipeline.AddStage(concurrent.Tee(output1, output2))
```

**Warning:** Tee closes the provided output channels when done. Do not reuse these channels.

### Merge

Merges multiple input channels into one output:

```go
output := concurrent.Merge(ctx, input1, input2, input3)
```

## Advanced Examples

### Batching Pipeline

```go
ctx := context.Background()
input := make(chan int)

pipeline := concurrent.NewPipeline[int](ctx)
pipeline.AddStage(concurrent.Map(func(n int) int { return n * 2 }))
pipeline.AddStage(concurrent.Batch[int](5))

output := pipeline.Run(input)

// Send data
go func() {
    for i := 1; i <= 12; i++ {
        input <- i
    }
    close(input)
}()

// Results are batches of 5
for batch := range output {
    fmt.Println(batch) // [2 4 6 8 10], [12 14 16 18 20], [22 24]
}
```

### Complex Pipeline

```go
pipeline := concurrent.NewPipeline[int](ctx).
    AddStage(concurrent.Map(func(n int) int { return n * n })).
    AddStage(concurrent.Filter(func(n int) bool { return n%2 == 0 })).
    AddStage(concurrent.Batch[int](3)).
    Build()
```

## API Reference

### `NewPipeline[T](ctx context.Context) *Pipeline[T]`

Creates a new pipeline with the given context.

### `AddStage(stage Stage[T, T]) *Pipeline[T]`

Adds a stage to the pipeline. Returns the pipeline for method chaining.

### `Run(input <-chan T) <-chan T`

Executes the pipeline with the given input channel. Returns the output channel.

### `Close()`

Cancels the pipeline context, stopping all stages.

### `Stage[T, R]`

A stage is a function that transforms an input channel to an output channel:

```go
type Stage[T any, R any] func(context.Context, <-chan T) <-chan R
```

## Best Practices

1. **Always close the pipeline**: Call `pipeline.Close()` when done to clean up resources
2. **Consume output**: Read from the output channel until it's closed
3. **Close input**: Close the input channel when done sending data
4. **Use context**: Pass a context with appropriate timeout or cancellation
5. **Type consistency**: Ensure stage types match (except for Batch/Unbatch)

## Cancellation

All stages respect context cancellation. When the pipeline context is canceled:
- Stages stop accepting new items
- In-flight operations complete gracefully
- Output channels are closed properly

