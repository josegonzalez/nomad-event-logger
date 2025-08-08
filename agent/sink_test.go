package agent

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestStdoutSink(t *testing.T) {
	tests := []struct {
		name     string
		event    *Event
		expected string
	}{
		{
			name: "basic event",
			event: &Event{
				Time: time.Unix(1640995200, 0).UTC(),
				Type: "test",
				Data: map[string]any{
					"message": "hello world",
				},
			},
			expected: `{"time":"2022-01-01T00:00:00Z","type":"test","data":{"message":"hello world"}}`,
		},
		{
			name: "event with complex data",
			event: &Event{
				Time: time.Unix(1640995200, 0).UTC(),
				Type: "allocation",
				Data: map[string]any{
					"id":      "alloc-123",
					"status":  "running",
					"metrics": []int{1, 2, 3},
				},
			},
			expected: `{"time":"2022-01-01T00:00:00Z","type":"allocation","data":{"id":"alloc-123","metrics":[1,2,3],"status":"running"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			sink := NewStdoutSink()
			err := sink.Write(tt.event)

			// Restore stdout
			w.Close()
			os.Stdout = oldStdout

			// Read captured output
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to read captured output: %v", err)
			}
			output := strings.TrimSpace(buf.String())

			if err != nil {
				t.Errorf("StdoutSink.Write() error = %v", err)
				return
			}

			// Parse and compare JSON (order doesn't matter)
			var expected, actual map[string]any
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(output), &actual); err != nil {
				t.Fatalf("Failed to parse actual JSON: %v", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("StdoutSink.Write() output = %v, want %v", output, tt.expected)
			}
		})
	}
}

func TestFileSink(t *testing.T) {
	tests := []struct {
		name     string
		event    *Event
		expected string
	}{
		{
			name: "basic event",
			event: &Event{
				Time: time.Unix(1640995200, 0).UTC(),
				Type: "test",
				Data: map[string]any{
					"message": "hello world",
				},
			},
			expected: `{"time":"2022-01-01T00:00:00Z","type":"test","data":{"message":"hello world"}}`,
		},
		{
			name: "multiple events",
			event: &Event{
				Time: time.Unix(1640995200, 0).UTC(),
				Type: "allocation",
				Data: map[string]any{
					"id":     "alloc-123",
					"status": "running",
				},
			},
			expected: `{"time":"2022-01-01T00:00:00Z","type":"allocation","data":{"id":"alloc-123","status":"running"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "test")
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tmpfile.Name())

			sink, err := NewFileSink(tmpfile.Name())
			if err != nil {
				t.Fatalf("Failed to create file sink: %v", err)
			}
			defer sink.Close()

			err = sink.Write(tt.event)
			if err != nil {
				t.Errorf("FileSink.Write() error = %v", err)
				return
			}

			// Read file content
			content, err := os.ReadFile(tmpfile.Name())
			if err != nil {
				t.Fatalf("Failed to read file: %v", err)
			}

			output := strings.TrimSpace(string(content))

			// Parse and compare JSON (order doesn't matter)
			var expected, actual map[string]any
			if err := json.Unmarshal([]byte(tt.expected), &expected); err != nil {
				t.Fatalf("Failed to parse expected JSON: %v", err)
			}
			if err := json.Unmarshal([]byte(output), &actual); err != nil {
				t.Fatalf("Failed to parse actual JSON: %v", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("FileSink.Write() output = %v, want %v", output, tt.expected)
			}
		})
	}
}

func TestFileSink_AppendMode(t *testing.T) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	sink, err := NewFileSink(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create file sink: %v", err)
	}
	defer sink.Close()

	// Write first event
	event1 := &Event{
		Time: time.Unix(1640995200, 0).UTC(),
		Type: "test1",
		Data: map[string]any{
			"message": "first",
		},
	}

	err = sink.Write(event1)
	if err != nil {
		t.Errorf("FileSink.Write() error = %v", err)
	}

	// Write second event
	event2 := &Event{
		Time: time.Unix(1640995201, 0).UTC(),
		Type: "test2",
		Data: map[string]any{
			"message": "second",
		},
	}

	err = sink.Write(event2)
	if err != nil {
		t.Errorf("FileSink.Write() error = %v", err)
	}

	// Read file content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(lines))
	}

	// Verify both events are present
	var event1Map, event2Map map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &event1Map); err != nil {
		t.Fatalf("Failed to parse first event: %v", err)
	}
	if err := json.Unmarshal([]byte(lines[1]), &event2Map); err != nil {
		t.Fatalf("Failed to parse second event: %v", err)
	}

	if event1Map["type"] != "test1" {
		t.Errorf("Expected first event type 'test1', got %v", event1Map["type"])
	}
	if event2Map["type"] != "test2" {
		t.Errorf("Expected second event type 'test2', got %v", event2Map["type"])
	}
}

func TestFileSink_InvalidPath(t *testing.T) {
	_, err := NewFileSink("/invalid/path/test")
	if err == nil {
		t.Error("Expected error for invalid path, got nil")
	}
}

func TestFileSink_Close(t *testing.T) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	sink, err := NewFileSink(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to create file sink: %v", err)
	}

	// Write an event
	event := &Event{
		Time: time.Unix(1640995200, 0).UTC(),
		Type: "test",
		Data: map[string]any{
			"message": "hello world",
		},
	}

	err = sink.Write(event)
	if err != nil {
		t.Errorf("FileSink.Write() error = %v", err)
	}

	// Close the sink
	err = sink.Close()
	if err != nil {
		t.Errorf("FileSink.Close() error = %v", err)
	}

	// Try to write after closing (should return an error)
	err = sink.Write(event)
	if err == nil {
		t.Error("Expected error when writing to closed file, got nil")
	}
}

func TestSinkInterface(t *testing.T) {
	var sink Sink = NewStdoutSink()
	// Test that the sink can be assigned to the interface
	_ = sink
}

func BenchmarkStdoutSink_Write(b *testing.B) {
	sink := NewStdoutSink()
	event := &Event{
		Time: time.Unix(1640995200, 0).UTC(),
		Type: "test",
		Data: map[string]any{
			"message": "hello world",
			"number":  42,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sink.Write(event)
		if err != nil {
			b.Fatalf("StdoutSink.Write() error = %v", err)
		}
	}
}

func BenchmarkFileSink_Write(b *testing.B) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "benchmark")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	sink, err := NewFileSink(tmpfile.Name())
	if err != nil {
		b.Fatalf("Failed to create file sink: %v", err)
	}
	defer sink.Close()

	event := &Event{
		Time: time.Unix(1640995200, 0).UTC(),
		Type: "test",
		Data: map[string]any{
			"message": "hello world",
			"number":  42,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := sink.Write(event)
		if err != nil {
			b.Fatalf("FileSink.Write() error = %v", err)
		}
	}
}
