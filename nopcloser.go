package choco

import (
	"io"
)

type nopCloser struct {
	io.ReadSeeker
}

// NopCloser returns a ReadSeekCloser with a no-op close method wrapping the provided io.ReadSeeker.
func NopCloser(rs io.ReadSeeker) io.ReadSeekCloser {
	return nopCloser{rs}
}

func (nopCloser) Close() error {
	return nil
}
