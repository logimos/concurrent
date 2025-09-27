package concurrent

import (
	"context"
	"sync"
)

// MapConcurrent applies fn to each element with at most n concurrent tasks.
// Returns the outputs in the original order. The first error aborts early.
func MapConcurrent[T any, R any](ctx context.Context, in []T, n int, fn func(context.Context, T) (R, error)) ([]R, error) {
	if n <= 0 {
		n = 1
	}
	sem := make(chan struct{}, n)
	var wg sync.WaitGroup

	out := make([]R, len(in))
	errs := make(chan error, 1)

	for i, v := range in {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		wg.Add(1)
		sem <- struct{}{}
		i, v := i, v
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			r, err := fn(ctx, v)
			if err != nil {
				select {
				case errs <- err:
				default:
				}
				return
			}
			out[i] = r
		}()
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errs:
		if err != nil {
			return nil, err
		}
		<-done
		return out, nil
	case <-done:
		return out, nil
	}
}
