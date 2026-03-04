package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrdersCreate(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	order, err := server.client.Orders().Create(context.Background(), &Order{
		AllowBackOrdering: true,
		Items: []OrderItem{
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID: "acc46374-0b34-4912-9f67-8340339db1e5",
					Qty:   2,
				},
			},
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID: "f36d2812-2195-4385-85e8-e59c3484a8bc",
					Qty:   1,
				},
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

	item1 := order.Items[0]
	assert.Equal(t, "did_order_items", item1.Type)
	assert.Equal(t, 1, item1.Attributes.Qty)
	assert.Equal(t, "0.0", item1.Attributes.Nrc)
	assert.Equal(t, "5.6", item1.Attributes.Mrc)

	assertRequestJSON(t, capturedBody, "orders/create_request.json")
}

func TestOrdersCreateAvailableDid(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_available_did.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.Orders().Create(context.Background(), &Order{
		Items: []OrderItem{
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID:          "acc46374-0b34-4912-9f67-8340339db1e5",
					AvailableDidID: "c43441e3-82d4-4d84-93e2-80998576c1ce",
				},
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "orders/create_request_available_did.json")
}

func TestOrdersCreateReservation(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_reservation.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.Orders().Create(context.Background(), &Order{
		Items: []OrderItem{
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID:            "32840f64-5c3f-4278-8c8d-887fbe2f03f4",
					DidReservationID: "e3ed9f97-1058-430c-9134-38f1c614ee9f",
				},
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "orders/create_request_reservation.json")
}

func TestOrdersCreateCapacity(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_capacity.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.Orders().Create(context.Background(), &Order{
		Items: []OrderItem{
			{
				Type: "capacity_order_items",
				Attributes: OrderItemAttributes{
					CapacityPoolID: "b7522a31-4bf3-4c23-81e8-e7a14b23663f",
					Qty:            1,
				},
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "orders/create_request_capacity.json")
}

func TestOrdersCreateBillingCycles(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_billing_cycles.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	billingCycles := 5
	_, err := server.client.Orders().Create(context.Background(), &Order{
		AllowBackOrdering: true,
		Items: []OrderItem{
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID:              "f36d2812-2195-4385-85e8-e59c3484a8bc",
					Qty:                1,
					BillingCyclesCount: &billingCycles,
				},
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "orders/create_request_billing_cycles.json")
}

func TestOrdersCreateNanpa(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_nanpa.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	_, err := server.client.Orders().Create(context.Background(), &Order{
		AllowBackOrdering: true,
		Items: []OrderItem{
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID:         "fe77889c-f05a-40ad-a845-96aca3c28054",
					Qty:           1,
					NanpaPrefixID: "eeed293b-f3d8-4ef8-91ef-1b077d174b3b",
				},
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "orders/create_request_nanpa.json")
}

func TestOrdersCreateWithCallback(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders_with_callback/create.json"},
	}, func(r *http.Request) {
		capturedBody, _ = io.ReadAll(r.Body)
	})

	cbURL := "https://example.com/callback"
	cbMethod := "POST"
	_, err := server.client.Orders().Create(context.Background(), &Order{
		AllowBackOrdering: true,
		CallbackURL:       &cbURL,
		CallbackMethod:    &cbMethod,
		Items: []OrderItem{
			{
				Type: "did_order_items",
				Attributes: OrderItemAttributes{
					SkuID: "f36d2812-2195-4385-85e8-e59c3484a8bc",
					Qty:   1,
				},
			},
		},
	})
	require.NoError(t, err)

	assertRequestJSON(t, capturedBody, "orders_with_callback/create_request.json")
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
	assert.Equal(t, "generic_order_items", order.Items[0].Type)
}
