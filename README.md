# 🐤 Choco-Go

A feather-light, composable middleware pipeline for taming HTTP requests in Go — inspired by our fastest, fluffiest and favorite bird, Clou.. wait, no — Chocobo, the trusty steed from the FF series.

---

##  Overview

**Choco-Go** lets you build a flexible and testable HTTP request pipeline in Go. Each unit of work, called a `PipelineStep`, can inspect, modify, or act on the request/response in a clean, functional style. The pipeline ends with a `Transport` that performs the actual HTTP request.

---

##  Configuring the Pipeline

You configure a pipeline using a set of `PipelineOptions`, which are functional modifiers applied when constructing the pipeline.

### `WithCustomTransport`

Injects a custom `Transport` implementation into the pipeline. Useful for mocking, instrumentation, or altering the way requests are sent.

```go
pipeline, err := NewPipeline(
	WithCustomTransport(myTransport),
)
```

If omitted, the pipeline defaults to using `http.DefaultClient`.

### `WithSteps`

Adds one or more `PipelineStep` instances to the request flow.

```go
pipeline, err := NewPipeline(
	WithSteps(LoggingStep(), HeaderInjector("X-App", "choco")),
)
```

Steps are executed in the order they are added, and wrap each other like middleware.

---

##  Implementing the `Transport` Interface

The `Transport` interface is the final rider in the request relay. It’s the component actually responsible for taking a fully-prepared `*http.Request`, sending it over the network, and returning the resulting `*http.Response` or an error.

```go
type Transport interface {
	Send(*http.Request) (*http.Response, error)
}
```

This is the **last step** in the pipeline — the one that flaps its wings and takes off into the HTTP skies! 🐤

While you can implement your own `Transport` (perhaps to mock network calls or use custom protocols), the typical implementation wraps a standard `http.Client`.

---

##  Request Flow

When you call `Pipeline.Execute(*Request)`, the request passes through each step in sequence until it reaches the transport. Then the response bubbles back up the chain, giving each step a chance to inspect or mutate it.

Here’s how it looks with `StepA`, `StepB`, `StepC`, and `TransportZ`:

```go
pipeline := NewPipeline(
	WithSteps(StepA, StepB, StepC),
	WithCustomTransport(TransportZ),
)
```

**Request flow:**

```
Request  -> StepA -> StepB -> StepC -> TransportZ --------+
                                                         |
                                                      HTTP server
                                                         |
Response <- StepA <- StepB <- StepC <- http.Response <---+
```

Each step can:

* Modify or enrich the outgoing request
* Log, retry, or mutate the incoming response
* Inject errors, headers, context, metrics, or just plain chaos (responsibly, please)

---

##  Implementing a `PipelineStep`

A `PipelineStep` can be implemented in two ways:

* **As a function** using `PipelineStepFunc` for *stateless* behavior.
* **As a struct** with a `Do(next)` method for *stateful* behavior.

Note: all requests sent through the same `Pipeline` share the same `PipelineStep` instances. If a step mutates internal state, it **must be thread-safe** to avoid data races in concurrent environments.

---

##  How Steps Work

When a request is executed, the pipeline builds a chain of handlers by wrapping a `Transport` in multiple `PipelineStep`s. Each step has full control over:

* The incoming `*Request`
* The execution of the next step
* The final `*http.Response` and `error` returned to the caller

Each step can:

* Inspect or mutate the request
* Call the `next` step
* Inspect or modify the response
* Handle errors

---

##  Example Use Cases

* **Stateless:**

  * Inject headers
  * Rewrite query parameters
  * Log metadata

* **Stateful:**

  * Track retry counts
  * Manage authentication tokens
  * Cache results
  * Enforce rate limits

---

##  Example: Logging Step

```go
func LoggingStep() PipelineStep {
	return PipelineStepFunc(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(r *Request) (*http.Response, error) {
			fmt.Println("Before request")
			resp, err := next(r)
			fmt.Println("After request")
			return resp, err
		}
	})
}
```

---

##  Example: Header Injection

```go
func HeaderInjector(key, value string) PipelineStep {
	return PipelineStepFunc(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(r *Request) (*http.Response, error) {
			r.req.Header.Set(key, value)
			return next(r)
		}
	})
}
```
---

##  TODO

* Built-in steps: retry, timeout, tracing
* Context-aware execution
* Enhanced request/response mutation helpers
