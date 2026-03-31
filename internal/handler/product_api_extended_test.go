package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestAPIProductCreateValidationCases(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")

	tests := []struct {
		name string
		body any
		want int
	}{
		{"empty name", map[string]any{"name": "", "price": 100, "category_id": cat.ID}, http.StatusBadRequest},
		{"negative price", map[string]any{"name": "Латте", "price": -1, "category_id": cat.ID}, http.StatusBadRequest},
		{"zero categoryID", map[string]any{"name": "Латте", "price": 100, "category_id": 0}, http.StatusBadRequest},
		{"category 999", map[string]any{"name": "Латте", "price": 100, "category_id": 999}, http.StatusBadRequest},
		{"very long name", map[string]any{"name": "Супер Мега Ультра Экстра Двойной Латте с Карамелью и Корицей", "price": 100, "category_id": cat.ID}, http.StatusCreated},
		{"unicode name", map[string]any{"name": "抹茶ラテ", "price": 500, "category_id": cat.ID}, http.StatusCreated},
		{"zero price valid", map[string]any{"name": "Бесплатный", "price": 0, "category_id": cat.ID}, http.StatusCreated},
		{"with description", map[string]any{"name": "Раф", "price": 400, "category_id": cat.ID, "description": "Сливки"}, http.StatusCreated},
		{"emoji name", map[string]any{"name": "☕ Кофе", "price": 200, "category_id": cat.ID}, http.StatusCreated},
		{"missing name", map[string]any{"price": 100, "category_id": cat.ID}, http.StatusBadRequest},
		{"negative categoryID", map[string]any{"name": "Тест", "price": 100, "category_id": -1}, http.StatusBadRequest},
		{"large price", map[string]any{"name": "Премиум", "price": 999999, "category_id": cat.ID}, http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := postJSON(srv.URL+"/api/v1/products", tt.body)
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

func TestAPIProductUpdateInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := putJSON(srv.URL+"/api/v1/products/abc", map[string]any{"name": "X", "price": 1, "category_id": 1})
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPIProductUpdateNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	resp, _ := putJSON(srv.URL+"/api/v1/products/99999", map[string]any{"name": "X", "price": 1, "category_id": cat.ID})
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAPIProductDeleteInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := doDelete(srv.URL + "/api/v1/products/abc")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPIProductListQueryParams(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	for i := range 15 {
		createProduct(t, srv, "P"+string(rune('A'+i%26)), cat.ID, float64(100+i))
	}

	tests := []struct {
		name    string
		query   string
		wantMin int
		wantMax int
	}{
		{"page_size=5", "?page_size=5", 5, 5},
		{"page=2 size=5", "?page=2&page_size=5", 5, 5},
		{"page=4 size=5", "?page=4&page_size=5", 0, 0},
		{"page_size=1", "?page_size=1", 1, 1},
		{"no params", "", 15, 20},
		{"category filter", "?category_id=" + fmt.Sprintf("%d", cat.ID), 15, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := http.Get(srv.URL + "/api/v1/products" + tt.query)
			var result map[string]any
			decodeBody(resp, &result)
			products, _ := result["products"].([]any)
			if len(products) < tt.wantMin || len(products) > tt.wantMax {
				t.Errorf("count = %d, want %d-%d", len(products), tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestAPIProductListEmptyDB(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/api/v1/products")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAPIProductSearchQueryParams(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	createProduct(t, srv, "Латте", cat.ID, 350)
	createProduct(t, srv, "Капучино", cat.ID, 300)

	tests := []struct {
		name string
		query string
	}{
		{"search Латте", "?q=Латте"},
		{"search nonexistent", "?q=NonExistent"},
		{"with limit", "?q=Латте&limit=1"},
		{"empty query", "?q="},
		{"special chars", "?q=%25"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(srv.URL + "/api/v1/products/search" + tt.query)
			if err != nil {
				t.Fatalf("GET: %v", err)
			}
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("status = %d, want 200", resp.StatusCode)
			}
		})
	}
}

func TestAPIProductCRUDFullCycle(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")

	// Create
	p := createProduct(t, srv, "Латте", cat.ID, 350)
	if p.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	// Read
	resp, _ := http.Get(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Update
	resp, _ = putJSON(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID), map[string]any{
		"name": "Латте Обновлённый", "price": 400, "category_id": cat.ID,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Delete
	resp, _ = doDelete(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify gone
	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete GET = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIProductCreateDuplicate(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	cat := createCategory(t, srv, "Кофе", "coffee")
	createProduct(t, srv, "Латте", cat.ID, 350)

	resp, _ := postJSON(srv.URL+"/api/v1/products", map[string]any{
		"name": "Латте", "price": 400, "category_id": cat.ID,
	})
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("duplicate name = %d, want 201", resp.StatusCode)
	}
}
