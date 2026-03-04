package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSupportingDocumentTemplatesList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/supporting_document_templates": {status: http.StatusOK, fixture: "supporting_document_templates/index.json"},
	})

	templates, err := client.SupportingDocumentTemplates().List(context.Background(), nil)
	require.NoError(t, err)

	require.Len(t, templates, 5)

	first := templates[0]
	assert.Equal(t, "206ccec2-1166-461f-9f58-3a56823db548", first.ID)
	assert.Equal(t, "Generic LOI", first.Name)
	assert.False(t, first.Permanent)
}
