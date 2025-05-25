package choco

import (
	"net/http"
)

// [Pipeline] defines a chain of [PipelineStep]s that process a [Request]
// in sequence and ultimately produce an [http.Response].
type Pipeline struct {
	steps     []PipelineStep
	transport Transport
}

// [PipelineStep] represents a single unit of work in a [Pipeline].
// Each step can either handle the [Request] directly or delegate to the [next] [RequestHandlerFunc].
//
// It must return either a non-nil [http.Response] or an [error].
// Failing to do so (e.g. by forgetting to call [next] and returning nil values)
// will cause pipeline execution to fail.
type PipelineStep interface {
	Do(req *Request, next RequestHandlerFunc) (*http.Response, error)
}

// [PipelineStepFunc] is a function adapter that allows a plain function to satisfy
// the [PipelineStep] interface.
type PipelineStepFunc func(req *Request, next RequestHandlerFunc) (*http.Response, error)

// Do makes [PipelineStepFunc] implement the [PipelineStep] interface.
func (f PipelineStepFunc) Do(req *Request, next RequestHandlerFunc) (*http.Response, error) {
	return f(req, next)
}

// [Transport] is the final component in the [Pipeline] responsible for
// executing the actual [http.Request] and returning an [http.Response].
type Transport interface {
	Send(*http.Request) (*http.Response, error)
}

// NewPipeline creates a new [Pipeline] by applying the provided [PipelineOptions].
// Default options are applied first, followed by overrides from [opts].
// Returns an error if any [PipelineOption] fail to apply.
func NewPipeline(opts ...PipelineOption) (Pipeline, error) {
	pipeline := Pipeline{
		transport: defaultTransport{
			client: http.DefaultClient,
		},
	}
	var err error
	for i := range opts {
		err = opts[i](&pipeline)
		if err != nil {
			return Pipeline{}, NewError("pipeline: failed to apply pipeline options")
		}
	}
	return pipeline, nil
}

// Execute runs the [Pipeline], passing the [Request] through all registered [PipelineStep]s.
// Each step may inspect, modify, short-circuit, or pass the request to the next step.
// If no step calls [next], the pipeline will not proceed and an error will be returned.
func (p Pipeline) Execute(req *Request) (*http.Response, error) {
	if req == nil {
		return nil, NewError("request is nil")
	}

	if p.transport == nil {
		return nil, NewError("pipeline transport is not set")
	}

	handler := func(r *Request) (*http.Response, error) {
		return p.transport.Send(r.Raw())
	}

	for i := len(p.steps) - 1; i >= 0; i-- {
		handler = wrapStep(p.steps[i], handler)
	}

	return handler(req)
}

// wrapStep composes a [PipelineStep] around a [RequestHandlerFunc].
// If the step fails to return a valid response or error (e.g. does not call [next]),
// the pipeline will fail with an appropriate error.
func wrapStep(step PipelineStep, next RequestHandlerFunc) RequestHandlerFunc {
	return func(r *Request) (*http.Response, error) {
		resp, err := step.Do(r, next)
		if resp == nil && err == nil {
			return nil, NewError("pipeline step did not call next and did not return a response or error")
		}
		return resp, err
	}
}

type defaultTransport struct {
	client *http.Client
}

func (t defaultTransport) Send(req *http.Request) (*http.Response, error) {
	return t.client.Do(req)
}
