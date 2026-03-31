package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func putJSON(url string, body any) (*http.Response, error) {
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func doDelete(url string) (*http.Response, error) {
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	return http.DefaultClient.Do(req)
}

func decodeBody(resp *http.Response, v any) {
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(v)
}

func createCategory(t *testing.T, srv *httptest.Server, name, slug string) entity.Category {
	t.Helper()
	resp, _ := postJSON(srv.URL+"/api/v1/categories", map[string]any{
		"name": name, "slug": slug, "description": "test",
	})
	var cat entity.Category
	decodeBody(resp, &cat)
	return cat
}

func createProduct(t *testing.T, srv *httptest.Server, name string, catID int64, price float64) entity.Product {
	t.Helper()
	resp, _ := postJSON(srv.URL+"/api/v1/products", map[string]any{
		"name": name, "category_id": catID, "price": price, "in_stock": true,
	})
	var p entity.Product
	decodeBody(resp, &p)
	return p
}

// ============ Category API ============

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

func TestAPICategoryGetByID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Кофе", "coffee")

	resp, _ := http.Get(fmt.Sprintf("%s/api/v1/categories/%d", srv.URL, cat.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}
	var got entity.Category
	decodeBody(resp, &got)
	if got.Name != "Кофе" {
		t.Errorf("Name = %q, want %q", got.Name, "Кофе")
	}
}

func TestAPICategoryGetByIDNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/categories/999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPICategoryUpdate(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Old", "old")

	resp, _ := putJSON(fmt.Sprintf("%s/api/v1/categories/%d", srv.URL, cat.ID), map[string]any{
		"name": "New", "slug": "new",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/categories/%d", srv.URL, cat.ID))
	var got entity.Category
	decodeBody(resp, &got)
	if got.Name != "New" {
		t.Errorf("Name after update = %q", got.Name)
	}
}

func TestAPICategoryDelete(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Del", "del")

	resp, _ := doDelete(fmt.Sprintf("%s/api/v1/categories/%d", srv.URL, cat.ID))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/categories/%d", srv.URL, cat.ID))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete, status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPICategoryValidationEmptyName(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/categories", map[string]any{
		"name": "", "slug": "test",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPICategoryValidationEmptySlug(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/categories", map[string]any{
		"name": "Test", "slug": "",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPICategoryDeleteNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := doDelete(srv.URL + "/api/v1/categories/999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPICategoryUpdateNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := putJSON(srv.URL+"/api/v1/categories/999", map[string]any{
		"name": "X", "slug": "x",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPICategoryInvalidJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/categories", "application/json", bytes.NewReader([]byte("not json")))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ============ Product API ============

func TestAPIProductCRUD(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Кофе", "coffee")

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

func TestAPIProductGetByID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Раф", cat.ID, 400)

	resp, _ := http.Get(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}
	var got entity.Product
	decodeBody(resp, &got)
	if got.Name != "Раф" {
		t.Errorf("Name = %q", got.Name)
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

func TestAPIProductUpdate(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Кофе", "coffee")
	p := createProduct(t, srv, "Латте", cat.ID, 350)

	resp, _ := putJSON(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID), map[string]any{
		"name": "Раф", "category_id": cat.ID, "price": 400,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	var got entity.Product
	decodeBody(resp, &got)
	if got.Name != "Раф" || got.Price != 400 {
		t.Errorf("after update: name=%q price=%v", got.Name, got.Price)
	}
}

func TestAPIProductUpdateValidationError(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "X", cat.ID, 100)

	resp, _ := putJSON(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID), map[string]any{
		"name": "", "category_id": cat.ID, "price": 100,
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIProductDelete(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "Del", cat.ID, 100)

	resp, _ := doDelete(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/products/%d", srv.URL, p.ID))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete, status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIProductDeleteNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := doDelete(srv.URL + "/api/v1/products/999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIProductSearch(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	createProduct(t, srv, "Капучино", cat.ID, 320)
	createProduct(t, srv, "Латте", cat.ID, 350)

	resp, _ := http.Get(srv.URL + "/api/v1/products/search?q=Капучино")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result map[string]any
	decodeBody(resp, &result)
	products := result["products"].([]any)
	if len(products) != 1 {
		t.Errorf("search results = %d, want 1", len(products))
	}
}

func TestAPIProductSearchEmpty(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/products/search?q=")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result map[string]any
	decodeBody(resp, &result)
	products := result["products"].([]any)
	if len(products) != 0 {
		t.Errorf("empty search results = %d, want 0", len(products))
	}
}

func TestAPIProductSearchNoResults(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	createProduct(t, srv, "Кофе", cat.ID, 200)

	resp, _ := http.Get(srv.URL + "/api/v1/products/search?q=несуществующий")
	var result map[string]any
	decodeBody(resp, &result)
	products, _ := result["products"].([]any)
	if len(products) != 0 {
		t.Errorf("no-match search = %d, want 0", len(products))
	}
}

func TestAPIProductListPagination(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	for i := range 15 {
		createProduct(t, srv, fmt.Sprintf("P%d", i), cat.ID, float64(100+i))
	}

	resp, _ := http.Get(srv.URL + "/api/v1/products?page=1&page_size=5")
	var result map[string]any
	decodeBody(resp, &result)
	products := result["products"].([]any)
	if len(products) != 5 {
		t.Errorf("page 1 len = %d, want 5", len(products))
	}
	total := int(result["total"].(float64))
	if total != 15 {
		t.Errorf("total = %d, want 15", total)
	}
}

func TestAPIProductListByCategory(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat1 := createCategory(t, srv, "Кофе", "coffee")
	cat2 := createCategory(t, srv, "Чай", "tea")
	createProduct(t, srv, "Латте", cat1.ID, 350)
	createProduct(t, srv, "Американо", cat1.ID, 250)
	createProduct(t, srv, "Зелёный чай", cat2.ID, 200)

	resp, _ := http.Get(fmt.Sprintf("%s/api/v1/products?category_id=%d", srv.URL, cat1.ID))
	var result map[string]any
	decodeBody(resp, &result)
	products := result["products"].([]any)
	if len(products) != 2 {
		t.Errorf("coffee products = %d, want 2", len(products))
	}
}

func TestAPIProductInvalidJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/products", "application/json", bytes.NewReader([]byte("bad")))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIProductInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/products/abc")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIProductSearchWithLimit(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	for range 10 {
		createProduct(t, srv, "Кофе", cat.ID, 200)
	}

	resp, _ := http.Get(srv.URL + "/api/v1/products/search?q=Кофе&limit=3")
	var result map[string]any
	decodeBody(resp, &result)
	products := result["products"].([]any)
	if len(products) != 3 {
		t.Errorf("search with limit = %d, want 3", len(products))
	}
}

// ============ News API ============

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

func TestAPINewsGetByID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "Тест", "content": "Контент",
	})
	var n entity.NewsItem
	decodeBody(resp, &n)

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET status = %d, want 200", resp.StatusCode)
	}
	var got entity.NewsItem
	decodeBody(resp, &got)
	if got.Title != "Тест" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestAPINewsGetByIDNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/news/999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsUpdate(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "Old", "content": "Old content",
	})
	var n entity.NewsItem
	decodeBody(resp, &n)

	resp, _ = putJSON(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID), map[string]any{
		"title": "New", "content": "New content",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	var got entity.NewsItem
	decodeBody(resp, &got)
	if got.Title != "New" {
		t.Errorf("Title = %q", got.Title)
	}
}

func TestAPINewsUpdateValidation(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "X", "content": "Y",
	})
	var n entity.NewsItem
	decodeBody(resp, &n)

	resp, _ = putJSON(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID), map[string]any{
		"title": "", "content": "Y",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsDelete(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "Del", "content": "Del",
	})
	var n entity.NewsItem
	decodeBody(resp, &n)

	resp, _ = doDelete(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE status = %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete, status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsDeleteNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := doDelete(srv.URL + "/api/v1/news/999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsValidationEmptyTitle(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "", "content": "X",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsValidationEmptyContent(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "X", "content": "",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsInvalidJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/news", "application/json", bytes.NewReader([]byte("bad")))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsListPagination(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	for i := range 12 {
		postJSON(srv.URL+"/api/v1/news", map[string]any{
			"title": fmt.Sprintf("N%d", i), "content": "C",
		})
	}

	resp, _ := http.Get(srv.URL + "/api/v1/news?page=1&page_size=5")
	var result map[string]any
	decodeBody(resp, &result)
	news := result["news"].([]any)
	if len(news) != 5 {
		t.Errorf("page 1 len = %d, want 5", len(news))
	}
	total := int(result["total"].(float64))
	if total != 12 {
		t.Errorf("total = %d, want 12", total)
	}
}

// ============ Order API ============

func TestAPIOrderFlow(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "Напитки", "drinks")
	p := createProduct(t, srv, "Капучино", cat.ID, 320)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "test-customer",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 2}},
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

func TestAPIOrderGetByID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, order.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET order status = %d, want 200", resp.StatusCode)
	}
	var got entity.Order
	decodeBody(resp, &got)
	if got.CustomerID != "cust" {
		t.Errorf("CustomerID = %q", got.CustomerID)
	}
}

func TestAPIOrderGetByIDNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/orders/999")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderProcess(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("process status = %d, body = %s", resp.StatusCode, body)
	}
	resp.Body.Close()
}

func TestAPIOrderComplete(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	http.Post(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, order.ID), "", nil)

	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("complete status = %d, body = %s", resp.StatusCode, body)
	}
	resp.Body.Close()
}

func TestAPIOrderCancel(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/cancel", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("cancel status = %d, want 200", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderCancelAlreadyCancelled(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	http.Post(fmt.Sprintf("%s/api/v1/orders/%d/cancel", srv.URL, order.ID), "", nil)

	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/cancel", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("cancel cancelled = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderCompleteWithoutProcess(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("complete new order = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderListByCustomer(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	for range 3 {
		postJSON(srv.URL+"/api/v1/orders", map[string]any{
			"customer_id": "alice",
			"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
		})
	}
	postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "bob",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})

	resp, _ := http.Get(srv.URL + "/api/v1/orders/customer/alice")
	var result map[string]any
	decodeBody(resp, &result)
	orders := result["orders"].([]any)
	if len(orders) != 3 {
		t.Errorf("alice orders = %d, want 3", len(orders))
	}
}

func TestAPIOrderEmptyCustomerID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "",
		"items":       []map[string]any{{"product_id": 1, "quantity": 1}},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderInvalidJSON(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/orders", "application/json", bytes.NewReader([]byte("bad")))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderMultipleItems(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p1 := createProduct(t, srv, "Латте", cat.ID, 350)
	p2 := createProduct(t, srv, "Раф", cat.ID, 400)

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items": []map[string]any{
			{"product_id": p1.ID, "quantity": 2},
			{"product_id": p2.ID, "quantity": 1},
		},
	})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d, body = %s", resp.StatusCode, body)
	}
	var order entity.Order
	decodeBody(resp, &order)
	expected := 350.0*2 + 400
	if order.Total != expected {
		t.Errorf("total = %v, want %v", order.Total, expected)
	}
}

func TestAPIOrderProcessNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/orders/999/process", "", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderCancelNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/orders/999/cancel", "", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderCompleteNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Post(srv.URL+"/api/v1/orders/999/complete", "", nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPIOrderFullLifecycleViaAPI(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	cat := createCategory(t, srv, "K", "k")
	p := createProduct(t, srv, "P", cat.ID, 100)

	// Create
	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "lifecycle",
		"items":       []map[string]any{{"product_id": p.ID, "quantity": 1}},
	})
	var order entity.Order
	decodeBody(resp, &order)

	// Process
	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/process", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("process: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Complete
	resp, _ = http.Post(fmt.Sprintf("%s/api/v1/orders/%d/complete", srv.URL, order.ID), "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("complete: %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify final state
	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/orders/%d", srv.URL, order.ID))
	var got entity.Order
	decodeBody(resp, &got)
	if got.Status != entity.OrderStatusCompleted {
		t.Errorf("final status = %q, want completed", got.Status)
	}
}

func TestAPIOrderInvalidProduct(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/orders", map[string]any{
		"customer_id": "cust",
		"items":       []map[string]any{{"product_id": 99999, "quantity": 1}},
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid product status = %d, want 400", resp.StatusCode)
	}
	resp.Body.Close()
}

// ============ Content-Type ============

func TestAPIResponseContentType(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/categories")
	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	resp.Body.Close()
}

func TestAPIErrorResponseFormat(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := http.Get(srv.URL + "/api/v1/products/999")
	var errResp map[string]string
	decodeBody(resp, &errResp)
	if _, ok := errResp["error"]; !ok {
		t.Error("error response should have 'error' key")
	}
}
