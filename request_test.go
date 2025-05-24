package choco

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

const testURL = "http://www.example.com/"

type fakeExec struct {
}

func (fakeExec) Send(*http.Request) (*http.Response, error) {
	header := http.Header{}
	header.Add("Ping", "Pong")

	return &http.Response{
		Header: header,
	}, nil
}

func TestNewRequest(t *testing.T) {
	ctx := context.Background()
	req, err := NewRequest(ctx, "GET", testURL)
	if err != nil {
		t.Fatal(err)
	}
	if m := req.Raw().Method; m != http.MethodGet {
		t.Fatalf("unexpected method %s", m)
	}
}
func TestPipelineSteps(t *testing.T) {
	type testcase struct {
		name         string
		steps        []PipelineStep
		expectedPing string
		expectError  bool
		expectLog    []string
	}

	state := statefullStep{}
	tests := []testcase{
		{
			name:         "stateful step logs before/after",
			steps:        []PipelineStep{&state},
			expectedPing: "Pong",
			expectError:  false,
			expectLog:    []string{"stateful-before", "stateful-after"},
		},
		{
			name:         "error short-circuits the pipeline",
			steps:        []PipelineStep{errorStep(), &state},
			expectedPing: "",
			expectError:  true,
			expectLog:    nil, // statefulStep should not run
		},
		{
			name:         "stateful and header removal",
			steps:        []PipelineStep{&state, removeHeaders()},
			expectedPing: "", // removed by RemoveHeaders
			expectError:  false,
			expectLog:    []string{"stateful-before", "stateful-after"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state.logs = nil
			p, err := NewPipeline(
				WithCustomTransport(fakeExec{}),
				WithSteps(tt.steps...),
			)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			req, err := NewRequest(ctx, "GET", testURL)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := p.Execute(req)
			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return // skip further checks
			}
			if err != nil {
				t.Fatalf("pipeline error: %v", err)
			}

			if got := resp.Header.Get("Ping"); got != tt.expectedPing {
				t.Errorf("expected header Ping=%q, got %q", tt.expectedPing, got)
			}
			if tt.expectLog != nil {
				if !equalStringSlice(state.logs, tt.expectLog) {
					t.Errorf("expected log %+v, got %+v", tt.expectLog, state)
				}
			}
		})
	}
}
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Steps
func removeHeaders() PipelineStep {
	return PipelineStepFunc(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(r *Request) (*http.Response, error) {
			r.Raw().Header.Set("Ping", "")
			resp, err := next(r)
			if resp != nil {
				resp.Header.Del("Ping")
			}
			return resp, err
		}
	})
}

func errorStep() PipelineStep {
	return PipelineStepFunc(func(next RequestHandlerFunc) RequestHandlerFunc {
		return func(r *Request) (*http.Response, error) {
			return nil, fmt.Errorf("injected error")
		}
	})
}

type statefullStep struct {
	logs []string
}

func (s *statefullStep) Do(next RequestHandlerFunc) RequestHandlerFunc {
	return func(r *Request) (*http.Response, error) {
		s.logs = append(s.logs, "stateful-before")
		resp, err := next(r)
		if err == nil {
			s.logs = append(s.logs, "stateful-after")
		}
		return resp, err
	}
}
