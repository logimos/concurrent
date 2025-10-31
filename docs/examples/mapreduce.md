# MapReduce Example

This example demonstrates how to use `MapConcurrent` for concurrent map operations.

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
    
    data := []int{1, 2, 3, 4, 5, 6}
    out, err := concurrent.MapConcurrent(ctx, data, 3, func(ctx context.Context, v int) (int, error) {
        return v * v, nil
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(out) // [1 4 9 16 25 36]
}
```

## Example: Processing URLs

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

type URLResult struct {
    URL        string
    StatusCode int
    Size       int64
    Error      error
}

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    urls := []string{
        "https://example.com",
        "https://google.com",
        "https://github.com",
        "https://golang.org",
    }
    
    results, err := concurrent.MapConcurrent(ctx, urls, 5, func(ctx context.Context, url string) (URLResult, error) {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return URLResult{URL: url, Error: err}, nil
        }
        
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return URLResult{URL: url, Error: err}, nil
        }
        defer resp.Body.Close()
        
        size, _ := io.Copy(io.Discard, resp.Body)
        
        return URLResult{
            URL:        url,
            StatusCode: resp.StatusCode,
            Size:       size,
        }, nil
    })
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("%s: ERROR - %v\n", result.URL, result.Error)
        } else {
            fmt.Printf("%s: %d (%d bytes)\n", result.URL, result.StatusCode, result.Size)
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

type FileResult struct {
    Path    string
    Size    int64
    IsDir   bool
    Error   error
}

func main() {
    ctx := context.Background()
    
    paths := []string{
        "/usr/bin",
        "/etc",
        "/tmp",
    }
    
    results, err := concurrent.MapConcurrent(ctx, paths, 10, func(ctx context.Context, path string) (FileResult, error) {
        info, err := os.Stat(path)
        if err != nil {
            return FileResult{Path: path, Error: err}, nil
        }
        
        return FileResult{
            Path:  path,
            Size:  info.Size(),
            IsDir: info.IsDir(),
        }, nil
    })
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    for _, result := range results {
        if result.Error != nil {
            fmt.Printf("%s: ERROR - %v\n", result.Path, result.Error)
        } else {
            fileType := "file"
            if result.IsDir {
                fileType = "directory"
            }
            fmt.Printf("%s: %s (%d bytes)\n", result.Path, fileType, result.Size)
        }
    }
}
```

## Example: Data Transformation

```go
package main

import (
    "context"
    "fmt"
    "strings"
    
    "github.com/logimos/concurrent"
)

type User struct {
    Name  string
    Email string
}

type ProcessedUser struct {
    Name      string
    Email     string
    Domain    string
    NameUpper string
}

func main() {
    ctx := context.Background()
    
    users := []User{
        {Name: "Alice", Email: "alice@example.com"},
        {Name: "Bob", Email: "bob@example.com"},
        {Name: "Charlie", Email: "charlie@test.com"},
    }
    
    results, err := concurrent.MapConcurrent(ctx, users, 3, func(ctx context.Context, user User) (ProcessedUser, error) {
        parts := strings.Split(user.Email, "@")
        domain := ""
        if len(parts) == 2 {
            domain = parts[1]
        }
        
        return ProcessedUser{
            Name:      user.Name,
            Email:     user.Email,
            Domain:    domain,
            NameUpper: strings.ToUpper(user.Name),
        }, nil
    })
    
    if err != nil {
        panic(err)
    }
    
    for _, result := range results {
        fmt.Printf("%s (%s) - Domain: %s\n", result.NameUpper, result.Email, result.Domain)
    }
}
```

