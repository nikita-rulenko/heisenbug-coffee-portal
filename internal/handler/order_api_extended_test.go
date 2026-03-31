package handler_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func TestAPIOrderCreateValidationCases(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)

	tests := []struct {
		name string
		body any
		want int
	}{
		{"empty customer", map[string]any{"customer_id": "", "items": []any{map[string]any{"product_id": p.ID, "quantity": 1}}}, http.StatusBadRequest},
		{"no items", map[string]any{"customer_id": "cust-1", "items": []any{}}, http.StatusBadRequest},
		{"zero quantity", map[string]any{"customer_id": "cust-1", "items": []any{map[string]any{"product_id": p.ID, "quantity": 0}}}, http.StatusBadRequest},
		{"negative quantity", map[string]any{"customer_id": "cust-1", "items": []any{map[string]any{"product_id": p.ID, "quantity": -1}}}, http.StatusBadRequest},
		{"invalid productID", map[string]any{"customer_id": "cust-1", "items": []any{map[string]any{"product_id": 0, "quantity": 1}}}, http.StatusBadRequest},
		{"product 99999", map[string]any{"customer_id": "cust-1", "items": []any{map[string]any{"product_id": 99999, "quantity": 1}}}, http.StatusBadRequest},
		{"valid order", map[string]any{"customer_id": "cust-1", "items": []any{map[string]any{"product_id": p.ID, "quantity": 2}}}, http.StatusCreated},
		{"multiple items", map[string]any{"customer_id": "cust-1", "items": []any{map[string]any{"product_id": p.ID, "quantity": 1}, map[string]any{"product_id": p.ID, "quantity": 3}}}, http.StatusCreated},
		{"unicode customer", map[string]any{"customer_id": "клиент-42", "items": []any{map[string]any{"product_id": p.ID, "quantity": 1}}}, http.StatusCreated},
		{"missing customer", map[string]any{"items": []any{map[string]any{"product_id": p.ID, "quantity": 1}}}, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := postJSON(srv.URL+"/api/v1/orders", tt.body)
			if err != nil {
				t.Fatalf("POST: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.want {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d, body: %s", resp.StatusCode, tt.want, body)
			}
		})
	}
}

func createTestOrderAPI(t *testing.T, srv *httptest.Server, customerID string, productID int64, qty int) entity.Order {
	t.Helper()
	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": customerID,
		"items":       []any{map[string]any{"product_id": productID, "quantity": qty}},
	})
	var o entity.Order
	decodeBody(resp, &o)
	return o
}

func TestAPIOrderProcessInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := postJSON(srv.URL+"/api/v1/orders/abc/process", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderCancelInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := postJSON(srv.URL+"/api/v1/orders/abc/cancel", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderCompleteInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := postJSON(srv.URL+"/api/v1/orders/abc/complete", nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderGetInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/api/v1/orders/abc")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderStatusAfterProcess(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-1", p.ID, 1)

	resp, _ := postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, o.ID))
	var got entity.Order
	decodeBody(resp, &got)
	if got.Status != entity.OrderStatusProcessing {
		t.Errorf("status = %q, want processing", got.Status)
	}
}

func TestAPIOrderStatusAfterComplete(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-1", p.ID, 1)

	postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	resp, _ := postJSON(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, o.ID), nil)
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, o.ID))
	var got entity.Order
	decodeBody(resp, &got)
	if got.Status != entity.OrderStatusCompleted {
		t.Errorf("status = %q, want completed", got.Status)
	}
}

func TestAPIOrderStatusAfterCancel(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-1", p.ID, 1)

	resp, _ := postJSON(fmt.Sprintf("%s/api/v1/orders/%d/cancel", srv.URL, o.ID), nil)
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, o.ID))
	var got entity.Order
	decodeBody(resp, &got)
	if got.Status != entity.OrderStatusCancelled {
		t.Errorf("status = %q, want cancelled", got.Status)
	}
}

func TestAPIOrderDoubleProcess(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-1", p.ID, 1)

	postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	resp, _ := postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("double process = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderProcessCompleted(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-1", p.ID, 1)

	postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	postJSON(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, o.ID), nil)
	resp, _ := postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("process completed = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderCancelCompleted(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-1", p.ID, 1)

	postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	postJSON(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, o.ID), nil)
	resp, _ := postJSON(fmt.Sprintf("%s/api/v1/orders/%d/cancel", srv.URL, o.ID), nil)
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("cancel completed = %d, want 400", resp.StatusCode)
	}
}

func TestAPIOrderCreateVerifiesTotal(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)

	tests := []struct {
		name  string
		qty   int
		total float64
	}{
		{"qty 1", 1, 350},
		{"qty 3", 3, 1050},
		{"qty 5", 5, 1750},
		{"qty 10", 10, 3500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
				"customer_id": "cust-total",
				"items":       []any{map[string]any{"product_id": p.ID, "quantity": tt.qty}},
			})
			var o map[string]any
			json.NewDecoder(resp.Body).Decode(&o)
			resp.Body.Close()
			got, _ := o["total"].(float64)
			if got != tt.total {
				t.Errorf("total = %v, want %v", got, tt.total)
			}
		})
	}
}

func TestAPIOrderCRUDFullCycle(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	o := createTestOrderAPI(t, srv, "cust-lifecycle", p.ID, 2)

	// GET
	resp, _ := http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, o.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Process
	resp, _ = postJSON(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, o.ID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("process = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Complete
	resp, _ = postJSON(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, o.ID), nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("complete = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify final status
	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, o.ID))
	var final entity.Order
	decodeBody(resp, &final)
	if final.Status != entity.OrderStatusCompleted {
		t.Errorf("final status = %q, want completed", final.Status)
	}
}
