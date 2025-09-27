package concurrent

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestFanOut(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		// Create fan-out with 3 workers
		output := FanOut(ctx, input, 3, func(_ context.Context, v int) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return v * 2, nil
		})

		// Send test data
		go func() {
			for i := 1; i <= 5; i++ {
				input <- i
			}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		if len(results) != 5 {
			t.Errorf("Expected 5 results, got %d", len(results))
		}

		// Check that all results are even (multiplied by 2)
		for _, v := range results {
			if v%2 != 0 {
				t.Errorf("Expected even number, got %d", v)
			}
		}
	})

	t.Run("error handling", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output := FanOut(ctx, input, 2, func(_ context.Context, v int) (int, error) {
			if v%2 == 0 {
				return 0, errors.New("even numbers fail")
			}
			return v * 2, nil
		})

		go func() {
			for i := 1; i <= 4; i++ {
				input <- i
			}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		// Only odd numbers should succeed
		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		input := make(chan int)
		output := FanOut(ctx, input, 2, func(_ context.Context, v int) (int, error) {
			time.Sleep(100 * time.Millisecond) // Longer than timeout
			return v * 2, nil
		})

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

		t.Logf("Got %d results before timeout", len(results))
	})

	t.Run("zero workers", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output := FanOut(ctx, input, 0, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})

		go func() {
			input <- 1
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})
}

func TestFanIn(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()

		// Create multiple input channels
		input1 := make(chan int)
		input2 := make(chan int)
		input3 := make(chan int)

		output := FanIn(ctx, input1, input2, input3)

		// Send data to inputs
		go func() {
			input1 <- 1
			input2 <- 2
			input3 <- 3
			close(input1)
			close(input2)
			close(input3)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	t.Run("empty inputs", func(t *testing.T) {
		ctx := context.Background()
		output := FanIn[int](ctx)

		// Should close immediately
		select {
		case <-output:
			t.Error("Expected output to be closed")
		default:
			// Good, output is closed
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		input := make(chan int)
		output := FanIn(ctx, input)

		go func() {
			time.Sleep(20 * time.Millisecond) // Longer than timeout
			input <- 1
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		// Should have no results due to timeout
		if len(results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(results))
		}
	})
}

func TestFanOutFanIn(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output := FanOutFanIn(ctx, input, 3, func(_ context.Context, v int) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return v * 2, nil
		})

		go func() {
			for i := 1; i <= 6; i++ {
				input <- i
			}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		if len(results) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results))
		}
	})
}

func TestRoundRobin(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output := RoundRobin(ctx, input, 3, func(_ context.Context, v int) (int, error) {
			time.Sleep(10 * time.Millisecond)
			return v * 2, nil
		})

		go func() {
			for i := 1; i <= 6; i++ {
				input <- i
			}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		if len(results) != 6 {
			t.Errorf("Expected 6 results, got %d", len(results))
		}
	})

	t.Run("zero workers", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output := RoundRobin(ctx, input, 0, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})

		go func() {
			input <- 1
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})
}

func BenchmarkFanOut(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := make(chan int, 100)
		output := FanOut(ctx, input, 4, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})

		go func() {
			for j := 0; j < 100; j++ {
				input <- j
			}
			close(input)
		}()

		for range output {
			// Consume results
		}
	}
}

func BenchmarkFanIn(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input1 := make(chan int, 50)
		input2 := make(chan int, 50)
		output := FanIn(ctx, input1, input2)

		go func() {
			for j := 0; j < 50; j++ {
				input1 <- j
				input2 <- j + 50
			}
			close(input1)
			close(input2)
		}()

		for range output {
			// Consume results
		}
	}
}
