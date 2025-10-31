# Pipeline Example

This example demonstrates how to build data processing pipelines using the `Pipeline` type.

## Basic Example

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
    
    // Multiply by 2
    pipeline.AddStage(concurrent.Map(func(n int) int {
        return n * 2
    }))
    
    // Filter even numbers
    pipeline.AddStage(concurrent.Filter(func(n int) bool {
        return n%2 == 0
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
        fmt.Println(result)
    }
    
    pipeline.Close()
}
```

## Example: Data Transformation Pipeline

```go
package main

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/logimos/concurrent"
)

type Data struct {
    Text string
    Len  int
}

func main() {
    ctx := context.Background()
    
    input := make(chan string)
    pipeline := concurrent.NewPipeline[string](ctx)
    
    // Transform: uppercase
    pipeline.AddStage(concurrent.Map(func(s string) string {
        return strings.ToUpper(s)
    }))
    
    // Filter: keep only long strings
    pipeline.AddStage(concurrent.Map(func(s string) Data {
        return Data{Text: s, Len: len(s)}
    }))
    
    pipeline.AddStage(concurrent.Filter(func(d Data) bool {
        return d.Len > 5
    }))
    
    output := pipeline.Run(input)
    
    go func() {
        input <- "hello"
        input <- "world"
        input <- "go"
        input <- "pipeline"
        close(input)
    }()
    
    for result := range output {
        fmt.Printf("%s (%d chars)\n", result.Text, result.Len)
    }
    
    pipeline.Close()
}
```

## Example: Batching Pipeline

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
    
    // Square each number
    pipeline.AddStage(concurrent.Map(func(n int) int {
        return n * n
    }))
    
    // Batch into groups of 5
    pipeline.AddStage(concurrent.Batch[int](5))
    
    output := pipeline.Run(input)
    
    go func() {
        for i := 1; i <= 12; i++ {
            input <- i
        }
        close(input)
    }()
    
    for batch := range output {
        fmt.Println(batch)
    }
    
    pipeline.Close()
}
```

Output:
```
[1 4 9 16 25]
[36 49 64 81 100]
[121 144]
```

## Example: Using Pipeline Builder

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
    
    pipeline := concurrent.NewPipelineBuilder[int](ctx).
        AddStage(concurrent.Map(func(n int) int { return n * 2 })).
        AddStage(concurrent.Filter(func(n int) bool { return n > 10 })).
        AddStage(concurrent.Batch[int](3)).
        Build()
    
    output := pipeline.Run(input)
    
    go func() {
        for i := 1; i <= 10; i++ {
            input <- i
        }
        close(input)
    }()
    
    for batch := range output {
        fmt.Println(batch)
    }
    
    pipeline.Close()
}
```

