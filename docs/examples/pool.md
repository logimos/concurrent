# Worker Pool Example

This example demonstrates how to use the `Pool` type to process jobs concurrently.

## Basic Example

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
        time.Sleep(15 * time.Millisecond)
        return fmt.Sprintf("processed-%d", v), nil
    })
    results := pool.Run(ctx, jobs)
    
    go func() {
        for i := 0; i < 8; i++ {
            jobs <- i
        }
        close(jobs)
    }()
    
    for r := range results {
        fmt.Println(r)
    }
}
```

## Output

```
processed-0
processed-1
processed-2
processed-3
processed-4
processed-5
processed-6
processed-7
```

## Advanced Example: HTTP Requests

```go
package main

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/logimos/concurrent"
)

type Result struct {
    URL    string
    Status int
    Error  error
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    urls := []string{
        "https://example.com",
        "https://google.com",
        "https://github.com",
    }
    
    jobs := make(chan string)
    pool := concurrent.NewPool(5, func(ctx context.Context, url string) (Result, error) {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return Result{URL: url, Error: err}, nil
        }
        
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return Result{URL: url, Error: err}, nil
        }
        defer resp.Body.Close()
        
        // Read body to ensure connection is fully consumed
        io.Copy(io.Discard, resp.Body)
        
        return Result{
            URL:    url,
            Status: resp.StatusCode,
        }, nil
    })
    
    results := pool.Run(ctx, jobs)
    
    // Send jobs
    go func() {
        for _, url := range urls {
            jobs <- url
        }
        close(jobs)
    }()
    
    // Collect results
    for r := range results {
        if r.Error != nil {
            fmt.Printf("%s: ERROR - %v\n", r.URL, r.Error)
        } else {
            fmt.Printf("%s: %d\n", r.URL, r.Status)
        }
    }
}
```

## Example: File Processing

```go
package main

import (
    "context"
    "fmt"
    "os"
    "path/filepath"
    
    "github.com/logimos/concurrent"
)

type FileInfo struct {
    Path string
    Size int64
    Err  error
}

func main() {
    ctx := context.Background()
    
    files := []string{
        "/path/to/file1.txt",
        "/path/to/file2.txt",
        "/path/to/file3.txt",
    }
    
    jobs := make(chan string)
    pool := concurrent.NewPool(10, func(ctx context.Context, path string) (FileInfo, error) {
        info, err := os.Stat(path)
        if err != nil {
            return FileInfo{Path: path, Err: err}, nil
        }
        
        return FileInfo{
            Path: path,
            Size: info.Size(),
        }, nil
    })
    
    results := pool.Run(ctx, jobs)
    
    go func() {
        for _, file := range files {
            jobs <- file
        }
        close(jobs)
    }()
    
    totalSize := int64(0)
    for r := range results {
        if r.Err != nil {
            fmt.Printf("Error processing %s: %v\n", r.Path, r.Err)
        } else {
            fmt.Printf("%s: %d bytes\n", r.Path, r.Size)
            totalSize += r.Size
        }
    }
    
    fmt.Printf("Total size: %d bytes\n", totalSize)
}
```

