package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/handler"
	sqliteRepo "github.com/nikita-rulenko/heisenbug-portal/internal/repository/sqlite"
	"github.com/nikita-rulenko/heisenbug-portal/internal/usecase"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	db, err := sqliteRepo.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	sqliteRepo.RunMigrations(db)
	t.Cleanup(func() { db.Close() })

	productRepo := sqliteRepo.NewProductRepo(db)
	categoryRepo := sqliteRepo.NewCategoryRepo(db)
	newsRepo := sqliteRepo.NewNewsRepo(db)
	orderRepo := sqliteRepo.NewOrderRepo(db)

	productUC := usecase.NewProductUseCase(productRepo, categoryRepo)
	categoryUC := usecase.NewCategoryUseCase(categoryRepo)
	newsUC := usecase.NewNewsUseCase(newsRepo)
	orderUC := usecase.NewOrderUseCase(orderRepo, productRepo)

	productH := handler.NewProductHandler(productUC)
	categoryH := handler.NewCategoryHandler(categoryUC)
	newsH := handler.NewNewsHandler(newsUC)
	orderH := handler.NewOrderHandler(orderUC)

	r := chi.NewRouter()
	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/products", productH.Routes())
		r.Mount("/categories", categoryH.Routes())
		r.Mount("/news", newsH.Routes())
		r.Mount("/orders", orderH.Routes())
	})

	return httptest.NewServer(r)
}

func postJSON(url string, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	return http.Post(url, "application/json", bytes.NewReader(b))
}

func TestAPICategoryCRUD(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := postJSON(srv.URL+"/api/v1/categories", map[string]any{
		"name": "Эспрессо", "slug": "espresso", "description": "Напитки",
	})
	if err != nil {
		t.Fatalf("POST category: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201", resp.StatusCode)
	}

	var cat entity.Category
	json.NewDecoder(resp.Body).Decode(&cat)
	resp.Body.Close()
	if cat.ID == 0 {
		t.Fatal("expected non-zero category ID")
	}

	resp, err = http.Get(srv.URL + "/api/v1/categories")
	if err != nil {
		t.Fatalf("GET categories: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}

	var listResp map[string]any
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()

	cats := listResp["categories"].([]any)
	if len(cats) != 1 {
		t.Errorf("categories len = %d, want 1", len(cats))
	}
}

func TestAPIProductCRUD(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/categories", map[string]any{
		"name": "Кофе", "slug": "coffee", "description": "Кофейные напитки",
	})
	var cat entity.Category
	json.NewDecoder(resp.Body).Decode(&cat)
	resp.Body.Close()

	resp, err := postJSON(srv.URL+"/api/v1/products", map[string]any{
		"name": "Латте", "category_id": cat.ID, "price": 350, "description": "Вкусный", "in_stock": true,
	})
	if err != nil {
		t.Fatalf("POST product: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST status = %d, want 201", resp.StatusCode)
	}

	var product entity.Product
	json.NewDecoder(resp.Body).Decode(&product)
	resp.Body.Close()

	if product.Name != "Латте" {
		t.Errorf("product name = %q, want %q", product.Name, "Латте")
	}

	resp, _ = http.Get(srv.URL + "/api/v1/products")
	var listResp map[string]any
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()

	products := listResp["products"].([]any)
	if len(products) != 1 {
		t.Errorf("products len = %d, want 1", len(products))
	}
}

func TestAPIProductValidationError(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/products", map[string]any{
		"name": "", "category_id": 1, "price": 100,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty name status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = postJSON(srv.URL+"/api/v1/products", map[string]any{
		"name": "Test", "category_id": 1, "price": -10,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("negative price status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsCRUD(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "Открытие", "content": "Мы открылись!", "author": "Admin",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST news status = %d, want 201", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(srv.URL + "/api/v1/news")
	var listResp map[string]any
	json.NewDecoder(resp.Body).Decode(&listResp)
	resp.Body.Close()

	news := listResp["news"].([]any)
	if len(news) != 1 {
		t.Errorf("news len = %d, want 1", len(news))
	}
}

func TestAPIOrderFlow(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	postJSON(srv.URL+"/api/v1/categories", map[string]any{
		"name": "Напитки", "slug": "drinks",
	})

	resp, _ := postJSON(srv.URL+"/api/v1/products", map[string]any{
		"name": "Капучино", "category_id": 1, "price": 320, "in_stock": true,
	})
	var product entity.Product
	json.NewDecoder(resp.Body).Decode(&product)
	resp.Body.Close()

	resp, _ = postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "test-customer",
		"items":       []map[string]any{{"product_id": product.ID, "quantity": 2}},
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("POST order status = %d, want 201", resp.StatusCode)
	}

	var order entity.Order
	json.NewDecoder(resp.Body).Decode(&order)
	resp.Body.Close()

	if order.Total != 640 {
		t.Errorf("order total = %v, want 640", order.Total)
	}
	if order.Status != entity.OrderStatusNew {
		t.Errorf("order status = %q, want %q", order.Status, entity.OrderStatusNew)
	}
}

func TestAPIProductNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/products/99999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderEmptyItems(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("empty order status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}
