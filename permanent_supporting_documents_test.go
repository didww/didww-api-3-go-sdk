package didww

import (
	"context"
	"io"
	"net/http"
	"testing"
)

func TestPermanentSupportingDocumentsCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/permanent_supporting_documents": {status: http.StatusCreated, fixture: "permanent_supporting_documents/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	doc, err := server.client.PermanentSupportingDocuments().Create(context.Background(), &PermanentSupportingDocument{
		TemplateID: "4199435f-646e-4e9d-a143-8f3b972b10c5",
		IdentityID: "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
		FileIDs:    []string{"254b3c2d-c40c-4ff7-93b1-a677aee7fa10"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.ID != "19510da3-c07e-4fa9-a696-6b9ab89cc172" {
		t.Errorf("expected ID '19510da3-c07e-4fa9-a696-6b9ab89cc172', got %q", doc.ID)
	}

	// Verify included template
	if doc.Template == nil {
		t.Fatal("expected non-nil Template")
	}
	if doc.Template.Name != "Germany Special Registration Form" {
		t.Errorf("expected template name 'Germany Special Registration Form', got %q", doc.Template.Name)
	}
	if !doc.Template.Permanent {
		t.Error("expected template Permanent to be true")
	}

	assertRequestJSON(t, capturedBody, "permanent_supporting_documents/create_request.json")
}

func TestPermanentSupportingDocumentsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/permanent_supporting_documents/19510da3-c07e-4fa9-a696-6b9ab89cc172": {status: http.StatusNoContent},
	})

	err := client.PermanentSupportingDocuments().Delete(context.Background(), "19510da3-c07e-4fa9-a696-6b9ab89cc172")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
