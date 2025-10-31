# Quick Start

This guide will help you get started with `concurrent` in minutes.

## Worker Pool

Process jobs concurrently with a fixed number of workers:

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

## Pipeline

Build composable data processing pipelines:

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

## MapConcurrent

Process a slice concurrently with bounded parallelism:

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

## Next Steps

- Read the [Feature Documentation](../features/pool.md) for detailed usage
- Check out [Examples](../examples/index.md) for more patterns
- Review the [API Reference](../api/index.md) for complete API details

