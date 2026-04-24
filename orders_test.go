package didww

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource"
	"github.com/didww/didww-api-3-go-sdk/resource/enums"
	"github.com/didww/didww-api-3-go-sdk/resource/orderitem"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdersCreate(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create.json"},
	})

	order, err := server.client.Orders().Create(context.Background(), &resource.Order{
		AllowBackOrdering: true,
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID: "acc46374-0b34-4912-9f67-8340339db1e5",
				Qty:   2,
			},
			&orderitem.DidOrderItem{
				SkuID: "f36d2812-2195-4385-85e8-e59c3484a8bc",
				Qty:   1,
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "5da18706-be9f-49b0-aeec-0480aacd49ad", order.ID)
	assert.Equal(t, "5.98", order.Amount)
	assert.Equal(t, enums.OrderStatusPending, order.Status)
	assert.Equal(t, "DID", order.Description)
	assert.Equal(t, "JXK-923618", order.Reference)
	require.Len(t, order.Items, 2)

	item1, ok := order.Items[0].(*orderitem.DidOrderItem)
	require.True(t, ok, "expected DidOrderItem")
	assert.Equal(t, 1, item1.Qty)
	assert.Equal(t, "0.0", item1.Nrc)
	assert.Equal(t, "5.6", item1.Mrc)
	assert.Equal(t, false, item1.ProratedMrc)
	assert.Nil(t, item1.BilledFrom)
	assert.Nil(t, item1.BilledTo)
	assert.Equal(t, "0.0", item1.SetupPrice)
	assert.Equal(t, "5.6", item1.MonthlyPrice)
	assert.Equal(t, "899f0119-b4e9-47d0-9b2c-8a9f282fcbe2", item1.DIDGroupID)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request.json")
}

func TestOrdersCreateAvailableDid(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_available_did.json"},
	})

	_, err := server.client.Orders().Create(context.Background(), &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.AvailableDidOrderItem{
				DidOrderItem: orderitem.DidOrderItem{
					SkuID: "acc46374-0b34-4912-9f67-8340339db1e5",
				},
				AvailableDidID: "c43441e3-82d4-4d84-93e2-80998576c1ce",
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request_available_did.json")
}

func TestOrdersCreateReservation(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_reservation.json"},
	})

	_, err := server.client.Orders().Create(context.Background(), &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.ReservationDidOrderItem{
				DidOrderItem: orderitem.DidOrderItem{
					SkuID: "32840f64-5c3f-4278-8c8d-887fbe2f03f4",
				},
				DidReservationID: "e3ed9f97-1058-430c-9134-38f1c614ee9f",
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request_reservation.json")
}

func TestOrdersCreateCapacity(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_capacity.json"},
	})

	order, err := server.client.Orders().Create(context.Background(), &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.CapacityOrderItem{
				CapacityPoolID: "b7522a31-4bf3-4c23-81e8-e7a14b23663f",
				Qty:            1,
			},
		},
	})
	require.NoError(t, err)

	require.Len(t, order.Items, 1)
	capItem, ok := order.Items[0].(*orderitem.CapacityOrderItem)
	require.True(t, ok, "expected CapacityOrderItem")
	assert.Equal(t, 1, capItem.Qty)
	assert.Equal(t, "25.0", capItem.Nrc)
	assert.Equal(t, "19.35", capItem.Mrc)
	assert.Equal(t, true, capItem.ProratedMrc)
	capBilledFrom := "2018-12-28"
	capBilledTo := "2019-01-20"
	assert.Equal(t, &capBilledFrom, capItem.BilledFrom)
	assert.Equal(t, &capBilledTo, capItem.BilledTo)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request_capacity.json")
}

func TestOrdersCreateBillingCycles(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_billing_cycles.json"},
	})

	billingCycles := 5
	_, err := server.client.Orders().Create(context.Background(), &resource.Order{
		AllowBackOrdering: true,
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID:              "f36d2812-2195-4385-85e8-e59c3484a8bc",
				Qty:                1,
				BillingCyclesCount: &billingCycles,
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request_billing_cycles.json")
}

func TestOrdersCreateNanpa(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_nanpa.json"},
	})

	_, err := server.client.Orders().Create(context.Background(), &resource.Order{
		AllowBackOrdering: true,
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID:         "fe77889c-f05a-40ad-a845-96aca3c28054",
				Qty:           1,
				NanpaPrefixID: "eeed293b-f3d8-4ef8-91ef-1b077d174b3b",
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request_nanpa.json")
}

func TestOrdersCreateEmergency(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_emergency.json"},
	})

	order, err := server.client.Orders().Create(context.Background(), &resource.Order{
		Items: []orderitem.OrderItem{
			&orderitem.EmergencyOrderItem{
				EmergencyCallingServiceID: "b6d9d793-578d-42d3-bc33-73dd8155e615",
				Qty:                       1,
			},
		},
	})
	require.NoError(t, err)

	assert.Equal(t, "a1b2c3d4-e5f6-7890-abcd-ef1234567890", order.ID)
	assert.Equal(t, "30.0", order.Amount)
	assert.Equal(t, enums.OrderStatusPending, order.Status)
	assert.Equal(t, "Emergency", order.Description)
	assert.Equal(t, "EMG-100001", order.Reference)
	require.Len(t, order.Items, 1)

	emItem, ok := order.Items[0].(*orderitem.EmergencyOrderItem)
	require.True(t, ok, "expected EmergencyOrderItem")
	assert.Equal(t, 1, emItem.Qty)
	assert.Equal(t, "5.0", emItem.Nrc)
	assert.Equal(t, "25.0", emItem.Mrc)
	assert.Equal(t, false, emItem.ProratedMrc)
	assert.Equal(t, "b6d9d793-578d-42d3-bc33-73dd8155e615", emItem.EmergencyCallingServiceID)

	assertRequestJSON(t, *capturedBodyPtr, "orders/create_request_emergency.json")
}

func TestOrdersCreateWithCallback(t *testing.T) {
	server, capturedBodyPtr := captureRequestBody(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders_with_callback/create.json"},
	})

	cbURL := "https://example.com/callback"
	cbMethod := "post"
	_, err := server.client.Orders().Create(context.Background(), &resource.Order{
		AllowBackOrdering: true,
		CallbackURL:       &cbURL,
		CallbackMethod:    &cbMethod,
		Items: []orderitem.OrderItem{
			&orderitem.DidOrderItem{
				SkuID: "f36d2812-2195-4385-85e8-e59c3484a8bc",
				Qty:   1,
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, *capturedBodyPtr, "orders_with_callback/create_request.json")
}

func TestOrdersFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/orders/9df11dac-9d83-448c-8866-19c998be33db": {status: http.StatusOK, fixture: "orders/show.json"},
	})

	order, err := client.Orders().Find(context.Background(), "9df11dac-9d83-448c-8866-19c998be33db")
	require.NoError(t, err)

	assert.Equal(t, "9df11dac-9d83-448c-8866-19c998be33db", order.ID)
	assert.Equal(t, "25.07", order.Amount)
	assert.Equal(t, enums.OrderStatusCompleted, order.Status)
	assert.Equal(t, "Payment processing fee", order.Description)
	assert.Equal(t, "SPT-474057", order.Reference)
	require.Len(t, order.Items, 1)
	genItem, ok := order.Items[0].(*orderitem.GenericOrderItem)
	require.True(t, ok, "expected GenericOrderItem")
	assert.Equal(t, 1, genItem.Qty)
	assert.Equal(t, "25.07", genItem.Nrc)
	assert.Equal(t, "0.0", genItem.Mrc)
	assert.Equal(t, false, genItem.ProratedMrc)
	billedFrom := "2018-08-17"
	billedTo := "2018-09-16"
	assert.Equal(t, &billedFrom, genItem.BilledFrom)
	assert.Equal(t, &billedTo, genItem.BilledTo)
}

func TestOrderItemMarshalExcludesReadonlyFields(t *testing.T) {
	// Simulate reusing a DidOrderItem populated from a response (with readonly fields set).
	item := &orderitem.DidOrderItem{
		BaseOrderItem: orderitem.BaseOrderItem{
			Nrc:          "0.0",
			Mrc:          "5.6",
			SetupPrice:   "0.0",
			MonthlyPrice: "5.6",
			ProratedMrc:  false,
		},
		SkuID: "acc46374-0b34-4912-9f67-8340339db1e5",
		Qty:   2,
	}
	raw, err := orderitem.MarshalItem(item)
	require.NoError(t, err)

	var envelope struct {
		Attrs json.RawMessage `json:"attributes"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope))

	var attrs map[string]any
	require.NoError(t, json.Unmarshal(envelope.Attrs, &attrs))

	// Readonly fields must not appear in the marshaled output.
	assert.NotContains(t, attrs, "nrc", "readonly field nrc should not be marshaled")
	assert.NotContains(t, attrs, "mrc", "readonly field mrc should not be marshaled")
	assert.NotContains(t, attrs, "setup_price", "readonly field setup_price should not be marshaled")
	assert.NotContains(t, attrs, "monthly_price", "readonly field monthly_price should not be marshaled")
	assert.NotContains(t, attrs, "prorated_mrc", "readonly field prorated_mrc should not be marshaled")
	assert.NotContains(t, attrs, "billed_from", "readonly field billed_from should not be marshaled")
	assert.NotContains(t, attrs, "billed_to", "readonly field billed_to should not be marshaled")

	// Writable fields must be present.
	assert.Contains(t, attrs, "sku_id")
	assert.Contains(t, attrs, "qty")
}
