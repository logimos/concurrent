# Rate Limiting

Rate limiting controls the rate of operations to prevent overwhelming downstream systems. The `concurrent` package provides token bucket rate limiters with burst support.

## Overview

Rate limiters use a token bucket algorithm:
- Tokens are added to the bucket at a fixed rate
- Operations consume tokens
- Operations wait if no tokens are available

## Basic Rate Limiter

### Example

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/logimos/concurrent"
)

func main() {
    // Allow 10 operations per second
    limiter := concurrent.NewRateLimiter(10, time.Second)
    
    ctx := context.Background()
    
    for i := 0; i < 20; i++ {
        if err := limiter.Wait(ctx); err != nil {
            break
        }
        
        // Perform operation
        fmt.Printf("Operation %d\n", i)
    }
}
```

### API

#### `NewRateLimiter(limit int, interval time.Duration) *RateLimiter`

Creates a rate limiter allowing `limit` operations per `interval`.

**Example:** `NewRateLimiter(100, time.Second)` allows 100 operations per second.

#### `Allow() bool`

Checks if an operation is allowed without blocking. Returns `true` if allowed, `false` otherwise.

#### `Wait(ctx context.Context) error`

Blocks until an operation is allowed. Returns an error if the context is canceled.

**Note:** `Refill()` must be called periodically for tokens to be replenished. See `RateLimit()` for automatic refill.

#### `Refill()`

Refills the token bucket based on elapsed time. Should be called periodically.

## Rate Limit Channel

The `RateLimit` function automatically handles token refilling and provides a channel-based interface.

### Example

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
    
    input := make(chan string)
    
    // Rate limit to 5 items per second
    output := concurrent.RateLimit(ctx, input, 5, time.Second)
    
    // Send items
    go func() {
        for i := 0; i < 20; i++ {
            input <- fmt.Sprintf("item-%d", i)
        }
        close(input)
    }()
    
    // Process rate-limited items
    for item := range output {
        fmt.Println(item)
    }
}
```

### API

```go
func RateLimit[T any](ctx context.Context, input <-chan T, limit int, interval time.Duration) <-chan T
```

**Parameters:**
- `ctx`: Context for cancellation
- `input`: Input channel
- `limit`: Maximum operations per interval
- `interval`: Time interval

**Returns:** Rate-limited output channel

## Burst Rate Limiter

Burst rate limiters allow bursts up to a maximum size while maintaining an average rate.

### Example

```go
// Allow bursts of 20, but average 10 per second
limiter := concurrent.NewBurstRateLimit(10, time.Second, 20)

ctx := context.Background()

// Can process 20 items immediately, then rate-limited
for i := 0; i < 30; i++ {
    if err := limiter.Wait(ctx); err != nil {
        break
    }
    processItem(i)
}
```

### API

#### `NewBurstRateLimit(limit int, interval time.Duration, burst int) *BurstRateLimit`

Creates a burst rate limiter.

**Parameters:**
- `limit`: Average operations per interval
- `interval`: Time interval
- `burst`: Maximum burst size (capped at 2x limit)

## Use Cases

### API Rate Limiting

```go
limiter := concurrent.NewRateLimiter(100, time.Minute) // 100 requests per minute

for _, url := range urls {
    if err := limiter.Wait(ctx); err != nil {
        return err
    }
    
    resp, err := http.Get(url)
    // Process response
}
```

### Pipeline Rate Limiting

```go
input := make(chan *Request)

// Limit processing to 50 requests per second
output := concurrent.RateLimit(ctx, input, 50, time.Second)

for req := range output {
    processRequest(req)
}
```

### Burst Traffic Handling

```go
// Allow bursts of 1000, average 100/second
limiter := concurrent.NewBurstRateLimit(100, time.Second, 1000)

// Handle traffic spikes gracefully
for req := range requests {
    if err := limiter.Wait(ctx); err != nil {
        return err
    }
    handleRequest(req)
}
```

## Best Practices

1. **Choose appropriate limits**: Balance throughput with downstream capacity
2. **Use burst limiters**: For handling traffic spikes
3. **Monitor token consumption**: Track how often tokens are exhausted
4. **Context cancellation**: Always use contexts with timeouts
5. **Refill frequency**: RateLimit automatically refills; manual limiters need periodic Refill()

## Implementation Details

- **Token Bucket**: Uses a buffered channel to store tokens
- **Refill Strategy**: Tokens are added based on elapsed time since last refill
- **Thread Safety**: All operations are thread-safe
- **Cancellation**: All operations respect context cancellation

## Common Patterns

### Background Refill

```go
limiter := concurrent.NewRateLimiter(10, time.Second)

// Refill in background
go func() {
    ticker := time.NewTicker(100 * time.Millisecond)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            limiter.Refill()
        }
    }
}()

// Use limiter
for {
    if err := limiter.Wait(ctx); err != nil {
        break
    }
    doWork()
}
```

### Non-blocking Check

```go
if limiter.Allow() {
    // Process immediately
    doWork()
} else {
    // Rate limited, skip or queue
    queueForLater()
}
```

