# 🐤 Choco-Go

A feather-light, composable middleware pipeline for taming HTTP requests in Go — inspired by our fastest, fluffiest and favorite bird, Clou.. wait, no — Chocobo, the trusty steed from the FF series.

---

## Overview

**Choco-Go** provides a flexible, testable way to construct HTTP request pipelines. Each unit of work, called a `PipelineStep`, can inspect, modify, or act on requests/responses in a clean, functional style. The pipeline ends with a `Transport` that executes the actual HTTP request.

---

## Configuring the `Pipeline`

Pipelines are configured with `PipelineOptions` — functional modifiers applied when constructing the pipeline.

### `WithCustomTransport`

Injects a custom `Transport` implementation into the pipeline. Useful for mocking, instrumentation, or altering how requests are sent.

```go
pipeline, err := NewPipeline(
    WithCustomTransport(myTransport),
)
```

If omitted, the pipeline defaults to `http.DefaultClient`.

### `WithSteps`

Adds one or more `PipelineStep`s to the request flow.

```go
pipeline, err := NewPipeline(
    WithSteps(LoggingStep(), HeaderInjector("X-App", "choco")),
)
```

Steps are executed in the order added and wrap each other like middleware.

---

## Implementing the `Transport` Interface

The `Transport` is the final rider in the relay. It’s the component responsible for executing the fully-formed `*http.Request` and returning a `*http.Response` or an error.

```go
type Transport interface {
    Send(*http.Request) (*http.Response, error)
}
```

This is the **last step** in the pipeline — the one that flaps its wings and takes off into the HTTP skies! 🐤

While you can implement your own `Transport` (e.g. for mocking or using custom protocols), the default implementation wraps a standard `http.Client`.

---

## Request Flow

Calling `pipeline.Execute(*Request)` processes the request through each step and finally the transport. The response then bubbles back through each step in reverse order.

Example setup:

```go
pipeline, err := NewPipeline(
    WithSteps(StepA, StepB, StepC),
    WithCustomTransport(TransportZ),
)
```

**Flow:**

```
Request  -> StepA -> StepB -> StepC -> TransportZ --------+
                                                         |
                                                      HTTP server
                                                         |
Response <- StepA <- StepB <- StepC <- http.Response <---+
```

Each step can:

* Enrich or rewrite the outgoing request
* Handle errors or retries
* Modify the response or inject metadata
* Collect metrics or tracing info

---

## Implementing a `PipelineStep`

A `PipelineStep` is any component that implements:

```go
type PipelineStep interface {
    Do(req *Request, next RequestHandlerFunc) (*http.Response, error)
}
```

You can implement it:

* **As a function**, using `PipelineStepFunc` for stateless behavior
* **As a struct**, for steps that require internal state

> 🔒 Note: Steps are shared across all executions of a pipeline. If you use internal state, ensure it is **thread-safe**.

---

## How Steps Work

When executing a request, the pipeline builds a handler chain where each step wraps the next. Each step can:

* Inspect or mutate the `*Request`
* Call the `next` handler (to continue execution)
* Modify or inspect the `*http.Response`
* Return early (e.g., for error injection or short-circuiting)

⚠️ If a step forgets to call `next`, and doesn't return a response or error, the pipeline will fail.

---

## Example Use Cases

**Stateless:**

* Inject headers
* Rewrite query parameters
* Log requests/responses

**Stateful:**

* Track retry counts
* Refresh access tokens
* Enforce rate limits
* Maintain a cache

---

## Example: Logging Step

```go
func LoggingStep() PipelineStep {
    return PipelineStepFunc(func(req *Request, next RequestHandlerFunc) (*http.Response, error) {
        fmt.Println("Before request")
        resp, err := next(req)
        fmt.Println("After request")
        return resp, err
    })
}
```

---

## Example: Header Injection

```go
func HeaderInjector(key, value string) PipelineStep {
    return PipelineStepFunc(func(req *Request, next RequestHandlerFunc) (*http.Response, error) {
        req.req.Header.Set(key, value)
        return next(req)
    })
}
```

---

## TODO

* [ ] Built-in steps: retry, timeout, tracing
* [ ] Context-aware execution
* [ ] Enhanced request/response mutation utilities

---