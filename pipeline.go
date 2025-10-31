package concurrent

import (
	"context"
	"sync"
)

// Stage is a transformation function from in -> out channel.
type Stage[T any, R any] func(context.Context, <-chan T) <-chan R

// Pipeline represents a data processing pipeline.
type Pipeline[T any] struct {
	stages []Stage[T, T]
	ctx    context.Context
	cancel context.CancelFunc
}

// NewPipeline creates a new pipeline.
func NewPipeline[T any](ctx context.Context) *Pipeline[T] {
	ctx, cancel := context.WithCancel(ctx)
	return &Pipeline[T]{
		stages: make([]Stage[T, T], 0),
		ctx:    ctx,
		cancel: cancel,
	}
}

// AddStage adds a stage to the pipeline.
func (p *Pipeline[T]) AddStage(stage Stage[T, T]) *Pipeline[T] {
	p.stages = append(p.stages, stage)
	return p
}

// Run executes the pipeline with the given input channel.
func (p *Pipeline[T]) Run(input <-chan T) <-chan T {
	if len(p.stages) == 0 {
		// No stages, just pass through
		output := make(chan T)
		go func() {
			defer close(output)
			for {
				select {
				case <-p.ctx.Done():
					return
				case item, ok := <-input:
					if !ok {
						return
					}
					select {
					case <-p.ctx.Done():
						return
					case output <- item:
					}
				}
			}
		}()
		return output
	}

	// Chain stages together
	ch := input
	for _, stage := range p.stages {
		ch = stage(p.ctx, ch)
	}
	return ch
}

// Close cancels the pipeline context.
func (p *Pipeline[T]) Close() {
	p.cancel()
}

// PipelineBuilder provides a fluent interface for building pipelines.
type PipelineBuilder[T any] struct {
	pipeline *Pipeline[T]
}

// NewPipelineBuilder creates a new pipeline builder.
func NewPipelineBuilder[T any](ctx context.Context) *PipelineBuilder[T] {
	return &PipelineBuilder[T]{
		pipeline: NewPipeline[T](ctx),
	}
}

// AddStage adds a stage to the pipeline.
func (pb *PipelineBuilder[T]) AddStage(stage Stage[T, T]) *PipelineBuilder[T] {
	pb.pipeline.AddStage(stage)
	return pb
}

// Build returns the completed pipeline.
func (pb *PipelineBuilder[T]) Build() *Pipeline[T] {
	return pb.pipeline
}

// Map creates a stage that applies a function to each item.
func Map[T any](fn func(T) T) Stage[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-input:
					if !ok {
						return
					}
					result := fn(item)
					select {
					case <-ctx.Done():
						return
					case output <- result:
					}
				}
			}
		}()
		return output
	}
}

// Filter creates a stage that filters items based on a predicate.
func Filter[T any](predicate func(T) bool) Stage[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-input:
					if !ok {
						return
					}
					if predicate(item) {
						select {
						case <-ctx.Done():
							return
						case output <- item:
						}
					}
				}
			}
		}()
		return output
	}
}

// Batch creates a stage that batches items into slices.
func Batch[T any](size int) Stage[T, []T] {
	if size <= 0 {
		size = 1
	}
	return func(ctx context.Context, input <-chan T) <-chan []T {
		output := make(chan []T)
		go func() {
			defer close(output)
			batch := make([]T, 0, size)
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-input:
					if !ok {
						// Send final batch if it has items
						if len(batch) > 0 {
							select {
							case <-ctx.Done():
								return
							case output <- append([]T(nil), batch...):
							}
						}
						return
					}
					batch = append(batch, item)
					if len(batch) >= size {
						select {
						case <-ctx.Done():
							return
						case output <- append([]T(nil), batch...):
						}
						batch = batch[:0] // Reset batch
					}
				}
			}
		}()
		return output
	}
}

// Unbatch creates a stage that unbatch slices into individual items.
func Unbatch[T any]() Stage[[]T, T] {
	return func(ctx context.Context, input <-chan []T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			for {
				select {
				case <-ctx.Done():
					return
				case batch, ok := <-input:
					if !ok {
						return
					}
					for _, item := range batch {
						select {
						case <-ctx.Done():
							return
						case output <- item:
						}
					}
				}
			}
		}()
		return output
	}
}

// Tee creates a stage that splits the input into multiple outputs.
// Note: Tee closes the provided output channels when the input channel closes.
// Do not reuse these channels after passing them to Tee.
func Tee[T any](outputs ...chan<- T) Stage[T, T] {
	return func(ctx context.Context, input <-chan T) <-chan T {
		output := make(chan T)
		go func() {
			defer close(output)
			// Close all output channels when done
			defer func() {
				for _, out := range outputs {
					close(out)
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
					// Send to all outputs concurrently
					var wg sync.WaitGroup
					for _, out := range outputs {
						wg.Add(1)
						go func(ch chan<- T) {
							defer wg.Done()
							select {
							case <-ctx.Done():
								return
							case ch <- item:
							}
						}(out)
					}

					// Also send to main output
					select {
					case <-ctx.Done():
						return
					case output <- item:
					}

					// Wait for all outputs to complete
					wg.Wait()
				}
			}
		}()
		return output
	}
}

// Merge creates a stage that merges multiple inputs into one output.
// The output channel is closed when all input channels are closed or context is cancelled.
func Merge[T any](ctx context.Context, inputs ...<-chan T) <-chan T {
	output := make(chan T)
	var wg sync.WaitGroup

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

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}
