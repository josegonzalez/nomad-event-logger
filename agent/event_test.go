package agent

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestNewEvent(t *testing.T) {
	eventType := "test"
	data := map[string]any{
		"key":    "value",
		"number": 42,
	}

	event := NewEvent(eventType, data)

	// Check that time is set (should be close to current time)
	now := time.Now()
	if event.Time.Before(now.Add(-5*time.Second)) || event.Time.After(now.Add(5*time.Second)) {
		t.Errorf("Time %v is not close to current time %v", event.Time, now)
	}

	if event.Type != eventType {
		t.Errorf("Expected event type %s, got %s", eventType, event.Type)
	}

	// Compare data using reflect.DeepEqual
	if !reflect.DeepEqual(event.Data, data) {
		t.Errorf("Expected data %v, got %v", data, event.Data)
	}
}

func TestEventToJSON(t *testing.T) {
	event := NewEvent("test", map[string]any{
		"message": "hello world",
	})

	jsonData, err := event.ToJSON()
	if err != nil {
		t.Fatalf("Failed to marshal event to JSON: %v", err)
	}

	// Verify JSON structure
	var parsed map[string]any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check required fields
	if _, ok := parsed["time"]; !ok {
		t.Error("JSON missing time field")
	}

	if _, ok := parsed["type"]; !ok {
		t.Error("JSON missing type field")
	}

	if _, ok := parsed["data"]; !ok {
		t.Error("JSON missing data field")
	}

	if parsed["type"] != "test" {
		t.Errorf("Expected type 'test', got %v", parsed["type"])
	}
}
