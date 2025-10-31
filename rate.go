package concurrent

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate of operations.
type RateLimiter struct {
	limit      int
	interval   time.Duration
	tokens     chan struct{}
	mu         sync.Mutex
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter with the specified limit and interval.
// For example, NewRateLimiter(100, time.Second) allows 100 operations per second.
func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	if limit <= 0 {
		limit = 1
	}
	if interval <= 0 {
		interval = time.Second
	}

	rl := &RateLimiter{
		limit:      limit,
		interval:   interval,
		tokens:     make(chan struct{}, limit),
		lastRefill: time.Now(),
	}

	// Fill the token bucket initially
	for i := 0; i < limit; i++ {
		rl.tokens <- struct{}{}
	}

	return rl
}

// Allow checks if an operation is allowed under the current rate limit.
// It returns true if the operation is allowed, false otherwise.
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// Wait blocks until an operation is allowed under the rate limit.
// Note: Refill() must be called periodically (e.g., via a background goroutine)
// for tokens to be replenished. Use RateLimit() function for automatic refill.
// This method does not automatically refill tokens - use RateLimit() for that.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	// Try refill once before waiting
	rl.Refill()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-rl.tokens:
		return nil
	}
}

// Refill refills the token bucket based on the elapsed time.
func (rl *RateLimiter) Refill() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.interval {
		// Calculate how many tokens to add
		tokensToAdd := int(elapsed/rl.interval) * rl.limit
		if tokensToAdd > rl.limit {
			tokensToAdd = rl.limit
		}

		// Add tokens to the bucket
		for i := 0; i < tokensToAdd; i++ {
			select {
			case rl.tokens <- struct{}{}:
			default:
				// Bucket is full
				break
			}
		}

		rl.lastRefill = now
	}
}

// RateLimit applies rate limiting to a channel of items.
func RateLimit[T any](ctx context.Context, input <-chan T, limit int, interval time.Duration) <-chan T {
	output := make(chan T)
	limiter := NewRateLimiter(limit, interval)

	go func() {
		defer close(output)

		// Start refill goroutine
		go func() {
			ticker := time.NewTicker(interval / 10) // Check 10 times per interval
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

		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-input:
				if !ok {
					return
				}

				// Wait for rate limit
				if err := limiter.Wait(ctx); err != nil {
					return
				}

				select {
				case <-ctx.Done():
					return
				case output <- item:
				}
			}
		}
	}()

	return output
}

// BurstRateLimit allows bursts up to a maximum size while maintaining an average rate.
type BurstRateLimit struct {
	limit      int
	interval   time.Duration
	burst      int
	tokens     chan struct{}
	mu         sync.Mutex
	lastRefill time.Time
}

// NewBurstRateLimit creates a rate limiter that allows bursts.
func NewBurstRateLimit(limit int, interval time.Duration, burst int) *BurstRateLimit {
	if limit <= 0 {
		limit = 1
	}
	if interval <= 0 {
		interval = time.Second
	}
	if burst <= 0 {
		burst = limit
	}
	if burst > limit*2 {
		burst = limit * 2 // Cap burst at 2x the limit
	}

	brl := &BurstRateLimit{
		limit:      limit,
		interval:   interval,
		burst:      burst,
		tokens:     make(chan struct{}, burst),
		lastRefill: time.Now(),
	}

	// Fill the token bucket initially
	for i := 0; i < burst; i++ {
		brl.tokens <- struct{}{}
	}

	return brl
}

// Allow checks if an operation is allowed under the burst rate limit.
func (brl *BurstRateLimit) Allow() bool {
	select {
	case <-brl.tokens:
		return true
	default:
		return false
	}
}

// Wait blocks until an operation is allowed under the burst rate limit.
func (brl *BurstRateLimit) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-brl.tokens:
		return nil
	}
}

// Refill refills the token bucket for burst rate limiting.
func (brl *BurstRateLimit) Refill() {
	brl.mu.Lock()
	defer brl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(brl.lastRefill)

	if elapsed >= brl.interval {
		// Calculate how many tokens to add
		tokensToAdd := int(elapsed/brl.interval) * brl.limit
		if tokensToAdd > brl.burst {
			tokensToAdd = brl.burst
		}

		// Add tokens to the bucket
		for i := 0; i < tokensToAdd; i++ {
			select {
			case brl.tokens <- struct{}{}:
			default:
				// Bucket is full
				break
			}
		}

		brl.lastRefill = now
	}
}
