package choco

import (
	"context"
	"io"
	"net/http"
	"net/http/httputil"
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

// Create a new request with a given method & endpoint (e.g: GET /api/v1/users)
func NewRequest(ctx context.Context, httpMethod string, endpoint string) (*Request, error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, endpoint, nil)
	if err != nil {
		return nil, err
	}
	return &Request{req: req}, nil
}

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
		r.DelHeader(HeaderContentLength)
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
		r.DelHeader(HeaderContentType)
	} else {
		r.SetContentType(contentType)
	}
	r.body = body
	return nil
}

func (r *Request) DumpRequest(body bool) ([]byte, error) {
	if r.req == nil {
		return nil, NewError("request: missing inner *http.Request")
	}
	if r.req.GetBody != nil {
		r.req.Body, _ = r.req.GetBody()
	}
	return httputil.DumpRequestOut(r.req, body)
}
