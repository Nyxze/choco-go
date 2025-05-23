# ChocoGo

Choco-go is a Go library that implements an HTTP request/response middleware pipeline. It is designed to extend the Go standard library's `net/http` functionality by providing a flexible, composable policy pipeline for HTTP client operations.

---

## Overview

Choco-go’s middleware pipeline consists of three main components:

- One or more **Policy** instances  
- A **Transporter** instance  
- A **Pipeline** instance that combines the Policies and the Transporter  

---

## Implementing the Policy Interface

A **Policy** encapsulates a step in the HTTP request lifecycle. Policies can be either:

- **Stateless:** implemented as a first-class function  
- **Stateful:** implemented as a method on a struct with internal state  

Policies intercept requests and responses, allowing you to add features such as tracing, retries, logging, or mutation.

> Note: HTTP requests passing through the same pipeline share the same Policy instances. If a Policy holds mutable state, it must be safe for concurrent use.

---

## Implementing the Transporter Interface

The **Transporter** is responsible for sending the HTTP request over the network and returning the response. It is always the last element in the pipeline.

Transporters can be implemented as stateful or stateless, similar to Policies.

The default Transporter uses the Go standard library's `http.Client`.

---

## Using Policy and Transporter Instances via a Pipeline

Create a pipeline by passing a Transporter and one or more Policy instances to `NewPipeline`:

```go
func NewPipeline(transport Transporter, policies ...Policy) Pipeline
```

The pipeline executes the Policies in order, followed by the Transporter:

```
Request -> PolicyA -> PolicyB -> PolicyC -> Transporter -> HTTP Endpoint
Response <- PolicyC <- PolicyB <- PolicyA <- Transporter
```

Example:

```go
pipeline := NewPipeline(transport, policyA, policyB, policyC)
```

Send a request through the pipeline using `Pipeline.Do`:

```go
req, err := NewRequest(ctx, "GET", "https://example.com")
if err != nil {
    // handle error
}

resp, err := pipeline.Do(req)
if err != nil {
    // handle error
}

// use resp...
```

---

## Creating a Request Instance

A `Request` wraps an `*http.Request` with extra internal state and helper methods.

Create a new request via:

```go
func NewRequest(ctx context.Context, method string, url string) (*Request, error)
```

If the request has a body, set it using:

```go
func (r *Request) SetBody(body io.ReadSeekCloser, contentType string) error
```

`io.ReadSeekCloser` is required to support retries, allowing the body stream to be rewound.

---

## Built-in Policies

Choco-go includes several built-in policies such as:

* **TracingPolicy:** adds tracing and diagnostics to requests
* **RetryPolicy:** automatically retries transient failures
* **LoggingPolicy:** logs requests and responses

---