package choco

// PipelineOption defines a functional option for configuring a Pipeline.
// These options are applied in order during Pipeline construction via NewPipeline.
type PipelineOption func(*Pipeline) error

// WithCustomTransport sets the transport responsible for executing the final HTTP request.
//
// Use this when you want to customize how requests are sent, for example by injecting
// a mock transport or wrapping the default http.Client.
//
// Example:
//
//	pipeline, err := NewPipeline(
//	    WithCustomTransport(myTransport),
//	    WithSteps(...),
//	)
//
// If omitted, the pipeline will default to using http.DefaultClient.
//
// Parameters:
//   - tr: A Transport implementation that defines how to send the final *http.Request.
//
// Returns:
//   - A PipelineOptions function to be used with NewPipeline.
func WithCustomTransport(tr Transport) PipelineOption {
	return func(p *Pipeline) error {
		p.transport = tr
		return nil
	}
}

// WithSteps appends one or more PipelineStep implementations to the pipeline.
//
// These steps will be executed in the order they are provided. Each step can mutate,
// inspect, or wrap the request/response lifecycle.
//
// Example:
//
//	pipeline, err := NewPipeline(
//	    WithSteps(Logging(), Retry(), Authentication()),
//	    WithCustomTransport(httpTransport),
//	)
//
// Steps are applied in the order passed, and executed in a nested fashion
// around the request/response cycle.
//
// Parameters:
//   - steps: One or more PipelineStep instances to include in the pipeline.
//
// Returns:
//   - A PipelineOptions function to be used with NewPipeline.
func WithSteps(steps ...PipelineStep) PipelineOption {
	return func(p *Pipeline) error {
		p.steps = append(p.steps, steps...)
		return nil
	}
}
