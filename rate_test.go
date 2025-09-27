package concurrent

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		rl := NewRateLimiter(2, 100*time.Millisecond)

		// Should allow first two operations
		if !rl.Allow() {
			t.Error("Expected first operation to be allowed")
		}
		if !rl.Allow() {
			t.Error("Expected second operation to be allowed")
		}

		// Third operation should be denied
		if rl.Allow() {
			t.Error("Expected third operation to be denied")
		}
	})

	t.Run("wait functionality", func(t *testing.T) {
		rl := NewRateLimiter(1, 50*time.Millisecond)

		// First operation should be allowed
		if err := rl.Wait(context.Background()); err != nil {
			t.Errorf("Expected first operation to be allowed, got error: %v", err)
		}

		// Second operation should be denied initially
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		if err := rl.Wait(ctx); err == nil {
			t.Error("Expected second operation to be denied")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		rl := NewRateLimiter(1, 100*time.Millisecond)

		// Use up the token
		rl.Allow()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		if err := rl.Wait(ctx); err == nil {
			t.Error("Expected context cancellation")
		}
	})

	t.Run("refill functionality", func(t *testing.T) {
		rl := NewRateLimiter(2, 50*time.Millisecond)

		// Use up both tokens
		rl.Allow()
		rl.Allow()

		// Should be denied
		if rl.Allow() {
			t.Error("Expected operation to be denied")
		}

		// Wait for refill
		time.Sleep(60 * time.Millisecond)
		rl.Refill()

		// Should be allowed again
		if !rl.Allow() {
			t.Error("Expected operation to be allowed after refill")
		}
	})
}

func TestRateLimit(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output := RateLimit(ctx, input, 2, 100*time.Millisecond)

		// Send data
		go func() {
			for i := 0; i < 5; i++ {
				input <- i
			}
			close(input)
		}()

		var results []int
		start := time.Now()
		for v := range output {
			results = append(results, v)
		}
		duration := time.Since(start)

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		// Should take at least 200ms (2 batches of 2 items each, 100ms apart)
		if duration < 200*time.Millisecond {
			t.Errorf("Expected duration >= 200ms, got %v", duration)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		input := make(chan int)
		output := RateLimit(ctx, input, 1, 100*time.Millisecond)

		go func() {
			for i := 0; i < 10; i++ {
				select {
				case input <- i:
				case <-ctx.Done():
					close(input)
					return
				}
			}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		// Should have few results due to timeout
		t.Logf("Got %d results before timeout", len(results))
	})
}

func TestBurstRateLimit(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		brl := NewBurstRateLimit(2, 100*time.Millisecond, 4)

		// Should allow up to 4 operations (burst)
		for i := 0; i < 4; i++ {
			if !brl.Allow() {
				t.Errorf("Expected operation %d to be allowed", i+1)
			}
		}

		// Fifth operation should be denied
		if brl.Allow() {
			t.Error("Expected fifth operation to be denied")
		}
	})

	t.Run("wait functionality", func(t *testing.T) {
		brl := NewBurstRateLimit(1, 50*time.Millisecond, 2)

		// First two operations should be allowed
		if err := brl.Wait(context.Background()); err != nil {
			t.Errorf("Expected first operation to be allowed, got error: %v", err)
		}
		if err := brl.Wait(context.Background()); err != nil {
			t.Errorf("Expected second operation to be allowed, got error: %v", err)
		}

		// Third operation should be denied initially
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		if err := brl.Wait(ctx); err == nil {
			t.Error("Expected third operation to be denied")
		}
	})

	t.Run("refill functionality", func(t *testing.T) {
		brl := NewBurstRateLimit(2, 50*time.Millisecond, 3)

		// Use up all tokens
		for i := 0; i < 3; i++ {
			brl.Allow()
		}

		// Should be denied
		if brl.Allow() {
			t.Error("Expected operation to be denied")
		}

		// Wait for refill
		time.Sleep(60 * time.Millisecond)
		brl.Refill()

		// Should be allowed again
		if !brl.Allow() {
			t.Error("Expected operation to be allowed after refill")
		}
	})
}

func BenchmarkRateLimiter(b *testing.B) {
	rl := NewRateLimiter(1000, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

func BenchmarkBurstRateLimit(b *testing.B) {
	brl := NewBurstRateLimit(1000, time.Second, 2000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		brl.Allow()
	}
}
