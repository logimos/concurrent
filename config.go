package concurrent

import (
	"context"
	"time"
)

// PoolOptions holds configuration options for worker pools.
type PoolOptions struct {
	Workers    int
	BufferSize int
	Timeout    time.Duration
	RetryCount int
	Backoff    time.Duration
	RateLimit  *RateLimitOptions
}

// RateLimitOptions holds configuration for rate limiting.
type RateLimitOptions struct {
	Limit    int
	Interval time.Duration
	Burst    int
}

// DefaultPoolOptions returns sensible default options for a worker pool.
func DefaultPoolOptions() PoolOptions {
	return PoolOptions{
		Workers:    4,
		BufferSize: 100,
		Timeout:    30 * time.Second,
		RetryCount: 3,
		Backoff:    100 * time.Millisecond,
		RateLimit: &RateLimitOptions{
			Limit:    100,
			Interval: time.Second,
			Burst:    200,
		},
	}
}

// PoolOption is a function that configures a PoolOptions.
type PoolOption func(*PoolOptions)

// WithWorkers sets the number of workers.
func WithWorkers(workers int) PoolOption {
	return func(opts *PoolOptions) {
		opts.Workers = workers
	}
}

// WithBufferSize sets the buffer size for channels.
func WithBufferSize(size int) PoolOption {
	return func(opts *PoolOptions) {
		opts.BufferSize = size
	}
}

// WithTimeout sets the timeout for operations.
func WithTimeout(timeout time.Duration) PoolOption {
	return func(opts *PoolOptions) {
		opts.Timeout = timeout
	}
}

// WithRetryConfig sets the retry configuration.
func WithRetryConfig(count int, backoff time.Duration) PoolOption {
	return func(opts *PoolOptions) {
		opts.RetryCount = count
		opts.Backoff = backoff
	}
}

// WithRateLimit sets the rate limiting configuration.
func WithRateLimit(limit int, interval time.Duration, burst int) PoolOption {
	return func(opts *PoolOptions) {
		opts.RateLimit = &RateLimitOptions{
			Limit:    limit,
			Interval: interval,
			Burst:    burst,
		}
	}
}

// Metrics holds performance metrics for concurrent operations.
type Metrics struct {
	ProcessedCount int64
	ErrorCount     int64
	Duration       time.Duration
	StartTime      time.Time
	EndTime        time.Time
}

// NewMetrics creates a new metrics instance.
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime: time.Now(),
	}
}

// RecordSuccess records a successful operation.
func (m *Metrics) RecordSuccess() {
	m.ProcessedCount++
}

// RecordError records a failed operation.
func (m *Metrics) RecordError() {
	m.ErrorCount++
}

// Finish marks the end of the operation and calculates duration.
func (m *Metrics) Finish() {
	m.EndTime = time.Now()
	m.Duration = m.EndTime.Sub(m.StartTime)
}

// SuccessRate returns the success rate as a percentage.
func (m *Metrics) SuccessRate() float64 {
	total := m.ProcessedCount + m.ErrorCount
	if total == 0 {
		return 0
	}
	return float64(m.ProcessedCount) / float64(total) * 100
}

// Throughput returns the operations per second.
func (m *Metrics) Throughput() float64 {
	if m.Duration == 0 {
		return 0
	}
	return float64(m.ProcessedCount) / m.Duration.Seconds()
}

// ErrorRate returns the error rate as a percentage.
func (m *Metrics) ErrorRate() float64 {
	return 100 - m.SuccessRate()
}

// ContextOptions holds options for context handling.
type ContextOptions struct {
	Timeout    time.Duration
	CancelFunc func()
}

// DefaultContextOptions returns default context options.
func DefaultContextOptions() ContextOptions {
	return ContextOptions{
		Timeout: 30 * time.Second,
	}
}

// WithContextTimeout creates a context with timeout.
func WithContextTimeout(timeout time.Duration) ContextOptions {
	return ContextOptions{
		Timeout: timeout,
	}
}

// CreateContext creates a context with the given options.
func CreateContext(opts ContextOptions) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if opts.Timeout > 0 {
		return context.WithTimeout(ctx, opts.Timeout)
	}
	return context.WithCancel(ctx)
}

// BackpressureOptions holds configuration for backpressure handling.
type BackpressureOptions struct {
	MaxBufferSize int
	DropOldest    bool
	BlockOnFull   bool
}

// DefaultBackpressureOptions returns default backpressure options.
func DefaultBackpressureOptions() BackpressureOptions {
	return BackpressureOptions{
		MaxBufferSize: 1000,
		DropOldest:    false,
		BlockOnFull:   true,
	}
}

// BackpressureOption is a function that configures backpressure options.
type BackpressureOption func(*BackpressureOptions)

// WithMaxBufferSize sets the maximum buffer size.
func WithMaxBufferSize(size int) BackpressureOption {
	return func(opts *BackpressureOptions) {
		opts.MaxBufferSize = size
	}
}

// WithDropOldest sets whether to drop oldest items when buffer is full.
func WithDropOldest(drop bool) BackpressureOption {
	return func(opts *BackpressureOptions) {
		opts.DropOldest = drop
	}
}

// WithBlockOnFull sets whether to block when buffer is full.
func WithBlockOnFull(block bool) BackpressureOption {
	return func(opts *BackpressureOptions) {
		opts.BlockOnFull = block
	}
}
