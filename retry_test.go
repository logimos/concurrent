package concurrent

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	t.Run("success on first attempt", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()

		attempts := 0
		err := Retry(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			return nil
		}, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})

	t.Run("success after retries", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()
		config.MaxRetries = 3
		config.BaseDelay = 10 * time.Millisecond

		attempts := 0
		err := Retry(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		}, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("failure after max retries", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()
		config.MaxRetries = 2
		config.BaseDelay = 10 * time.Millisecond

		attempts := 0
		err := Retry(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			return errors.New("permanent error")
		}, config)

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if attempts != 3 { // 1 initial + 2 retries
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		config := DefaultRetryConfig()
		config.MaxRetries = 10
		config.BaseDelay = 100 * time.Millisecond

		attempts := 0
		err := Retry(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			return errors.New("temporary error")
		}, config)

		if err == nil {
			t.Error("Expected context error, got nil")
		}
		if attempts < 1 {
			t.Errorf("Expected at least 1 attempt, got %d", attempts)
		}
	})

	t.Run("retryable error", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()
		config.MaxRetries = 2
		config.BaseDelay = 10 * time.Millisecond

		attempts := 0
		err := Retry(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			if attempts < 2 {
				return NewRetryableError(errors.New("temporary error"), true)
			}
			return nil
		}, config)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()
		config.MaxRetries = 3
		config.BaseDelay = 10 * time.Millisecond

		attempts := 0
		err := Retry(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			return NewRetryableError(errors.New("permanent error"), false)
		}, config)

		if err == nil {
			t.Error("Expected error, got nil")
		}
		if attempts != 1 {
			t.Errorf("Expected 1 attempt, got %d", attempts)
		}
	})
}

func TestWithRetry(t *testing.T) {
	t.Run("wrapped function", func(t *testing.T) {
		ctx := context.Background()
		config := DefaultRetryConfig()
		config.MaxRetries = 2
		config.BaseDelay = 10 * time.Millisecond

		attempts := 0
		wrappedFn := WithRetry(func(_ context.Context, item string) error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		}, config)

		err := wrappedFn(ctx, "test")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 2 {
			t.Errorf("Expected 2 attempts, got %d", attempts)
		}
	})
}

func TestRetryWithBackoff(t *testing.T) {
	t.Run("exponential backoff", func(t *testing.T) {
		ctx := context.Background()

		attempts := 0
		err := RetryWithBackoff(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		}, 3, 10*time.Millisecond)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})
}

func TestRetryForever(t *testing.T) {
	t.Run("eventual success", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()

		attempts := 0
		err := RetryForever(ctx, "test", func(_ context.Context, item string) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		}, 10*time.Millisecond)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if attempts != 3 {
			t.Errorf("Expected 3 attempts, got %d", attempts)
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	t.Run("closed state", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)

		if cb.State() != StateClosed {
			t.Errorf("Expected closed state, got %v", cb.State())
		}

		// First operation should succeed
		err := cb.Execute(context.Background(), func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("open state", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 100*time.Millisecond)

		// Cause failures to open circuit
		cb.Execute(context.Background(), func() error {
			return errors.New("error")
		})
		cb.Execute(context.Background(), func() error {
			return errors.New("error")
		})

		if cb.State() != StateOpen {
			t.Errorf("Expected open state, got %v", cb.State())
		}

		// Operations should be blocked
		err := cb.Execute(context.Background(), func() error {
			return nil
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("half-open state", func(t *testing.T) {
		cb := NewCircuitBreaker(2, 50*time.Millisecond)

		// Cause failures to open circuit
		cb.Execute(context.Background(), func() error {
			return errors.New("error")
		})
		cb.Execute(context.Background(), func() error {
			return errors.New("error")
		})

		// Wait for reset timeout
		time.Sleep(60 * time.Millisecond)

		// One operation should be allowed (this will transition to half-open)
		err := cb.Execute(context.Background(), func() error {
			return nil
		})

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Circuit should be closed again
		if cb.State() != StateClosed {
			t.Errorf("Expected closed state, got %v", cb.State())
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		cb := NewCircuitBreaker(1, 100*time.Millisecond)

		// Cause failure to open circuit
		cb.Execute(context.Background(), func() error {
			return errors.New("error")
		})

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		err := cb.Execute(ctx, func() error {
			return nil
		})

		if err == nil {
			t.Error("Expected context error, got nil")
		}
	})
}

func TestRetryableError(t *testing.T) {
	t.Run("retryable error", func(t *testing.T) {
		err := NewRetryableError(errors.New("test error"), true)

		if !IsRetryable(err) {
			t.Error("Expected error to be retryable")
		}

		if err.Error() != "test error" {
			t.Errorf("Expected error message 'test error', got %s", err.Error())
		}
	})

	t.Run("non-retryable error", func(t *testing.T) {
		err := NewRetryableError(errors.New("test error"), false)

		if IsRetryable(err) {
			t.Error("Expected error to be non-retryable")
		}
	})

	t.Run("unknown error", func(t *testing.T) {
		err := errors.New("unknown error")

		if !IsRetryable(err) {
			t.Error("Expected unknown error to be retryable")
		}
	})
}

func BenchmarkRetry(b *testing.B) {
	ctx := context.Background()
	config := DefaultRetryConfig()
	config.MaxRetries = 1
	config.BaseDelay = 1 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Retry(ctx, "test", func(_ context.Context, item string) error {
			return nil
		}, config)
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	cb := NewCircuitBreaker(100, time.Second)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Execute(context.Background(), func() error {
			return nil
		})
	}
}
