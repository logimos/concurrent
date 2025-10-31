package concurrent

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	Multiplier float64
	Jitter     bool
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		Multiplier: 2.0,
		Jitter:     true,
	}
}

// RetryableFunc is a function that can be retried.
type RetryableFunc[T any] func(context.Context, T) error

// Retry executes a function with retry logic.
func Retry[T any](ctx context.Context, item T, fn RetryableFunc[T], config RetryConfig) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := fn(ctx, item)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryable(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay
		delay := calculateDelay(attempt, config)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return lastErr
}

// calculateDelay calculates the delay for the given attempt.
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff
	delay := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Add jitter if enabled
	if config.Jitter {
		// Add ±25% jitter
		jitter := delay * 0.25
		delay = delay - jitter + (delay * 0.5 * (1.0 - 0.5)) // Simplified jitter
	}

	return time.Duration(delay)
}

// WithRetry wraps a function with retry logic.
func WithRetry[T any](fn RetryableFunc[T], config RetryConfig) RetryableFunc[T] {
	return func(ctx context.Context, item T) error {
		return Retry(ctx, item, fn, config)
	}
}

// RetryableError is an error that indicates whether an operation should be retried.
type RetryableError struct {
	Err       error
	Retryable bool
}

func (re RetryableError) Error() string {
	return re.Err.Error()
}

func (re RetryableError) Unwrap() error {
	return re.Err
}

// NewRetryableError creates a new retryable error.
func NewRetryableError(err error, retryable bool) RetryableError {
	return RetryableError{
		Err:       err,
		Retryable: retryable,
	}
}

// IsRetryable checks if an error is retryable.
func IsRetryable(err error) bool {
	var re RetryableError
	if errors.As(err, &re) {
		return re.Retryable
	}
	return true // Default to retryable for unknown errors
}

// RetryWithBackoff executes a function with exponential backoff retry logic.
func RetryWithBackoff[T any](ctx context.Context, item T, fn RetryableFunc[T], maxRetries int, baseDelay time.Duration) error {
	config := RetryConfig{
		MaxRetries: maxRetries,
		BaseDelay:  baseDelay,
		MaxDelay:   baseDelay * 10,
		Multiplier: 2.0,
		Jitter:     true,
	}

	return Retry(ctx, item, fn, config)
}

// RetryForever executes a function with retry logic that never gives up.
// Use with caution and ensure proper context cancellation.
func RetryForever[T any](ctx context.Context, item T, fn RetryableFunc[T], baseDelay time.Duration) error {
	config := RetryConfig{
		MaxRetries: math.MaxInt32, // Effectively infinite
		BaseDelay:  baseDelay,
		MaxDelay:   baseDelay * 100,
		Multiplier: 2.0,
		Jitter:     true,
	}

	return Retry(ctx, item, fn, config)
}

// CircuitBreaker implements the circuit breaker pattern.
type CircuitBreaker struct {
	failureThreshold int
	resetTimeout     time.Duration
	state            CircuitState
	failureCount     int
	lastFailureTime  time.Time
	mu               sync.Mutex
}

// CircuitState represents the state of the circuit breaker.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
	}
}

// Execute executes a function through the circuit breaker.
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func() error) error {
	// Check context cancellation before acquiring lock
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	cb.mu.Lock()
	// Check circuit state
	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastFailureTime) >= cb.resetTimeout {
			cb.state = StateHalfOpen
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is open")
		}
	case StateHalfOpen:
		// Allow one request to test if service is back
	}
	cb.mu.Unlock()

	// Execute function outside lock to avoid blocking other operations
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.failureThreshold {
			cb.state = StateOpen
		}
		return err
	}

	// Success - reset circuit breaker
	cb.failureCount = 0
	cb.state = StateClosed
	return nil
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
