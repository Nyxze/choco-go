package blitz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Options func(*Request) error
type Request struct {
	MaxRetries  int
	Url         *url.URL
	Timeout     time.Duration
	Context     context.Context
	Cmd         Executor
	Middlewares []Middleware
	r           *http.Request
}

var DEFAULT_HEADERS map[string]string = map[string]string{
	"Content-Type": "application/json",
	"Accept":       "application/json",
}

type Executor interface {
	Do(*http.Request) (*http.Response, error)
}
type ExecutorFunc func(*http.Request) (*http.Response, error)

func (f ExecutorFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

type Middleware = func(Executor) Executor

func newError(message string, args ...any) error {
	msg := fmt.Sprintf("[BLITZ]:%s", message)
	return fmt.Errorf(msg, args)
}

func NewRequest(ctx context.Context, method, _url string, payload any, opts ...Options) (*Request, error) {
	var reader io.Reader
	// Check payload serialization implementation
	switch data := payload.(type) {
	case json.Marshaler:
		content, err := data.MarshalJSON()
		if err != nil {
			return nil, newError("serialization: failed to marshal as JSON", err)
		}
		reader = bytes.NewBuffer(content)
	case []byte:
		reader = bytes.NewBuffer(data)
	case io.Reader:
		reader = data
	}

	// Fallback as JSON anyway
	if reader == nil {
		content, err := json.Marshal(payload)
		if err != nil {
			return nil, newError("serialization: failed to marshal as JSON", err)
		}
		reader = bytes.NewBuffer(content)
	}

	req, err := http.NewRequestWithContext(ctx, method, _url, nil)
	if err != nil {
		return nil, newError("failed to create request.", err)
	}
	// Apply default headers
	for k, v := range DEFAULT_HEADERS {
		req.Header.Set(k, v)
	}
	r := Request{
		MaxRetries: 2,
		Url:        &url.URL{},
		r:          req,
		Context:    ctx,
		Cmd:        http.DefaultClient,
	}
	for _, opt := range opts {
		opt(&r)
	}
	return &r, nil
}

func (r Request) Send() error {

	// Validate
	if r.Url == nil {
		return newError("request: no url set")
	}
	if r.Cmd == nil {
		return newError("request: missing executor")
	}
	handler := r.Cmd
	// Apply middlware in reverse order
	for i := len(r.Middlewares) - 1; i >= 0; i-- {
		handler = r.Middlewares[i](handler)
	}
	_, _ = handler.Do(r.r)
	return nil
}

func WithHeader(key, value string) Options {
	return func(r *Request) error {
		r.r.Header.Set(key, value)
		return nil
	}
}
func WithMiddleware(m ...Middleware) Options {
	return func(r *Request) error {
		r.Middlewares = append(r.Middlewares, m...)
		return nil
	}
}
func WithCustomExecutor(e Executor) Options {
	return func(r *Request) error {
		if e == nil {
			return newError("executor: invalid custom executor", e)
		}
		r.Cmd = e
		return nil
	}
}
