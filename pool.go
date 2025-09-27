package concurrent

import (
	"context"
	"sync"
)

// Pool runs jobs with a fixed number of workers.
// If fn returns an error, that job's result is simply dropped.
// Use a wrapper fn if you need to propagate per-item errors.
type Pool[T any, R any] struct {
	workers int
	fn      func(context.Context, T) (R, error)
}

// NewPool creates a pool with n workers and a processing function.
func NewPool[T any, R any](n int, fn func(context.Context, T) (R, error)) *Pool[T, R] {
	if n <= 0 {
		n = 1
	}
	return &Pool[T, R]{workers: n, fn: fn}
}

// Run executes jobs until ctx is canceled or jobs is closed.
// The caller MUST consume the results channel until it is closed.
func (p *Pool[T, R]) Run(ctx context.Context, jobs <-chan T) <-chan R {
	results := make(chan R)

	var wg sync.WaitGroup
	wg.Add(p.workers)

	for i := 0; i < p.workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}
					// compute outside select to avoid blocking ctx.Done path
					r, err := p.fn(ctx, j)
					if err != nil {
						continue
					}
					select {
					case <-ctx.Done():
						return
					case results <- r:
					}
				}
			}
		}()
	}

	// Closer
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}
