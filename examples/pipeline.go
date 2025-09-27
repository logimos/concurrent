package concurrent

import "context"

// Stage is a same-type transformation from in -> out.
// Keep stages pure(ish) and cancellation-aware.
type Stage[T any] func(ctx context.Context, in <-chan T) <-chan T

// Chain composes multiple stages into one.
func Chain[T any](stages ...Stage[T]) Stage[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		ch := in
		for _, s := range stages {
			ch = s(ctx, ch)
		}
		return ch
	}
}

// Map creates a stage that applies f to each item.
func Map[T any](f func(T) T) Stage[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					nv := f(v)
					select {
					case <-ctx.Done():
						return
					case out <- nv:
					}
				}
			}
		}()
		return out
	}
}

// Filter passes only items where pred(v) is true.
func Filter[T any](pred func(T) bool) Stage[T] {
	return func(ctx context.Context, in <-chan T) <-chan T {
		out := make(chan T)
		go func() {
			defer close(out)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					if pred(v) {
						select {
						case <-ctx.Done():
							return
						case out <- v:
						}
					}
				}
			}
		}()
		return out
	}
}

// Batch groups items into slices of size n (final batch may be smaller).
func Batch[T any](n int) Stage[T] {
	if n <= 0 {
		n = 1
	}
	return func(ctx context.Context, in <-chan T) <-chan T {
		// We expose []T via any trick? Simpler: use []T as T by letting caller define T as []Elem.
		// To keep type-safe without reflection, we provide a dedicated BatchSlice below.
		panic("use BatchSlice for batching; define T as []Elem")
	}
}

// BatchSlice groups items into slices of size n.
// Usage: start with a channel of Elem, then adapt with ToSlice to get chan []Elem stages.
func BatchSlice[T any](n int) Stage[[]T] {
	if n <= 0 {
		n = 1
	}
	return func(ctx context.Context, in <-chan []T) <-chan []T {
		// This stage expects upstream to send single-element slices.
		out := make(chan []T)
		go func() {
			defer close(out)
			buf := make([]T, 0, n)
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						if len(buf) > 0 {
							select {
							case <-ctx.Done():
								return
							case out <- append([]T(nil), buf...):
							}
						}
						return
					}
					if len(v) == 1 {
						buf = append(buf, v[0])
					} else if len(v) > 1 {
						// allow passing pre-chunked slices too
						buf = append(buf, v...)
					}
					if len(buf) >= n {
						chunk := append([]T(nil), buf[:n]...)
						buf = append([]T(nil), buf[n:]...)
						select {
						case <-ctx.Done():
							return
						case out <- chunk:
						}
					}
				}
			}
		}()
		return out
	}
}

// ToSlice lifts a chan T into a chan []T by wrapping each element.
func ToSlice[T any]() Stage[[]T] {
	return func(ctx context.Context, in <-chan []T) <-chan []T {
		// This adapter is intended to be used by bridging: see Lift below.
		return in
	}
}

// Lift converts chan T -> chan []T by wrapping values (helper for batching pipelines).
func Lift[T any](ctx context.Context, in <-chan T) <-chan []T {
	out := make(chan []T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-ctx.Done():
					return
				case out <- []T{v}:
				}
			}
		}
	}()
	return out
}

// Unlift converts chan []T -> chan T (flatten).
func Unlift[T any](ctx context.Context, in <-chan []T) <-chan T {
	out := make(chan T)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case vs, ok := <-in:
				if !ok {
					return
				}
				for _, v := range vs {
					select {
					case <-ctx.Done():
						return
					case out <- v:
					}
				}
			}
		}
	}()
	return out
}
