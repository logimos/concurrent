# Examples

This section provides practical examples demonstrating how to use the `concurrent` package in real-world scenarios.

## Available Examples

- [Worker Pool](pool.md) - Basic worker pool usage
- [Pipeline](pipeline.md) - Building data processing pipelines
- [MapReduce](mapreduce.md) - Concurrent map operations

## Running Examples

All examples can be found in the `examples/` directory. To run an example:

```bash
# Worker Pool example
cd examples/pool
go run main.go

# Pipeline example
cd examples/pipeline
go run main.go

# MapReduce example
cd examples/mapreduce
go run main.go
```

## Common Patterns

### Error Handling

When using concurrent operations, always handle errors appropriately:

```go
results, err := concurrent.MapConcurrent(ctx, data, 10, func(ctx context.Context, item Item) (Result, error) {
    result, err := process(item)
    if err != nil {
        return Result{}, err
    }
    return result, nil
})

if err != nil {
    log.Fatalf("Processing failed: %v", err)
}
```

### Context Cancellation

Always use contexts with timeouts or cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

results := pool.Run(ctx, jobs)
```

### Resource Cleanup

Ensure proper cleanup of resources:

```go
pipeline := concurrent.NewPipeline[int](ctx)
defer pipeline.Close()

output := pipeline.Run(input)
```

