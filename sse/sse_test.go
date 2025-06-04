package sse_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"nyxze/choco-go/seqio"
	"nyxze/choco-go/sse"
)

type TestEvent struct {
	Event string `sse:"event"`
	Data  string `sse:"data"`
	Error []byte `sse:"error"`
}

func TestSseIter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []TestEvent
	}{
		{
			name: "basic event/data fields",
			input: strings.Join([]string{
				"event: start",
				"data: Hello",
				"",
				"event: update",
				"data: world!",
				"",
			}, "\n"),
			expected: []TestEvent{
				{Event: "start", Data: "Hello"},
				{Event: "update", Data: "world!"},
			},
		},
		{
			name: "handle byte slice for error field",
			input: strings.Join([]string{
				"event: update",
				"error: failed",
				"",
			}, "\n"),
			expected: []TestEvent{
				{Event: "update", Error: []byte("failed")},
			},
		},
		{
			name: "ignore unknown prefixes",
			input: strings.Join([]string{
				"foo: ignore me",
				"event: update",
				"data: value",
				"",
			}, "\n"),
			expected: []TestEvent{
				{Event: "update", Data: "value"},
			},
		},
		{
			name: "stop on [DONE] in data",
			input: strings.Join([]string{
				"event: update",
				"data: hello",
				"",
				"data: [DONE]",
				"",
			}, "\n"),
			expected: []TestEvent{
				{Event: "update", Data: "hello"},
			},
		},
		{
			name: "stop on [DONE] in event",
			input: strings.Join([]string{
				"data: still going",
				"event: [DONE]",
				"",
			}, "\n"),
			expected: []TestEvent{
				{Data: "still going"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := io.NopCloser(strings.NewReader(tt.input))
			iter := sse.NewSSEIter[TestEvent](reader, "[DONE]")

			ctx := context.Background()
			var results []TestEvent
			for v := range seqio.Range(ctx, iter) {
				results = append(results, v)
			}
			if err := iter.Err(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(results) != len(tt.expected) {
				t.Fatalf("expected %d events, got %d", len(tt.expected), len(results))
			}

			for i := range results {
				got := results[i]
				want := tt.expected[i]

				if got.Event != want.Event || got.Data != want.Data || !bytes.Equal(got.Error, want.Error) {
					t.Errorf("event %d mismatch\nGot:  %+v\nWant: %+v", i, got, want)
				}
			}
		})
	}
}
