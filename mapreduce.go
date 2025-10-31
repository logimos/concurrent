package concurrent

import (
	"context"
	"sync"
)

// MapConcurrent applies fn to each element with at most n concurrent tasks.
// Returns the outputs in the original order. The first error aborts early.
// If ctx is cancelled, it waits for in-flight operations to complete before returning.
func MapConcurrent[T any, R any](ctx context.Context, in []T, n int, fn func(context.Context, T) (R, error)) ([]R, error) {
	if n <= 0 {
		n = 1
	}
	if len(in) == 0 {
		return []R{}, nil
	}

	sem := make(chan struct{}, n)
	var wg sync.WaitGroup
	var mu sync.Mutex

	out := make([]R, len(in))
	errs := make(chan error, 1)
	cancelled := false

	for i, v := range in {
		// Check cancellation before starting new goroutine
		select {
		case <-ctx.Done():
			mu.Lock()
			cancelled = true
			mu.Unlock()
			goto waitComplete
		default:
		}

		mu.Lock()
		if cancelled {
			mu.Unlock()
			goto waitComplete
		}
		mu.Unlock()

		wg.Add(1)
		sem <- struct{}{}
		i, v := i, v
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			// Check context cancellation before processing
			select {
			case <-ctx.Done():
				return
			default:
			}

			r, err := fn(ctx, v)
			if err != nil {
				select {
				case errs <- err:
				default:
				}
				return
			}

			// Write result atomically
			mu.Lock()
			out[i] = r
			mu.Unlock()
		}()
	}

waitComplete:

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		// Wait for in-flight operations to complete
		<-done
		return nil, ctx.Err()
	case err := <-errs:
		if err != nil {
			// Wait for in-flight operations to complete
			<-done
			return nil, err
		}
		<-done
		return out, nil
	case <-done:
		return out, nil
	}
}
