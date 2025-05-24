package choco

import (
	"net/http"
)

// Pipeline defines a chain of PipelineSteps that process the request
// in sequence and ultimately produce an HTTP response.
type Pipeline struct {
	steps []PipelineStep
	tr    Transport
}

// PipelineStep represents a single unit of work in the pipeline.
// It wraps a RequestHandlerFunc and returns a new one.
type PipelineStep interface {
	Do(RequestHandlerFunc) RequestHandlerFunc
}

// PipelineStepFunc is a function adapter to allow the use of
// ordinary functions as PipelineSteps.
type PipelineStepFunc func(next RequestHandlerFunc) RequestHandlerFunc

// Do makes PipelineStepFunc satisfy the PipelineStep interface.
func (f PipelineStepFunc) Do(next RequestHandlerFunc) RequestHandlerFunc {
	return f(next)
}

// Transport is the final component responsible for executing the actual HTTP request.
type Transport interface {
	Send(*http.Request) (*http.Response, error)
}

// NewPipeline creates a new Pipeline from the given steps,
// appending a final step that wraps the provided Transport.
func NewPipeline(opts ...PipelineOptions) (Pipeline, error) {

	// Apply defaut
	pipeline := Pipeline{}
	var err error
	for i := range opts {
		err = opts[i](&pipeline)
		if err != nil {
			return Pipeline{}, NewError("pipeline: failed to apply pipeline options")
		}
	}
	return pipeline, nil
}

// Execute runs the pipeline, passing the request through all steps.
func (p Pipeline) Execute(req *Request) (*http.Response, error) {
	if req == nil {
		return nil, NewError("request is nil")
	}

	if p.tr == nil {
		return nil, NewError("pipeline transport is not set")
	}

	// Start with the transport as the terminal handler in the pipeline.
	handler := func(r *Request) (*http.Response, error) {
		return p.tr.Send(r.Raw())
	}

	// Wrap the handler with each step, from last to first.
	for i := len(p.steps) - 1; i >= 0; i-- {
		handler = p.steps[i].Do(handler)
	}

	// Execute the fully composed pipeline.
	return handler(req)

}
