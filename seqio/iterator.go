package seqio

import (
	"bufio"
	"io"
)

// Iterator is a generic interface for forward-only iteration over values of type T.
// It is typically used to consume streamed or decoded data one item at a time.
type Iterator[T any] interface {
	// Next advances the iterator to the next value.
	// It returns false when there are no more values or when an error occurs.
	// After Next returns false, Err should be called to check whether the
	// iteration stopped due to an error or normal completion.
	Next() bool

	// Curr returns the current value.
	// It must only be called after a successful call to Next.
	Curr() T

	// Err returns the error that caused the iteration to stop, if any.
	// It should be called after Next returns false to determine whether the
	// iterator completed normally or due to an error.
	Err() error
}

func NewLineIter(r io.ReadCloser) Iterator[string] {
	s := bufio.NewScanner(r)
	return &lineIter{
		scanner: s,
	}
}

// lineIter is an Iterator implementation that yields lines of text
// from an underlying bufio.Scanner, one line at a time.
type lineIter struct {
	scanner *bufio.Scanner
}

// Next advances the scanner to the next line.
// It returns false when there are no more lines or an error occurs.
func (l *lineIter) Next() bool {
	return l.scanner.Scan()
}

// Curr returns the current line read by the scanner.
// It must only be called after a successful call to Next.
func (l *lineIter) Curr() string {
	return l.scanner.Text()
}

// Err returns the error encountered during scanning, if any.
// It should be called after Next returns false.
func (l *lineIter) Err() error {
	return l.scanner.Err()
}
