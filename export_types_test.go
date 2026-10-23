package didww

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v4/resource"
	"github.com/didww/didww-api-3-go-sdk/v4/resource/enums"
)

func TestExportsCreateSmsIn(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusCreated, fixture: "exports/create_sms_in.json"},
	})

	export, err := server.client.Exports().Create(context.Background(), &resource.Export{
		ExportType: enums.ExportTypeSmsIn,
		Filters: map[string]interface{}{
			"from":                   "2026-10-01 00:00:00",
			"to":                     "2026-10-15 00:00:00",
			"trunk_type":             []string{enums.ExportTrunkTypeSmppSmsc, enums.ExportTrunkTypeHTTPIn},
			"destination_address_eq": "12125550100",
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "exports/create_sms_in_request.json")

	assert.Equal(t, "5a2d3e4f-6b7c-4d8e-9f0a-1b2c3d4e5f60", export.ID)
	assert.Equal(t, enums.ExportTypeSmsIn, export.ExportType)
	assert.Equal(t, enums.ExportStatusPending, export.Status)
	assert.Equal(t, []interface{}{"SMPP SMSC", "HTTP IN"}, export.Filters["trunk_type"])
}

func TestExportsCreateSmsOut(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/exports": {status: http.StatusCreated, fixture: "exports/create_sms_out.json"},
	})

	export, err := server.client.Exports().Create(context.Background(), &resource.Export{
		ExportType: enums.ExportTypeSmsOut,
		Filters: map[string]interface{}{
			"from": "2026-10-01 00:00:00",
			"to":   "2026-10-15 00:00:00",
			"sms_trunk.id": []string{
				"b2c3d4e5-0000-4000-8000-000000000001",
				"b2c3d4e5-0000-4000-8000-000000000003",
			},
			"service_type":              "a2p",
			"sender_id_verification.id": "c3d4e5f6-0000-4000-8000-000000000002",
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "exports/create_sms_out_request.json")

	assert.Equal(t, "4f1c2d3e-5a6b-4c7d-8e9f-0a1b2c3d4e5f", export.ID)
	assert.Equal(t, enums.ExportTypeSmsOut, export.ExportType)
	assert.True(t, export.IsPending())
	assert.Equal(t, "a2p", export.Filters["service_type"])
	assert.Equal(t, "c3d4e5f6-0000-4000-8000-000000000002", export.Filters["sender_id_verification.id"])
}
