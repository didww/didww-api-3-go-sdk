package enums

import "testing"

func TestExportType(t *testing.T) {
	tests := []struct {
		name     string
		value    ExportType
		expected string
	}{
		{"CDR In", ExportTypeCdrIn, "cdr_in"},
		{"CDR Out", ExportTypeCdrOut, "cdr_out"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestExportStatus(t *testing.T) {
	tests := []struct {
		name     string
		value    ExportStatus
		expected string
	}{
		{"Pending", ExportStatusPending, "pending"},
		{"Processing", ExportStatusProcessing, "processing"},
		{"Completed", ExportStatusCompleted, "completed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}

func TestCallbackMethod(t *testing.T) {
	tests := []struct {
		name     string
		value    CallbackMethod
		expected string
	}{
		{"POST", CallbackMethodPOST, "post"},
		{"GET", CallbackMethodGET, "get"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.value) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.value))
			}
		})
	}
}
