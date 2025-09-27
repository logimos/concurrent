package concurrent

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewPipeline(t *testing.T) {
	t.Run("basic pipeline", func(t *testing.T) {
		ctx := context.Background()
		pipeline := NewPipeline[int](ctx)

		// Add stages: multiply by 2, then filter even numbers
		pipeline.AddStage(Map(func(v int) int {
			return v * 2
		})).AddStage(Filter(func(v int) bool {
			return v > 5
		}))

		input := make(chan int)
		output := pipeline.Run(input)

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

		// After *2: [2,4,6,8,10], after filter >5: [6,8,10]
		expected := []int{6, 8, 10}
		if len(results) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(results))
		}

		for i, v := range results {
			if v != expected[i] {
				t.Errorf("Expected %d at index %d, got %d", expected[i], i, v)
			}
		}
	})

	t.Run("empty pipeline", func(t *testing.T) {
		ctx := context.Background()
		pipeline := NewPipeline[int](ctx)

		input := make(chan int)
		output := pipeline.Run(input)

		go func() {
			input <- 1
			input <- 2
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		// Should pass through unchanged
		expected := []int{1, 2}
		if len(results) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(results))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		pipeline := NewPipeline[int](ctx)
		pipeline.AddStage(Map(func(v int) int {
			time.Sleep(100 * time.Millisecond) // Longer than timeout
			return v * 2
		}))

		input := make(chan int)
		output := pipeline.Run(input)

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

	t.Run("pipeline close", func(t *testing.T) {
		ctx := context.Background()
		pipeline := NewPipeline[int](ctx)

		input := make(chan int)
		output := pipeline.Run(input)

		go func() {
			for i := 0; i < 5; i++ {
				input <- i
			}
			close(input)
		}()

		// Close pipeline after a short delay
		go func() {
			time.Sleep(10 * time.Millisecond)
			pipeline.Close()
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		t.Logf("Got %d results before close", len(results))
	})
}

func TestPipelineBuilder(t *testing.T) {
	t.Run("fluent interface", func(t *testing.T) {
		ctx := context.Background()
		pipeline := NewPipelineBuilder[int](ctx).
			AddStage(Map(func(v int) int {
				return v * 2
			})).
			AddStage(Filter(func(v int) bool {
				return v > 5
			})).
			Build()

		input := make(chan int)
		output := pipeline.Run(input)

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

		expected := []int{6, 8, 10}
		if len(results) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(results))
		}
	})
}

func TestMap(t *testing.T) {
	t.Run("basic mapping", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		stage := Map(func(v int) int {
			return v * 2
		})

		output := stage(ctx, input)

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

		expected := []int{2, 4, 6}
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

func TestFilter(t *testing.T) {
	t.Run("basic filtering", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		stage := Filter(func(v int) bool {
			return v%2 == 0
		})

		output := stage(ctx, input)

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

		expected := []int{2, 4}
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

func TestBatch(t *testing.T) {
	t.Run("basic batching", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		stage := Batch[int](3)
		output := stage(ctx, input)

		go func() {
			for i := 1; i <= 7; i++ {
				input <- i
			}
			close(input)
		}()

		var results [][]int
		for v := range output {
			results = append(results, v)
		}

		// Should have 3 batches: [1,2,3], [4,5,6], [7]
		if len(results) != 3 {
			t.Errorf("Expected 3 batches, got %d", len(results))
		}

		if len(results[0]) != 3 || len(results[1]) != 3 || len(results[2]) != 1 {
			t.Errorf("Unexpected batch sizes: %v", results)
		}
	})

	t.Run("zero batch size", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		stage := Batch[int](0)
		output := stage(ctx, input)

		go func() {
			input <- 1
			close(input)
		}()

		var results [][]int
		for v := range output {
			results = append(results, v)
		}

		// Should have 1 batch with 1 item
		if len(results) != 1 {
			t.Errorf("Expected 1 batch, got %d", len(results))
		}
	})
}

func TestUnbatch(t *testing.T) {
	t.Run("basic unbatching", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan []int)

		stage := Unbatch[int]()
		output := stage(ctx, input)

		go func() {
			input <- []int{1, 2, 3}
			input <- []int{4, 5}
			close(input)
		}()

		var results []int
		for v := range output {
			results = append(results, v)
		}

		expected := []int{1, 2, 3, 4, 5}
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

func TestTee(t *testing.T) {
	t.Run("basic tee", func(t *testing.T) {
		ctx := context.Background()
		input := make(chan int)

		output1 := make(chan int, 10)
		output2 := make(chan int, 10)

		stage := Tee(output1, output2)
		output := stage(ctx, input)

		// Start goroutines to consume tee outputs
		var tee1, tee2 []int
		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			defer wg.Done()
			for v := range output1 {
				tee1 = append(tee1, v)
			}
		}()

		go func() {
			defer wg.Done()
			for v := range output2 {
				tee2 = append(tee2, v)
			}
		}()

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

		// Wait for tee outputs to complete
		wg.Wait()

		// Main output should have all items
		expected := []int{1, 2, 3}
		if len(results) != len(expected) {
			t.Errorf("Expected %d results, got %d", len(expected), len(results))
		}

		if len(tee1) != 3 || len(tee2) != 3 {
			t.Errorf("Expected tee outputs to have 3 items each, got %d and %d", len(tee1), len(tee2))
		}
	})
}

func TestMerge(t *testing.T) {
	t.Run("basic merge", func(t *testing.T) {
		input1 := make(chan int)
		input2 := make(chan int)
		input3 := make(chan int)

		output := Merge(input1, input2, input3)

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
		output := Merge[int]()

		// Should close immediately
		select {
		case <-output:
			t.Error("Expected output to be closed")
		default:
			// Good, output is closed
		}
	})
}

func BenchmarkPipeline(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pipeline := NewPipeline[int](ctx)
		pipeline.AddStage(Map(func(v int) int {
			return v * 2
		})).AddStage(Filter(func(v int) bool {
			return v > 5
		}))

		input := make(chan int, 100)
		output := pipeline.Run(input)

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

func BenchmarkMap(b *testing.B) {
	ctx := context.Background()
	stage := Map(func(v int) int {
		return v * 2
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := make(chan int, 100)
		output := stage(ctx, input)

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

func BenchmarkFilter(b *testing.B) {
	ctx := context.Background()
	stage := Filter(func(v int) bool {
		return v%2 == 0
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := make(chan int, 100)
		output := stage(ctx, input)

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
