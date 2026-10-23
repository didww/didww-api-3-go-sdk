package didww

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/didww/didww-api-3-go-sdk/v3/resource"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/v3/resource/smstrunkconfiguration"
)

const (
	testSmsTrunkHTTPInID  = "1a2b3c4d-0002-4a5b-8c6d-7e8f9a0b1c02"
	testSmsTrunkHTTPOutID = "1a2b3c4d-0003-4a5b-8c6d-7e8f9a0b1c03"
	testSmsTrunkUnknownID = "1a2b3c4d-0006-4a5b-8c6d-7e8f9a0b1c06"
	testSmsTrunkGroupID   = "2b3c4d5e-0001-4b6c-9d7e-8f9a0b1c2d01"
)

func TestSmsTrunksList(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunks": {status: http.StatusOK, fixture: "sms_trunks/index.json"},
	})

	trunks, err := client.SmsTrunks().List(context.Background(), nil)
	require.NoError(t, err)
	require.Len(t, trunks, 5)

	types := make([]string, len(trunks))
	for i, trunk := range trunks {
		require.NotNil(t, trunk.Configuration)
		types[i] = trunk.Configuration.ConfigurationType()
	}
	assert.Equal(t, []string{
		smstrunkconfiguration.TypeSmsToEmail,
		smstrunkconfiguration.TypeHTTPIn,
		smstrunkconfiguration.TypeHTTPOut,
		smstrunkconfiguration.TypeSmppEsme,
		smstrunkconfiguration.TypeSmppSmsc,
	}, types)

	email := trunks[0]
	assert.Equal(t, "1a2b3c4d-0001-4a5b-8c6d-7e8f9a0b1c01", email.ID)
	assert.Equal(t, "sms_to_email trunk", email.Name)
	assert.Nil(t, email.ExternalReferenceID)
	assert.False(t, email.Blocked)
	require.NotNil(t, email.Priority)
	assert.Equal(t, 100, *email.Priority)
	assert.False(t, email.AllowAnyDIDAsSource)
	assert.True(t, email.AllowIn)
	assert.False(t, email.AllowOut)
	assert.Equal(t, time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC), email.CreatedAt)
	assert.Equal(t, time.Date(2026, 10, 1, 10, 0, 0, 0, time.UTC), email.UpdatedAt)
	emailCfg, ok := email.Configuration.(*smstrunkconfiguration.SmsToEmail)
	require.True(t, ok, "expected sms_to_email configuration")
	assert.Equal(t, "sms@example.com", emailCfg.SendTo)
	assert.Equal(t, "Inbound SMS", emailCfg.Subject)
	assert.Equal(t, "You have a new message.", emailCfg.Message)

	httpIn, ok := trunks[1].Configuration.(*smstrunkconfiguration.HTTPIn)
	require.True(t, ok, "expected http_in configuration")
	assert.Equal(t, "https://example.com/sms", httpIn.URL)
	assert.Equal(t, smstrunkconfiguration.HTTPMethodPost, httpIn.HTTPMethod)
	assert.Nil(t, httpIn.QueryParameters)
	assert.Equal(t, map[string]any{"X-Source": "didww"}, httpIn.Headers)
	assert.Equal(t, smstrunkconfiguration.BodyTypeJSON, httpIn.BodyType)
	assert.Equal(t, `{"text":"{SMS_TEXT}","from":"{SMS_SRC_ADDR}"}`, httpIn.Body)
	require.NotNil(t, httpIn.FollowRedirect)
	assert.False(t, *httpIn.FollowRedirect)
	require.NotNil(t, httpIn.ValidateHTTPSCertificate)
	assert.True(t, *httpIn.ValidateHTTPSCertificate)

	out := trunks[2]
	assert.Nil(t, out.Priority)
	assert.True(t, out.AllowOut)
	assert.Equal(t, "https://sms-out.example.com/outbound_messages", out.Meta["host"])
	httpOut, ok := out.Configuration.(*smstrunkconfiguration.HTTPOut)
	require.True(t, ok, "expected http_out configuration")
	assert.Equal(t, "https://example.com/sms/status", httpOut.CallbackURL)
	assert.Equal(t, []string{"203.0.113.10"}, httpOut.AllowedIPs)
	assert.Equal(t, "generated-name", httpOut.AuthName)
	assert.Equal(t, "generated-password", httpOut.AuthPassword)

	esme, ok := trunks[3].Configuration.(*smstrunkconfiguration.SmppEsme)
	require.True(t, ok, "expected smpp_esme configuration")
	assert.Empty(t, esme.SystemType)
	assert.Equal(t, []string{"203.0.113.10"}, esme.AllowedIPs)
	assert.Equal(t, smstrunkconfiguration.ServerLocationNYC, esme.ServerLocation)
	assert.Equal(t, "esme1234", esme.SystemID)
	assert.Equal(t, "gen12345", esme.Password)

	smsc, ok := trunks[4].Configuration.(*smstrunkconfiguration.SmppSmsc)
	require.True(t, ok, "expected smpp_smsc configuration")
	assert.Equal(t, "smsc.example.com", smsc.Host)
	require.NotNil(t, smsc.TxPort)
	assert.Equal(t, 2775, *smsc.TxPort)
	require.NotNil(t, smsc.RxPort)
	assert.Equal(t, 2776, *smsc.RxPort)
	require.NotNil(t, smsc.TransceiverMode)
	assert.True(t, *smsc.TransceiverMode)
	require.NotNil(t, smsc.ConnectionTimeout)
	assert.Equal(t, 30, *smsc.ConnectionTimeout)
	require.NotNil(t, smsc.UseSSL)
	assert.False(t, *smsc.UseSSL)
	assert.Equal(t, "didww", smsc.SystemID)
	assert.Empty(t, smsc.SystemType)
	assert.Equal(t, "secret", smsc.Password)
	assert.Equal(t, smstrunkconfiguration.ServerLocationNYC, smsc.ServerLocation)
}

func TestSmsTrunksListWithIncludedSmsTrunkGroup(t *testing.T) {
	var query string
	server := newTestServerWithInspector(t, map[string]testRoute{
		"GET /v3/sms_trunks": {status: http.StatusOK, fixture: "sms_trunks/index_with_sms_trunk_group.json"},
	}, func(r *http.Request) {
		if r.URL.Path == "/v3/sms_trunks" {
			query = r.URL.RawQuery
		}
	})

	params := NewQueryParams().
		Filter("configuration.type", smstrunkconfiguration.TypeHTTPIn).
		Include("sms_trunk_group").
		Sort("-priority")
	trunks, err := server.client.SmsTrunks().List(context.Background(), params)
	require.NoError(t, err)

	assert.Contains(t, query, "filter[configuration.type]=http_in")
	assert.Contains(t, query, "include=sms_trunk_group")
	require.Len(t, trunks, 1)
	require.NotNil(t, trunks[0].SmsTrunkGroup)
	assert.Equal(t, testSmsTrunkGroupID, trunks[0].SmsTrunkGroup.ID)
	assert.Equal(t, "Inbound SMS", trunks[0].SmsTrunkGroup.Name)
}

func TestSmsTrunksFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunks/" + testSmsTrunkHTTPOutID: {status: http.StatusOK, fixture: "sms_trunks/show.json"},
	})

	trunk, err := client.SmsTrunks().Find(context.Background(), testSmsTrunkHTTPOutID)
	require.NoError(t, err)

	assert.Equal(t, testSmsTrunkHTTPOutID, trunk.ID)
	assert.Equal(t, "http_out trunk", trunk.Name)
	assert.Equal(t, "https://sms-out.example.com/outbound_messages", trunk.Meta["host"])
	cfg, ok := trunk.Configuration.(*smstrunkconfiguration.HTTPOut)
	require.True(t, ok, "expected http_out configuration")
	assert.Equal(t, "generated-password", cfg.AuthPassword)
}

func TestSmsTrunksFindWithIncludedSenderIDVerifications(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunks/" + testSmsTrunkHTTPOutID: {status: http.StatusOK, fixture: "sms_trunks/show_with_sender_id_verifications.json"},
	})

	params := NewQueryParams().Include("sender_id_verifications")
	trunk, err := client.SmsTrunks().Find(context.Background(), testSmsTrunkHTTPOutID, params)
	require.NoError(t, err)

	require.Len(t, trunk.SenderIDVerifications, 1)
	siv := trunk.SenderIDVerifications[0]
	assert.Equal(t, "e5a4e2f1-3b26-4a5c-9a0a-2f7c1f3e8d10", siv.ID)
	assert.Equal(t, "ACME alerts", siv.Name)
	assert.Equal(t, enums.SenderIDVerificationStatusActive, siv.Status)
	assert.Equal(t, enums.SenderTypeLongCode, siv.SenderType)
}

func TestSmsTrunksFindNotFound(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/sms_trunks/" + testSmsTrunkHTTPOutID: {status: http.StatusNotFound, fixture: "sms_trunks/not_found.json"},
	})

	_, err := client.SmsTrunks().Find(context.Background(), testSmsTrunkHTTPOutID)
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusNotFound, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "Record not found", apiErr.Errors[0].Title)
}

func TestSmsTrunksCreatePerConfigurationType(t *testing.T) {
	tests := []struct {
		configType string
		config     smstrunkconfiguration.SmsTrunkConfiguration
	}{
		{smstrunkconfiguration.TypeSmsToEmail, &smstrunkconfiguration.SmsToEmail{
			SendTo:  "sms@example.com",
			Subject: "Inbound SMS",
			Message: "You have a new message.",
		}},
		{smstrunkconfiguration.TypeHTTPIn, &smstrunkconfiguration.HTTPIn{
			URL:                      "https://example.com/sms",
			HTTPMethod:               smstrunkconfiguration.HTTPMethodPost,
			Headers:                  map[string]any{"X-Source": "didww"},
			BodyType:                 smstrunkconfiguration.BodyTypeJSON,
			Body:                     `{"text":"{SMS_TEXT}","from":"{SMS_SRC_ADDR}"}`,
			FollowRedirect:           Ptr(false),
			ValidateHTTPSCertificate: Ptr(true),
		}},
		{smstrunkconfiguration.TypeHTTPOut, &smstrunkconfiguration.HTTPOut{
			CallbackURL: "https://example.com/sms/status",
			AllowedIPs:  []string{"203.0.113.10"},
			// Read-only credentials are never sent.
			AuthName:     "ignored-name",
			AuthPassword: "ignored-password",
		}},
		{smstrunkconfiguration.TypeSmppEsme, &smstrunkconfiguration.SmppEsme{
			AllowedIPs:     []string{"203.0.113.10"},
			ServerLocation: smstrunkconfiguration.ServerLocationNYC,
			// Read-only credentials are never sent.
			SystemID: "ignored-id",
			Password: "ignored-password",
		}},
		{smstrunkconfiguration.TypeSmppSmsc, &smstrunkconfiguration.SmppSmsc{
			Host:              "smsc.example.com",
			TxPort:            Ptr(2775),
			RxPort:            Ptr(2776),
			TransceiverMode:   Ptr(true),
			ConnectionTimeout: Ptr(30),
			UseSSL:            Ptr(false),
			SystemID:          "didww",
			Password:          "secret",
			ServerLocation:    smstrunkconfiguration.ServerLocationNYC,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.configType, func(t *testing.T) {
			server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
				"POST /v3/sms_trunks": {status: http.StatusCreated, fixture: "sms_trunks/create_" + tt.configType + ".json"},
			})

			trunk, err := server.client.SmsTrunks().Create(context.Background(), &resource.SmsTrunk{
				Name:          tt.configType + " trunk",
				Configuration: tt.config,
			})
			require.NoError(t, err)

			assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/create_"+tt.configType+"_request.json")
			assert.NotEmpty(t, trunk.ID)
			assert.Equal(t, tt.configType+" trunk", trunk.Name)
			require.NotNil(t, trunk.Configuration)
			assert.Equal(t, tt.configType, trunk.Configuration.ConfigurationType())
			assert.IsType(t, tt.config, trunk.Configuration)
		})
	}
}

func TestSmsTrunksCreateWithGroupAndOutgoingDIDs(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/sms_trunks": {status: http.StatusCreated, fixture: "sms_trunks/create_smpp_smsc.json"},
	})

	_, err := server.client.SmsTrunks().Create(context.Background(), &resource.SmsTrunk{
		Name:            "smpp_smsc trunk",
		Configuration:   &smstrunkconfiguration.SmppSmsc{Host: "smsc.example.com"},
		SmsTrunkGroupID: testSmsTrunkGroupID,
		OutgoingDIDIDs:  []string{"9df99644-f1a5-4a3c-99a4-559d758eb96b"},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/create_with_relationships_request.json")
}

func TestSmsTrunksCreateError(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/sms_trunks": {status: http.StatusUnprocessableEntity, fixture: "sms_trunks/create_error.json"},
	})

	_, err := server.client.SmsTrunks().Create(context.Background(), &resource.SmsTrunk{
		Name:          "sms_to_email trunk",
		Configuration: &smstrunkconfiguration.SmsToEmail{},
	})
	require.Error(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/create_error_request.json")
	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
	require.Len(t, apiErr.Errors, 1)
	assert.Equal(t, "configuration - send_to can't be blank", apiErr.Errors[0].Detail)
	assert.Equal(t, "/data/attributes/configuration", apiErr.Errors[0].Source.Pointer)
}

func TestSmsTrunksUpdateNameAndConfiguration(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sms_trunks/" + testSmsTrunkHTTPInID: {status: http.StatusOK, fixture: "sms_trunks/update_http_in.json"},
	})

	trunk, err := server.client.SmsTrunks().Update(context.Background(), &resource.SmsTrunk{
		ID:            testSmsTrunkHTTPInID,
		Name:          "Webhook",
		Configuration: &smstrunkconfiguration.HTTPIn{URL: "https://example.com/sms/v2"},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/update_http_in_request.json")
	assert.Equal(t, "Webhook", trunk.Name)
	cfg, ok := trunk.Configuration.(*smstrunkconfiguration.HTTPIn)
	require.True(t, ok, "expected http_in configuration")
	assert.Equal(t, "https://example.com/sms/v2", cfg.URL)
	assert.Equal(t, smstrunkconfiguration.HTTPMethodPost, cfg.HTTPMethod)
}

func TestSmsTrunksUpdateFromLoadedDoesNotSendCredentials(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"GET /v3/sms_trunks/" + testSmsTrunkHTTPOutID:   {status: http.StatusOK, fixture: "sms_trunks/show.json"},
		"PATCH /v3/sms_trunks/" + testSmsTrunkHTTPOutID: {status: http.StatusOK, fixture: "sms_trunks/update_http_out.json"},
	})

	trunk, err := server.client.SmsTrunks().Find(context.Background(), testSmsTrunkHTTPOutID)
	require.NoError(t, err)

	cfg, ok := trunk.Configuration.(*smstrunkconfiguration.HTTPOut)
	require.True(t, ok, "expected http_out configuration")
	cfg.CallbackURL = "https://example.com/sms/status/v2"

	updated, err := server.client.SmsTrunks().Update(context.Background(), trunk)
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/update_http_out_request.json")
	updatedCfg, ok := updated.Configuration.(*smstrunkconfiguration.HTTPOut)
	require.True(t, ok, "expected http_out configuration")
	assert.Equal(t, "https://example.com/sms/status/v2", updatedCfg.CallbackURL)
	assert.Equal(t, "generated-password", updatedCfg.AuthPassword)
}

func TestSmsTrunksUpdateFromLoadedWithoutChangesSendsNothing(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"GET /v3/sms_trunks/" + testSmsTrunkHTTPOutID:   {status: http.StatusOK, fixture: "sms_trunks/show.json"},
		"PATCH /v3/sms_trunks/" + testSmsTrunkHTTPOutID: {status: http.StatusOK, fixture: "sms_trunks/show.json"},
	})

	trunk, err := server.client.SmsTrunks().Find(context.Background(), testSmsTrunkHTTPOutID)
	require.NoError(t, err)

	_, err = server.client.SmsTrunks().Update(context.Background(), trunk)
	require.NoError(t, err)

	doc := parsePatchBody(t, *capturedBodyPtr)
	assert.Empty(t, doc.Attrs)
	assert.Empty(t, doc.Rels)
}

func TestSmsTrunksUpdateClearRelationships(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sms_trunks/" + testSmsTrunkHTTPInID: {status: http.StatusOK, fixture: "sms_trunks/update_http_in.json"},
	})

	_, err := server.client.SmsTrunks().Update(context.Background(), &resource.SmsTrunk{
		ID:                   testSmsTrunkHTTPInID,
		ClearOutgoingDIDs:    true,
		NullifySmsTrunkGroup: true,
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/update_clear_relationships_request.json")
}

func TestSmsTrunksUnknownConfigurationIsPreserved(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"GET /v3/sms_trunks/" + testSmsTrunkUnknownID:   {status: http.StatusOK, fixture: "sms_trunks/show_unknown_configuration.json"},
		"PATCH /v3/sms_trunks/" + testSmsTrunkUnknownID: {status: http.StatusOK, fixture: "sms_trunks/show_unknown_configuration.json"},
	})

	trunk, err := server.client.SmsTrunks().Find(context.Background(), testSmsTrunkUnknownID)
	require.NoError(t, err)

	generic, ok := trunk.Configuration.(*smstrunkconfiguration.Generic)
	require.True(t, ok, "expected generic configuration")
	assert.Equal(t, "future_transport", generic.Type)
	assert.Equal(t, "future_transport", generic.ConfigurationType())
	assert.Equal(t, "https://example.com/future", generic.Attributes["endpoint"])
	assert.Equal(t, float64(3), generic.Attributes["retries"])
	assert.Equal(t, true, generic.Attributes["enabled"])

	generic.Attributes["endpoint"] = "https://example.com/future/v2"
	_, err = server.client.SmsTrunks().Update(context.Background(), trunk)
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/update_unknown_configuration_request.json")
}

func TestSmsTrunksUpdateGenericSendsNulls(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"PATCH /v3/sms_trunks/" + testSmsTrunkHTTPInID: {status: http.StatusOK, fixture: "sms_trunks/update_http_in.json"},
	})

	_, err := server.client.SmsTrunks().Update(context.Background(), &resource.SmsTrunk{
		ID: testSmsTrunkHTTPInID,
		Configuration: &smstrunkconfiguration.Generic{
			Type:       smstrunkconfiguration.TypeHTTPIn,
			Attributes: map[string]any{"http_method": "get", "body_type": nil, "body": nil},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "sms_trunks/update_generic_nulls_request.json")
}

func TestSmsTrunksDelete(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/sms_trunks/" + testSmsTrunkHTTPInID: {status: http.StatusNoContent},
	})

	err := client.SmsTrunks().Delete(context.Background(), testSmsTrunkHTTPInID)
	require.NoError(t, err)
}

func TestSmsTrunksDeleteInUse(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"DELETE /v3/sms_trunks/" + testSmsTrunkHTTPInID: {status: http.StatusUnprocessableEntity, fixture: "sms_trunks/delete_error.json"},
	})

	err := client.SmsTrunks().Delete(context.Background(), testSmsTrunkHTTPInID)
	require.Error(t, err)

	var apiErr *APIError
	require.True(t, errors.As(err, &apiErr), "expected *APIError")
	assert.Equal(t, http.StatusUnprocessableEntity, apiErr.HTTPStatus)
	assert.Equal(t, "base - Cannot delete an SMS trunk while DIDs are routed to it", apiErr.Errors[0].Detail)
}

func TestSmsTrunkConfigurationsRedactSecrets(t *testing.T) {
	tests := []struct {
		name    string
		config  any
		secrets []string
		visible []string
	}{
		{
			name:    "http_out",
			config:  &smstrunkconfiguration.HTTPOut{CallbackURL: "https://example.com/status", AuthName: "generated-name", AuthPassword: "generated-password"},
			secrets: []string{"generated-name", "generated-password"},
			visible: []string{"https://example.com/status"},
		},
		{
			name:    "smpp_esme",
			config:  &smstrunkconfiguration.SmppEsme{ServerLocation: "nyc", SystemID: "esme1234", Password: "gen12345"},
			secrets: []string{"esme1234", "gen12345"},
			visible: []string{"nyc"},
		},
		{
			name:    "smpp_smsc",
			config:  &smstrunkconfiguration.SmppSmsc{Host: "smsc.example.com", SystemID: "didww", Password: "secret"},
			secrets: []string{"secret"},
			visible: []string{"smsc.example.com", "didww"},
		},
		{
			name:    "http_in",
			config:  &smstrunkconfiguration.HTTPIn{URL: "https://example.com/sms", Headers: map[string]any{"Authorization": "Bearer abc123"}},
			secrets: []string{"Authorization", "Bearer abc123"},
			visible: []string{"https://example.com/sms"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trunk := &resource.SmsTrunk{Name: "trunk", Configuration: tt.config.(smstrunkconfiguration.SmsTrunkConfiguration)}
			for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
				out := fmt.Sprintf(verb, tt.config)
				for _, secret := range tt.secrets {
					assert.NotContains(t, out, secret, "verb %s leaked a secret", verb)
				}
				for _, value := range tt.visible {
					assert.Contains(t, out, value, "verb %s", verb)
				}
				assert.Contains(t, out, "[FILTERED]", "verb %s", verb)
			}
			nested := fmt.Sprintf("%+v", trunk)
			for _, secret := range tt.secrets {
				assert.NotContains(t, nested, secret, "nested configuration leaked a secret")
			}
		})
	}
}

func TestSmsTrunkConfigurationRedactionKeepsWireValues(t *testing.T) {
	cfg := &smstrunkconfiguration.SmppSmsc{Host: "smsc.example.com", Password: "secret"}
	raw, err := smstrunkconfiguration.Marshal(cfg)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"smpp_smsc","attributes":{"host":"smsc.example.com","password":"secret"}}`, string(raw))
}

func TestSmsTrunkConfigurationRoundTrip(t *testing.T) {
	tests := []string{
		`{"type":"sms_to_email","attributes":{"send_to":"sms@example.com","subject":"Inbound SMS","message":"Hi"}}`,
		`{"type":"http_in","attributes":{"url":"https://example.com/sms","http_method":"put","query_parameters":{"source":"didww"},` +
			`"headers":{"X-Source":"didww"},"body_type":"html_url_encoded","body":{"text":"{SMS_TEXT}"},"follow_redirect":true,"validate_https_certificate":false}}`,
		`{"type":"http_out","attributes":{"callback_url":"https://example.com/status","allowed_ips":["203.0.113.10"]}}`,
		`{"type":"smpp_esme","attributes":{"system_type":"cmt","allowed_ips":["203.0.113.10"],"server_location":"ams"}}`,
		`{"type":"smpp_smsc","attributes":{"host":"smsc.example.com","tx_port":2775,"rx_port":2776,"transceiver_mode":false,` +
			`"connection_timeout":0,"use_ssl":true,"system_id":"didww","system_type":"cmt","password":"secret","server_location":"nyc"}}`,
		`{"type":"future_transport","attributes":{"endpoint":"https://example.com/future","nested":{"a":[1,2]}}}`,
	}
	for _, input := range tests {
		cfg, err := smstrunkconfiguration.Parse([]byte(input))
		require.NoError(t, err)
		out, err := smstrunkconfiguration.Marshal(cfg)
		require.NoError(t, err)
		assert.JSONEq(t, input, string(out))
	}
}

func TestSmsTrunkConfigurationMarshalStripsReadOnlyCredentials(t *testing.T) {
	esme, err := smstrunkconfiguration.Parse([]byte(`{"type":"smpp_esme","attributes":{"server_location":"nyc","system_id":"esme1234","password":"gen12345"}}`))
	require.NoError(t, err)
	out, err := smstrunkconfiguration.Marshal(esme)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"smpp_esme","attributes":{"server_location":"nyc"}}`, string(out))

	httpOut, err := smstrunkconfiguration.Parse([]byte(`{"type":"http_out","attributes":{"callback_url":null,"allowed_ips":null,"auth_name":"n","auth_password":"p"}}`))
	require.NoError(t, err)
	out, err = smstrunkconfiguration.Marshal(httpOut)
	require.NoError(t, err)
	assert.JSONEq(t, `{"type":"http_out","attributes":{}}`, string(out))
}
