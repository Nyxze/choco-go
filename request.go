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

// Return body associated to the request
func (r *Request) Body() io.ReadSeekCloser {
	return r.body
}

// Return the underlying [http.Request]
func (r *Request) Raw() *http.Request {
	return r.req
}

// Close the body associated to this request
func (r *Request) Close() error {
	if r.body == nil {
		return nil
	}
	return r.body.Close()
}

// SetBody sets the specified ReadSeekCloser as the HTTP request body, and sets Content-Type and Content-Length accordingly.
//   - body is the request body; if nil or empty, Content-Length won't be set
//   - contentType is the value for the Content-Type header; if empty, Content-Type will be deleted
func (r *Request) SetBody(body io.ReadSeekCloser, contentType string) error {
	if body == nil {
		return NewError("body is nil")
	}
	size, err := body.Seek(0, io.SeekEnd)
	raw := r.Raw()
	if err != nil {
		return NewError("failed to defined size of body")
	}
	if size == 0 {
		body = nil
		raw.Header.Del(HeaderContentLength)
	} else {
		// Rewind to start
		_, err = body.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
		// Set rewind func
		raw.GetBody = func() (io.ReadCloser, error) {
			_, err = body.Seek(0, io.SeekStart)
			return body, err
		}
	}
	raw.Body = body
	raw.ContentLength = size

	if contentType == "" {
		r.Raw().Header.Del(HeaderContentType)
	} else {
		raw.Header.Set(HeaderContentType, contentType)
	}
	r.body = body
	return nil
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
