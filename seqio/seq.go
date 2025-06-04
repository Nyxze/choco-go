package seqio

import (
	"context"
	"fmt"
	"io"
	"iter"
)

// Lines creates an iterator sequence that yields lines from an [io.ReadCloser].
// It uses a [bufio.Scanner] to read lines from the reader.
// The iteration stops when the reader is exhausted, an error occurs, or the context is cancelled.
func Lines(ctx context.Context, r io.ReadCloser) iter.Seq[string] {
	return Range(ctx, NewLineIter(r))
}

// Range creates an iterator sequence from an [Iterator]. It yields values of type T
// until either the [Iterator] is exhausted, an error occurs, or the context is cancelled.
// The iterator's Err() method is checked after iteration completes to catch any errors
// that occurred during iteration.
func Range[T any](ctx context.Context, i Iterator[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if !i.Next() {
					if err := i.Err(); err != nil && err != io.EOF {
						fmt.Printf("Range: iteration error: %v\n", err)
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
