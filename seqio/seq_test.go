package seqio

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestReadLineSeq(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLen  int
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name:    "empty input",
			input:   "",
			wantLen: 0,
		},
		{
			name:    "single line",
			input:   "hello world",
			wantLen: 1,
		},
		{
			name:    "multiple lines",
			input:   "hello\nworld\n!",
			wantLen: 3,
		},
		{
			name:    "lines with different endings",
			input:   "hello\r\nworld\n!",
			wantLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := io.NopCloser(strings.NewReader(tt.input))
			var lines []string
			var lastErr error
			ctx := context.Background()
			for line := range Lines(ctx, r) {
				lines = append(lines, line)
			}

			if (lastErr != nil) != tt.wantErr {
				t.Errorf("ReadLineSeq() error = %v, wantErr %v", lastErr, tt.wantErr)
			}

			if tt.errCheck != nil && lastErr != nil && !tt.errCheck(lastErr) {
				t.Errorf("ReadLineSeq() error = %v, did not match error check", lastErr)
			}

			if len(lines) != tt.wantLen {
				t.Errorf("ReadLineSeq() got %d lines, want %d", len(lines), tt.wantLen)
			}
		})
	}
}

func TestReadLineSeqString(t *testing.T) {
	input := "hello\nworld\n!"
	r := io.NopCloser(strings.NewReader(input))
	var lines []string
	ctx := context.Background()
	for line := range Lines(ctx, r) {
		lines = append(lines, line)
	}

	want := []string{"hello", "world", "!"}
	if len(lines) != len(want) {
		t.Errorf("ReadLineSeqString() got %d lines, want %d", len(lines), len(want))
	}

	for i := range lines {
		if lines[i] != want[i] {
			t.Errorf("ReadLineSeqString() line %d = %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestReadLineSeqWithContext(t *testing.T) {
	// Create a large input that will take some time to process
	var longInput strings.Builder
	for i := 0; i < 1000; i++ {
		longInput.WriteString("line\n")
	}

	tests := []struct {
		name        string
		ctx         context.Context
		input       io.ReadCloser
		wantTimeout bool
	}{
		{
			name:        "context timeout",
			ctx:         timeoutContext(t),
			input:       io.NopCloser(strings.NewReader(longInput.String())),
			wantTimeout: true,
		},
		{
			name:        "context background",
			ctx:         context.Background(),
			input:       io.NopCloser(strings.NewReader("hello\nworld")),
			wantTimeout: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the test-specific context

			for _ = range Lines(tt.ctx, tt.input) {
				// simulate some processing delay if needed
			}

			gotTimeout := tt.ctx.Err() != nil && errors.Is(tt.ctx.Err(), context.DeadlineExceeded)

			if gotTimeout != tt.wantTimeout {
				t.Errorf("Line() timeout = %v, want %v", gotTimeout, tt.wantTimeout)
			}
		})
	}
}

func timeoutContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	t.Cleanup(cancel)
	return ctx
}

// mockIter is a test implementation of Decoder
type mockIter[T any] struct {
	items []T
	pos   int
	err   error
}

func newMockDecoder[T any](items []T) *mockIter[T] {
	return &mockIter[T]{
		items: items,
		pos:   0,
	}
}

func (d *mockIter[T]) Next() bool {
	if d.err != nil {
		return false
	}
	d.pos++
	return d.pos <= len(d.items)
}

func (d *mockIter[T]) Curr() T {
	if d.pos == 0 || d.pos > len(d.items) {
		var zero T
		return zero
	}
	return d.items[d.pos-1]
}

func (d *mockIter[T]) Err() error {
	return d.err
}

// errorIter is a test implementation that always returns an error
type errorIter[T any] struct {
	err error
}

func (d *errorIter[T]) Next() bool {
	return false
}

func (d *errorIter[T]) Curr() T {
	var zero T
	return zero
}

func (d *errorIter[T]) Err() error {
	return d.err
}
func TestReadSeq(t *testing.T) {
	t.Run("reads all items from decoder", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		decoder := newMockDecoder(items)
		ctx := context.Background()

		var result []int
		for item := range Range(ctx, decoder) {
			result = append(result, item)
		}

		if len(result) != len(items) {
			t.Errorf("got %d items, want %d", len(result), len(items))
		}

		for i, v := range items {
			if result[i] != v {
				t.Errorf("item %d: got %d, want %d", i, result[i], v)
			}
		}
	})

	t.Run("handles decoder error", func(t *testing.T) {
		expectedErr := errors.New("decode error")
		decoder := &errorIter[int]{err: expectedErr}
		ctx := context.Background()

		var count int
		for range Range(ctx, decoder) {
			count++
		}

		if count != 0 {
			t.Errorf("expected no items, got %d", count)
		}
	})

	t.Run("handles context cancellation", func(t *testing.T) {
		items := []int{1, 2, 3, 4, 5}
		decoder := newMockDecoder(items)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		var count int
		for range Range(ctx, decoder) {
			count++
		}

		if count != 0 {
			t.Errorf("expected no items after context cancellation, got %d", count)
		}
	})
}
