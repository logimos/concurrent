# API Reference

Complete API documentation for the `concurrent` package.

## Package Overview

Package `concurrent` provides concurrency helpers for Go applications.

## Types

### Pool

```go
type Pool[T any, R any] struct {
    // ...
}
```

Worker pool for concurrent job processing.

**Methods:**
- `NewPool[T, R](n int, fn func(context.Context, T) (R, error)) *Pool[T, R]`
- `Run(ctx context.Context, jobs <-chan T) <-chan R`

### Pipeline

```go
type Pipeline[T any] struct {
    // ...
}
```

Composable data processing pipeline.

**Methods:**
- `NewPipeline[T](ctx context.Context) *Pipeline[T]`
- `AddStage(stage Stage[T, T]) *Pipeline[T]`
- `Run(input <-chan T) <-chan T`
- `Close()`

### PipelineBuilder

```go
type PipelineBuilder[T any] struct {
    // ...
}
```

Fluent builder for pipelines.

**Methods:**
- `NewPipelineBuilder[T](ctx context.Context) *PipelineBuilder[T]`
- `AddStage(stage Stage[T, T]) *PipelineBuilder[T]`
- `Build() *Pipeline[T]`

### RateLimiter

```go
type RateLimiter struct {
    // ...
}
```

Token bucket rate limiter.

**Methods:**
- `NewRateLimiter(limit int, interval time.Duration) *RateLimiter`
- `Allow() bool`
- `Wait(ctx context.Context) error`
- `Refill()`

### BurstRateLimit

```go
type BurstRateLimit struct {
    // ...
}
```

Burst-capable rate limiter.

**Methods:**
- `NewBurstRateLimit(limit int, interval time.Duration, burst int) *BurstRateLimit`
- `Allow() bool`
- `Wait(ctx context.Context) error`
- `Refill()`

### RetryConfig

```go
type RetryConfig struct {
    MaxRetries int
    BaseDelay  time.Duration
    MaxDelay   time.Duration
    Multiplier float64
    Jitter     bool
}
```

Configuration for retry behavior.

### CircuitBreaker

```go
type CircuitBreaker struct {
    // ...
}
```

Circuit breaker for failure handling.

**Methods:**
- `NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker`
- `Execute(ctx context.Context, fn func() error) error`
- `State() CircuitState`

### CircuitState

```go
type CircuitState int
```

Circuit breaker state.

**Constants:**
- `StateClosed`
- `StateOpen`
- `StateHalfOpen`

## Functions

### MapConcurrent

```go
func MapConcurrent[T any, R any](
    ctx context.Context,
    in []T,
    n int,
    fn func(context.Context, T) (R, error),
) ([]R, error)
```

Processes a slice concurrently with bounded parallelism.

### FanOut

```go
func FanOut[T any, R any](
    ctx context.Context,
    input <-chan T,
    workers int,
    fn func(context.Context, T) (R, error),
) <-chan R
```

Distributes work to multiple workers.

### FanIn

```go
func FanIn[T any](
    ctx context.Context,
    inputs ...<-chan T,
) <-chan T
```

Merges multiple input channels.

### FanOutFanIn

```go
func FanOutFanIn[T any, R any](
    ctx context.Context,
    input <-chan T,
    workers int,
    fn func(context.Context, T) (R, error),
) <-chan R
```

Combines fan-out and fan-in.

### RoundRobin

```go
func RoundRobin[T any, R any](
    ctx context.Context,
    input <-chan T,
    workers int,
    fn func(context.Context, T) (R, error),
) <-chan R
```

Distributes work in round-robin fashion.

### RateLimit

```go
func RateLimit[T any](
    ctx context.Context,
    input <-chan T,
    limit int,
    interval time.Duration,
) <-chan T
```

Rate limits a channel of items.

### Retry

```go
func Retry[T any](
    ctx context.Context,
    item T,
    fn RetryableFunc[T],
    config RetryConfig,
) error
```

Executes a function with retry logic.

### RetryWithBackoff

```go
func RetryWithBackoff[T any](
    ctx context.Context,
    item T,
    fn RetryableFunc[T],
    maxRetries int,
    baseDelay time.Duration,
) error
```

Retries with exponential backoff.

### RetryForever

```go
func RetryForever[T any](
    ctx context.Context,
    item T,
    fn RetryableFunc[T],
    baseDelay time.Duration,
) error
```

Retries indefinitely until success.

### WithRetry

```go
func WithRetry[T any](
    fn RetryableFunc[T],
    config RetryConfig,
) RetryableFunc[T]
```

Wraps a function with retry logic.

### DefaultRetryConfig

```go
func DefaultRetryConfig() RetryConfig
```

Returns default retry configuration.

### NewRetryableError

```go
func NewRetryableError(err error, retryable bool) RetryableError
```

Creates a retryable error.

### IsRetryable

```go
func IsRetryable(err error) bool
```

Checks if an error is retryable.

## Pipeline Stages

### Map

```go
func Map[T any](fn func(T) T) Stage[T, T]
```

Applies a function to each item.

### Filter

```go
func Filter[T any](predicate func(T) bool) Stage[T, T]
```

Filters items based on a predicate.

### Batch

```go
func Batch[T any](size int) Stage[T, []T]
```

Batches items into slices.

### Unbatch

```go
func Unbatch[T any]() Stage[[]T, T]
```

Unbatches slices into individual items.

### Tee

```go
func Tee[T any](outputs ...chan<- T) Stage[T, T]
```

Splits input to multiple outputs.

### Merge

```go
func Merge[T any](ctx context.Context, inputs ...<-chan T) <-chan T
```

Merges multiple inputs into one output.

## Type Aliases

### Stage

```go
type Stage[T any, R any] func(context.Context, <-chan T) <-chan R
```

A pipeline stage transformation function.

### RetryableFunc

```go
type RetryableFunc[T any] func(context.Context, T) error
```

A function that can be retried.

### RetryableError

```go
type RetryableError struct {
    Err       error
    Retryable bool
}
```

An error that indicates retryability.

