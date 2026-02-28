package didww

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/didww/didww-api-3-go-sdk/resource/enums"
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.ID != "5da18706-be9f-49b0-aeec-0480aacd49ad" {
		t.Errorf("expected ID '5da18706-be9f-49b0-aeec-0480aacd49ad', got %q", order.ID)
	}
	if order.Amount != "5.98" {
		t.Errorf("expected Amount '5.98', got %q", order.Amount)
	}
	if order.Status != enums.OrderStatusPending {
		t.Errorf("expected Status 'Pending', got %q", order.Status)
	}
	if order.Description != "DID" {
		t.Errorf("expected Description 'DID', got %q", order.Description)
	}
	if order.Reference != "JXK-923618" {
		t.Errorf("expected Reference 'JXK-923618', got %q", order.Reference)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}

	item1 := order.Items[0]
	if item1.Type != "did_order_items" {
		t.Errorf("expected item type 'did_order_items', got %q", item1.Type)
	}
	if item1.Attributes.Qty != 1 {
		t.Errorf("expected Qty 1, got %d", item1.Attributes.Qty)
	}
	if item1.Attributes.Nrc != "0.0" {
		t.Errorf("expected Nrc '0.0', got %q", item1.Attributes.Nrc)
	}
	if item1.Attributes.Mrc != "5.6" {
		t.Errorf("expected Mrc '5.6', got %q", item1.Attributes.Mrc)
	}

	assertRequestJSON(t, capturedBody, "orders/create_request.json")
}

func TestOrdersCreateAvailableDid(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_1.json"},
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "orders/create_request_available_did.json")
}

func TestOrdersCreateReservation(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_3.json"},
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "orders/create_request_reservation.json")
}

func TestOrdersCreateCapacity(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_2.json"},
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "orders/create_request_capacity.json")
}

func TestOrdersCreateBillingCycles(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_5.json"},
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "orders/create_request_billing_cycles.json")
}

func TestOrdersCreateNanpa(t *testing.T) {
	var capturedBody []byte
	server := newTestServerWithInspector(t, map[string]testRoute{
		"POST /v3/orders": {status: http.StatusCreated, fixture: "orders/create_6.json"},
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertRequestJSON(t, capturedBody, "orders_with_callback/create_request.json")
}

func TestOrdersFind(t *testing.T) {
	_, client := newTestServer(t, map[string]testRoute{
		"GET /v3/orders/9df11dac-9d83-448c-8866-19c998be33db": {status: http.StatusOK, fixture: "orders/show.json"},
	})

	order, err := client.Orders().Find(context.Background(), "9df11dac-9d83-448c-8866-19c998be33db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.ID != "9df11dac-9d83-448c-8866-19c998be33db" {
		t.Errorf("expected ID '9df11dac-9d83-448c-8866-19c998be33db', got %q", order.ID)
	}
	if order.Amount != "25.07" {
		t.Errorf("expected Amount '25.07', got %q", order.Amount)
	}
	if order.Status != enums.OrderStatusCompleted {
		t.Errorf("expected Status 'Completed', got %q", order.Status)
	}
	if order.Description != "Payment processing fee" {
		t.Errorf("expected Description 'Payment processing fee', got %q", order.Description)
	}
	if order.Reference != "SPT-474057" {
		t.Errorf("expected Reference 'SPT-474057', got %q", order.Reference)
	}
	if len(order.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(order.Items))
	}
	if order.Items[0].Type != "generic_order_items" {
		t.Errorf("expected item type 'generic_order_items', got %q", order.Items[0].Type)
	}
}
