package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
)

func TestPermanentSupportingDocumentsCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/permanent_supporting_documents": {status: http.StatusCreated, fixture: "permanent_supporting_documents/create.json"},
	})

	doc, err := server.client.PermanentSupportingDocuments().Create(context.Background(), &resource.PermanentSupportingDocument{
		TemplateID: "4199435f-646e-4e9d-a143-8f3b972b10c5",
		IdentityID: "5e9df058-50d2-4e34-b0d4-d1746b86f41a",
		FileIDs:    []string{"254b3c2d-c40c-4ff7-93b1-a677aee7fa10"},
	})
	require.NoError(t, err)

	assert.Equal(t, "19510da3-c07e-4fa9-a696-6b9ab89cc172", doc.ID)

	// Verify included template
	require.NotNil(t, doc.Template)
	assert.Equal(t, "Germany Special Registration Form", doc.Template.Name)
	assert.True(t, doc.Template.Permanent)

	assertRequestJSON(t, *capturedBodyPtr, "permanent_supporting_documents/create_request.json")
}

func TestPermanentSupportingDocumentsDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/permanent_supporting_documents/19510da3-c07e-4fa9-a696-6b9ab89cc172": {status: http.StatusNoContent},
	})

	err := client.PermanentSupportingDocuments().Delete(context.Background(), "19510da3-c07e-4fa9-a696-6b9ab89cc172")
	require.NoError(t, err)
}
