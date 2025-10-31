# Installation

## Requirements

- Go 1.23 or later

## Install

Install the package using `go get`:

```bash
go get github.com/logimos/concurrent
```

## Import

Import the package in your Go code:

```go
import "github.com/logimos/concurrent"
```

## Verify Installation

Create a simple test file to verify the installation:

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/logimos/concurrent"
)

func main() {
    ctx := context.Background()
    data := []int{1, 2, 3}
    
    results, err := concurrent.MapConcurrent(ctx, data, 2, func(ctx context.Context, n int) (int, error) {
        return n * 2, nil
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println(results) // [2 4 6]
}
```

Run it:

```bash
go run main.go
```

If you see `[2 4 6]` printed, the installation is successful!

