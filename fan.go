package concurrent

import (
	"context"
	"sync"
)

// FanOut distributes work from a single input channel to multiple worker channels.
// Each worker processes items concurrently and sends results to a single output channel.
func FanOut[T any, R any](ctx context.Context, input <-chan T, workers int, fn func(context.Context, T) (R, error)) <-chan R {
	if workers <= 0 {
		workers = 1
	}

	output := make(chan R)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-input:
					if !ok {
						return
					}
					result, err := fn(ctx, item)
					if err != nil {
						// Log error or handle as needed
						continue
					}
					select {
					case <-ctx.Done():
						return
					case output <- result:
					}
				}
			}
		}()
	}

	// Close output when all workers are done
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// FanIn merges multiple input channels into a single output channel.
// The output channel is closed when all input channels are closed.
func FanIn[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	output := make(chan T)
	var wg sync.WaitGroup

	// Start a goroutine for each input channel
	for _, input := range inputs {
		wg.Add(1)
		go func(ch <-chan T) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-ch:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case output <- item:
					}
				}
			}
		}(input)
	}

	// Close output when all input channels are done
	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

// FanOutFanIn combines fan-out and fan-in patterns for parallel processing.
// It distributes work to multiple workers and then merges the results.
func FanOutFanIn[T any, R any](ctx context.Context, input <-chan T, workers int, fn func(context.Context, T) (R, error)) <-chan R {
	// Create intermediate channels for each worker
	workerChannels := make([]<-chan R, workers)

	// Distribute work to workers
	for i := 0; i < workers; i++ {
		workerInput := make(chan T)
		workerOutput := FanOut(ctx, workerInput, 1, fn)
		workerChannels[i] = workerOutput

		// Start distributor goroutine for this worker
		go func(ch chan T) {
			defer close(ch)
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-input:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case ch <- item:
					}
				}
			}
		}(workerInput)
	}

	// Merge all worker outputs
	return FanIn(ctx, workerChannels...)
}

// RoundRobin distributes work in round-robin fashion to multiple workers.
func RoundRobin[T any, R any](ctx context.Context, input <-chan T, workers int, fn func(context.Context, T) (R, error)) <-chan R {
	if workers <= 0 {
		workers = 1
	}

	workerChannels := make([]chan T, workers)
	workerOutputs := make([]<-chan R, workers)

	// Create worker channels and start workers
	for i := 0; i < workers; i++ {
		workerChannels[i] = make(chan T)
		workerOutputs[i] = FanOut(ctx, workerChannels[i], 1, fn)
	}

	// Start distributor
	go func() {
		defer func() {
			for _, ch := range workerChannels {
				close(ch)
			}
		}()

		workerIndex := 0
		for {
			select {
			case <-ctx.Done():
				return
			case item, ok := <-input:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case workerChannels[workerIndex] <- item:
					workerIndex = (workerIndex + 1) % workers
				}
			}
		}
	}()

	// Merge all worker outputs
	return FanIn(ctx, workerOutputs...)
}
