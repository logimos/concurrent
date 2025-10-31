# Retry & Circuit Breaker

The `concurrent` package provides robust retry mechanisms with exponential backoff and circuit breaker patterns for handling transient failures.

## Retry

### Basic Retry

The `Retry` function executes a function with configurable retry logic.

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
    
    config := concurrent.DefaultRetryConfig()
    config.MaxRetries = 5
    config.BaseDelay = 100 * time.Millisecond
    
    err := concurrent.Retry(ctx, "some-data", func(ctx context.Context, data string) error {
        // Attempt operation
        return doSomething(data)
    }, config)
    
    if err != nil {
        fmt.Printf("Failed after retries: %v\n", err)
    }
}
```

### Retry Configuration

```go
type RetryConfig struct {
    MaxRetries int           // Maximum number of retry attempts
    BaseDelay  time.Duration // Initial delay between retries
    MaxDelay   time.Duration // Maximum delay cap
    Multiplier float64       // Exponential backoff multiplier
    Jitter     bool          // Add randomness to delays
}
```

### Default Configuration

```go
config := concurrent.DefaultRetryConfig()
// MaxRetries: 3
// BaseDelay: 100ms
// MaxDelay: 5s
// Multiplier: 2.0
// Jitter: true
```

### Custom Configuration

```go
config := concurrent.RetryConfig{
    MaxRetries: 5,
    BaseDelay:  200 * time.Millisecond,
    MaxDelay:   10 * time.Second,
    Multiplier: 2.5,
    Jitter:     true,
}

err := concurrent.Retry(ctx, item, fn, config)
```

## Retry Functions

### `RetryWithBackoff`

Convenience function with exponential backoff:

```go
err := concurrent.RetryWithBackoff(ctx, item, fn, 5, 100*time.Millisecond)
```

### `RetryForever`

Retries indefinitely until success or context cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := concurrent.RetryForever(ctx, item, fn, 1*time.Second)
```

**Warning:** Use with caution and ensure proper context cancellation.

### `WithRetry`

Wraps a function with retry logic:

```go
retryableFn := concurrent.WithRetry(fn, config)

// Use the wrapped function
err := retryableFn(ctx, item)
```

## Retryable Errors

### Marking Errors as Retryable

```go
import "github.com/logimos/concurrent"

// Create a retryable error
err := concurrent.NewRetryableError(someError, true)

// Or non-retryable
err := concurrent.NewRetryableError(someError, false)
```

### Checking if Error is Retryable

```go
if concurrent.IsRetryable(err) {
    // Error can be retried
} else {
    // Error should not be retried
}
```

### Example: Selective Retries

```go
fn := func(ctx context.Context, url string) error {
    resp, err := http.Get(url)
    if err != nil {
        // Network errors are retryable
        return concurrent.NewRetryableError(err, true)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == 404 {
        // 404 is not retryable
        return concurrent.NewRetryableError(fmt.Errorf("not found"), false)
    }
    
    if resp.StatusCode >= 500 {
        // Server errors are retryable
        return concurrent.NewRetryableError(fmt.Errorf("server error"), true)
    }
    
    return nil
}
```

## Circuit Breaker

Circuit breakers prevent cascading failures by stopping requests to a failing service.

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/logimos/concurrent"
)

func main() {
    // Circuit opens after 5 failures, resets after 30 seconds
    cb := concurrent.NewCircuitBreaker(5, 30*time.Second)
    
    ctx := context.Background()
    
    err := cb.Execute(ctx, func() error {
        return callService()
    })
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        fmt.Printf("Circuit state: %v\n", cb.State())
    }
}
```

### Circuit States

- **Closed**: Normal operation, requests pass through
- **Open**: Circuit is open, requests fail immediately
- **Half-Open**: Testing if service recovered, allows one request

### API

#### `NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker`

Creates a circuit breaker.

**Parameters:**
- `failureThreshold`: Number of failures before opening circuit
- `resetTimeout`: Time before attempting to close circuit

#### `Execute(ctx context.Context, fn func() error) error`

Executes a function through the circuit breaker.

#### `State() CircuitState`

Returns the current circuit state.

### Example: API Calls with Circuit Breaker

```go
cb := concurrent.NewCircuitBreaker(5, 30*time.Second)

for _, url := range urls {
    err := cb.Execute(ctx, func() error {
        resp, err := http.Get(url)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 500 {
            return fmt.Errorf("server error: %d", resp.StatusCode)
        }
        
        return nil
    })
    
    if err != nil {
        state := cb.State()
        if state == concurrent.StateOpen {
            fmt.Println("Circuit is open, skipping remaining requests")
            break
        }
        fmt.Printf("Request failed: %v\n", err)
    }
}
```

## Combining Retry and Circuit Breaker

```go
cb := concurrent.NewCircuitBreaker(3, 10*time.Second)
retryConfig := concurrent.DefaultRetryConfig()

fn := func(ctx context.Context, item string) error {
    return cb.Execute(ctx, func() error {
        return makeAPICall(item)
    })
}

err := concurrent.Retry(ctx, item, fn, retryConfig)
```

## Best Practices

1. **Choose retry limits**: Don't retry indefinitely without bounds
2. **Use jitter**: Prevents thundering herd problems
3. **Respect cancellation**: Always use contexts with timeouts
4. **Mark errors appropriately**: Distinguish retryable from non-retryable errors
5. **Monitor circuit state**: Track circuit breaker state for observability
6. **Tune thresholds**: Adjust failure thresholds based on your use case

## Common Patterns

### HTTP Request with Retry

```go
config := concurrent.RetryConfig{
    MaxRetries: 3,
    BaseDelay:  1 * time.Second,
    MaxDelay:   10 * time.Second,
    Multiplier: 2.0,
    Jitter:     true,
}

err := concurrent.Retry(ctx, url, func(ctx context.Context, url string) error {
    req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return concurrent.NewRetryableError(err, true)
    }
    defer resp.Body.Close()
    
    if resp.StatusCode >= 500 {
        return concurrent.NewRetryableError(fmt.Errorf("server error"), true)
    }
    
    return nil
}, config)
```

### Database Operations

```go
cb := concurrent.NewCircuitBreaker(5, 30*time.Second)

err := cb.Execute(ctx, func() error {
    return db.Query(query)
})

if err != nil {
    if cb.State() == concurrent.StateOpen {
        // Fallback to cache or return error
        return getFromCache()
    }
    return err
}
```

