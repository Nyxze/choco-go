package choco

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Request wraps the standard http.Request.
type Request struct {
	// Inner request
	req *http.Request

	// Content of the request
	body io.ReadSeekCloser
}

// RequestHandlerFunc defines a function that processes a *Request
// and returns an HTTP response or error.
type RequestHandlerFunc func(*Request) (*http.Response, error)

func (r Request) Body() io.ReadSeekCloser {
	return r.body
}
func (r Request) Raw() *http.Request {
	return r.req
}

func NewRequest(ctx context.Context, httpMethod string, endpoint string) (*Request, error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if req.URL.Host == "" {
		return nil, NewError("no Host in request URL")
	}
	if !(req.URL.Scheme == "http" || req.URL.Scheme == "https") {
		return nil, fmt.Errorf("unsupported protocol scheme %s", req.URL.Scheme)
	}
	return &Request{req: req}, nil
}
func NewError(message string, args ...any) error {
	msg := fmt.Sprintf("[choco]:%s", message)
	return fmt.Errorf(msg, args)
}
