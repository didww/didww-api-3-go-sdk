package enums

import (
	"encoding/json"
	"testing"
)

func testStringEnumJSON[T ~string](t *testing.T, value T, expectedJSON string) {
	t.Helper()

	// Marshal
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	if string(data) != expectedJSON {
		t.Errorf("expected marshalled %s, got %s", expectedJSON, string(data))
	}

	// Unmarshal
	var decoded T
	err = json.Unmarshal([]byte(expectedJSON), &decoded)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if decoded != value {
		t.Errorf("expected unmarshalled %v, got %v", value, decoded)
	}
}

func testIntEnumJSON[T ~int](t *testing.T, value T, expectedJSON string) {
	t.Helper()

	// Marshal
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}
	if string(data) != expectedJSON {
		t.Errorf("expected marshalled %s, got %s", expectedJSON, string(data))
	}

	// Unmarshal
	var decoded T
	err = json.Unmarshal([]byte(expectedJSON), &decoded)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}
	if decoded != value {
		t.Errorf("expected unmarshalled %v, got %v", value, decoded)
	}
}
