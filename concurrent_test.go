package concurrent

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"
)

// TestPool tests the worker pool functionality
func TestPool(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()
		jobs := make(chan int)

		pool := NewPool[int, string](3, func(_ context.Context, v int) (string, error) {
			time.Sleep(10 * time.Millisecond)
			return fmt.Sprintf("job-%d", v), nil
		})

		results := pool.Run(ctx, jobs)

		go func() {
			for i := 0; i < 5; i++ {
				jobs <- i
			}
			close(jobs)
		}()

		var resultsSlice []string
		for r := range results {
			resultsSlice = append(resultsSlice, r)
		}

		if len(resultsSlice) != 5 {
			t.Errorf("Expected 5 results, got %d", len(resultsSlice))
		}
	})

	t.Run("error handling", func(t *testing.T) {
		ctx := context.Background()
		jobs := make(chan int)

		pool := NewPool[int, string](2, func(_ context.Context, v int) (string, error) {
			if v%2 == 0 {
				return "", errors.New("even numbers fail")
			}
			return fmt.Sprintf("job-%d", v), nil
		})

		results := pool.Run(ctx, jobs)

		go func() {
			for i := 0; i < 4; i++ {
				jobs <- i
			}
			close(jobs)
		}()

		var resultsSlice []string
		for r := range results {
			resultsSlice = append(resultsSlice, r)
		}

		// Only odd numbers should succeed
		if len(resultsSlice) != 2 {
			t.Errorf("Expected 2 results, got %d", len(resultsSlice))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		jobs := make(chan int)
		pool := NewPool[int, string](1, func(_ context.Context, v int) (string, error) {
			time.Sleep(100 * time.Millisecond) // Longer than timeout
			return fmt.Sprintf("job-%d", v), nil
		})

		results := pool.Run(ctx, jobs)

		go func() {
			for i := 0; i < 10; i++ {
				select {
				case jobs <- i:
				case <-ctx.Done():
					close(jobs)
					return
				}
			}
			close(jobs)
		}()

		var resultsSlice []string
		for r := range results {
			resultsSlice = append(resultsSlice, r)
		}

		// Should have few or no results due to timeout
		t.Logf("Got %d results before timeout", len(resultsSlice))
	})

	t.Run("zero workers", func(t *testing.T) {
		pool := NewPool[int, string](0, func(_ context.Context, v int) (string, error) {
			return fmt.Sprintf("job-%d", v), nil
		})

		if pool.workers != 1 {
			t.Errorf("Expected 1 worker for zero input, got %d", pool.workers)
		}
	})
}

// TestMapConcurrent tests the concurrent map functionality
func TestMapConcurrent(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		ctx := context.Background()
		in := []int{1, 2, 3, 4}
		out, err := MapConcurrent(ctx, in, 2, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		expected := []int{2, 4, 6, 8}
		if len(out) != len(expected) {
			t.Errorf("Expected length %d, got %d", len(expected), len(out))
		}

		for i, v := range out {
			if v != expected[i] {
				t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
			}
		}
	})

	t.Run("error handling", func(t *testing.T) {
		ctx := context.Background()
		in := []int{1, 2, 3, 4}
		_, err := MapConcurrent(ctx, in, 2, func(_ context.Context, v int) (int, error) {
			if v == 2 {
				return 0, errors.New("error on 2")
			}
			return v * 2, nil
		})

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		in := make([]int, 100)
		for i := range in {
			in[i] = i
		}

		_, err := MapConcurrent(ctx, in, 2, func(_ context.Context, v int) (int, error) {
			time.Sleep(5 * time.Millisecond)
			return v * 2, nil
		})

		if err == nil {
			t.Error("Expected context error, got nil")
		}
	})

	t.Run("zero concurrency", func(t *testing.T) {
		ctx := context.Background()
		in := []int{1, 2, 3}
		out, err := MapConcurrent(ctx, in, 0, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		expected := []int{2, 4, 6}
		for i, v := range out {
			if v != expected[i] {
				t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		ctx := context.Background()
		in := []int{}
		out, err := MapConcurrent(ctx, in, 2, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		if err != nil {
			t.Fatal(err)
		}

		if len(out) != 0 {
			t.Errorf("Expected empty output, got %d elements", len(out))
		}
	})
}

// TestLegacyPipeline tests the legacy pipeline functionality
func TestLegacyPipeline(t *testing.T) {
	t.Run("basic pipeline", func(t *testing.T) {
		ctx := context.Background()
		pipeline := NewPipeline[int](ctx)

		// Add stages: multiply by 2, then convert to string
		pipeline.AddStage(Map(func(v int) int {
			return v * 2
		})).AddStage(Filter(func(v int) bool {
			return v > 2
		}))

		input := make(chan int)
		output := pipeline.Run(input)

		// Send test data
		go func() {
			for i := 1; i <= 3; i++ {
				input <- i
			}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		// After *2: [2,4,6], after filter >2: [4,6]
		expected := []int{4, 6}
		if len(results) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(results))
		}

		for i, v := range results {
			if v != expected[i] {
				t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
			}
		}
	})
}

// Benchmark tests
func BenchmarkPool(b *testing.B) {
	ctx := context.Background()
	pool := NewPool[int, int](runtime.NumCPU(), func(_ context.Context, v int) (int, error) {
		return v * 2, nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		jobs := make(chan int, 100)
		results := pool.Run(ctx, jobs)

		go func() {
			for j := 0; j < 100; j++ {
				jobs <- j
			}
			close(jobs)
		}()

		for range results {
			// Consume results
		}
	}
}

func BenchmarkMapConcurrent(b *testing.B) {
	ctx := context.Background()
	data := make([]int, 100)
	for i := range data {
		data[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := MapConcurrent(ctx, data, runtime.NumCPU(), func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

// TestConcurrency tests for race conditions
func TestConcurrency(t *testing.T) {
	ctx := context.Background()

	// Test Pool with concurrent access
	t.Run("pool concurrency", func(t *testing.T) {
		jobs := make(chan int, 100)
		pool := NewPool[int, int](10, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})

		results := pool.Run(ctx, jobs)

		var wg sync.WaitGroup
		// Multiple goroutines sending jobs
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(start int) {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					jobs <- start*20 + j
				}
			}(i)
		}

		go func() {
			wg.Wait()
			close(jobs)
		}()

		count := 0
		for range results {
			count++
		}

		if count != 100 {
			t.Errorf("Expected 100 results, got %d", count)
		}
	})

	// Test MapConcurrent with concurrent access
	t.Run("mapconcurrent concurrency", func(t *testing.T) {
		data := make([]int, 100)
		for i := range data {
			data[i] = i
		}

		_, err := MapConcurrent(ctx, data, 10, func(_ context.Context, v int) (int, error) {
			return v * 2, nil
		})

		if err != nil {
			t.Fatal(err)
		}
	})
}

// TestMemoryLeaks tests for potential memory leaks
func TestMemoryLeaks(t *testing.T) {
	ctx := context.Background()

	// Test that channels are properly closed
	t.Run("channel cleanup", func(t *testing.T) {
		jobs := make(chan int)
		pool := NewPool[int, int](2, func(_ context.Context, v int) (int, error) {
			return v, nil
		})

		results := pool.Run(ctx, jobs)

		go func() {
			for i := 0; i < 5; i++ {
				jobs <- i
			}
			close(jobs)
		}()

		// Consume all results
		for range results {
		}

		// Give time for cleanup
		time.Sleep(10 * time.Millisecond)
	})
}
