package didww

import (
	"context"
	"net/http"
	"testing"
)

func TestSupportingDocumentTemplatesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/supporting_document_templates": {status: http.StatusOK, fixture: "supporting_document_templates/index.json"},
	})

	templates, err := client.SupportingDocumentTemplates().List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(templates) != 5 {
		t.Fatalf("expected 5 templates, got %d", len(templates))
	}

	first := templates[0]
	if first.ID != "206ccec2-1166-461f-9f58-3a56823db548" {
		t.Errorf("expected ID '206ccec2-1166-461f-9f58-3a56823db548', got %q", first.ID)
	}
	if first.Name != "Generic LOI" {
		t.Errorf("expected Name 'Generic LOI', got %q", first.Name)
	}
	if first.Permanent {
		t.Error("expected Permanent to be false")
	}
}
