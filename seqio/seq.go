package seqio

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"iter"
)

func Lines(ctx context.Context, r io.ReadCloser) iter.Seq[string] {
	s := bufio.NewScanner(r)
	return Range(ctx, &lineIter{scanner: s})
}

// Range creates an iterator sequence from an [Iterator]. It yields values of type T
// until either the [Iterator] is exhausted, an error occurs, or the context is cancelled.
// The decoder's Err() method is checked after iteration completes to catch any errors
// that occurred during decoding.
func Range[T any](ctx context.Context, i Iterator[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if !i.Next() {
					if err := i.Err(); err != nil {
						fmt.Printf("Range: iteration error: %v", err)
						return
					}
					return
				}
				if !yield(i.Curr()) {
					return
				}
			}
		}
	}
}

func Select[T any, V any](seq iter.Seq[T], apply func(T) V) iter.Seq[V] {
	return func(yield func(V) bool) {
		for item := range seq {
			value := apply(item)
			if !yield(value) {
				return
			}
		}
	}
}
