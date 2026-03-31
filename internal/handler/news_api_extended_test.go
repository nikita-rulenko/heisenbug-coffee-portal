package handler_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func createNewsItem(t *testing.T, srv *httptest.Server, title, content string) entity.NewsItem {
	t.Helper()
	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": title, "content": content, "author": "test",
	})
	var n entity.NewsItem
	decodeBody(resp, &n)
	return n
}

func TestAPINewsCreateValidationCases(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	tests := []struct {
		name string
		body any
		want int
	}{
		{"empty title", map[string]any{"title": "", "content": "X"}, http.StatusBadRequest},
		{"empty content", map[string]any{"title": "X", "content": ""}, http.StatusBadRequest},
		{"both empty", map[string]any{"title": "", "content": ""}, http.StatusBadRequest},
		{"missing title", map[string]any{"content": "X"}, http.StatusBadRequest},
		{"missing content", map[string]any{"title": "X"}, http.StatusBadRequest},
		{"valid minimal", map[string]any{"title": "Новость", "content": "Текст"}, http.StatusCreated},
		{"with author", map[string]any{"title": "Тест", "content": "Контент", "author": "Автор"}, http.StatusCreated},
		{"unicode title", map[string]any{"title": "ニュース", "content": "内容"}, http.StatusCreated},
		{"long title", map[string]any{"title": "Очень длинный заголовок новости для кофейного портала тестирование", "content": "C"}, http.StatusCreated},
		{"emoji title", map[string]any{"title": "☕ Открытие!", "content": "Мы открылись"}, http.StatusCreated},
		{"long content", map[string]any{"title": "T", "content": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore."}, http.StatusCreated},
		{"special chars", map[string]any{"title": "Кофе & Чай <br>", "content": "A&B"}, http.StatusCreated},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := postJSON(srv.URL+"/api/v1/news", tt.body)
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

func TestAPINewsUpdateInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := putJSON(srv.URL+"/api/v1/news/abc", map[string]any{"title": "X", "content": "Y"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPINewsDeleteInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := doDelete(srv.URL + "/api/v1/news/abc")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPINewsGetInvalidID(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/api/v1/news/abc")
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", resp.StatusCode)
	}
}

func TestAPINewsUpdateNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := putJSON(srv.URL+"/api/v1/news/99999", map[string]any{"title": "X", "content": "Y"})
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
}

func TestAPINewsUpdateValidationCases(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	n := createNewsItem(t, srv, "Тест", "Контент")

	tests := []struct {
		name string
		body any
		want int
	}{
		{"empty title", map[string]any{"title": "", "content": "X"}, http.StatusBadRequest},
		{"empty content", map[string]any{"title": "X", "content": ""}, http.StatusBadRequest},
		{"valid", map[string]any{"title": "Обновлено", "content": "Новый контент"}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := putJSON(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID), tt.body)
			defer resp.Body.Close()
			if resp.StatusCode != tt.want {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("status = %d, want %d, body: %s", resp.StatusCode, tt.want, body)
			}
		})
	}
}

func TestAPINewsListEmpty(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	resp, _ := http.Get(srv.URL + "/api/v1/news")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestAPINewsListPaginationEdgeCases(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	for i := range 20 {
		postJSON(srv.URL+"/api/v1/news", map[string]any{
			"title": fmt.Sprintf("News%d", i), "content": "Content",
		})
	}

	tests := []struct {
		name    string
		query   string
		wantMin int
		wantMax int
	}{
		{"page_size=5", "?page_size=5", 5, 5},
		{"page=2 size=5", "?page=2&page_size=5", 5, 5},
		{"page=5 size=5", "?page=5&page_size=5", 0, 0},
		{"page_size=1", "?page_size=1", 1, 1},
		{"no params", "", 10, 20},
		{"page=1 size=20", "?page=1&page_size=20", 20, 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, _ := http.Get(srv.URL + "/api/v1/news" + tt.query)
			var result map[string]any
			decodeBody(resp, &result)
			news, _ := result["news"].([]any)
			if len(news) < tt.wantMin || len(news) > tt.wantMax {
				t.Errorf("count = %d, want %d-%d", len(news), tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestAPINewsCRUDFullCycle(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// Create
	n := createNewsItem(t, srv, "Открытие", "Мы открылись!")
	if n.ID == 0 {
		t.Fatal("expected non-zero ID")
	}

	// Read
	resp, _ := http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET = %d", resp.StatusCode)
	}
	var got entity.NewsItem
	decodeBody(resp, &got)
	if got.Title != "Открытие" {
		t.Errorf("title = %q, want Открытие", got.Title)
	}

	// Update
	resp, _ = putJSON(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID), map[string]any{
		"title": "Обновление", "content": "Новый текст", "author": "Admin",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify update
	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	decodeBody(resp, &got)
	if got.Title != "Обновление" || got.Content != "Новый текст" {
		t.Errorf("after update: title=%q content=%q", got.Title, got.Content)
	}

	// Delete
	resp, _ = doDelete(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("DELETE = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Verify gone
	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("after delete GET = %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAPINewsCreateWithAuthor(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, _ := postJSON(srv.URL+"/api/v1/news", map[string]any{
		"title": "Тест", "content": "Контент", "author": "Иванов",
	})
	var n entity.NewsItem
	decodeBody(resp, &n)

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	var got entity.NewsItem
	decodeBody(resp, &got)
	if got.Author != "Иванов" {
		t.Errorf("author = %q, want Иванов", got.Author)
	}
}

func TestAPINewsDeleteAndVerifyList(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	n1 := createNewsItem(t, srv, "Первая", "Текст 1")
	createNewsItem(t, srv, "Вторая", "Текст 2")

	doDelete(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n1.ID))

	resp, _ := http.Get(srv.URL + "/api/v1/news")
	var result map[string]any
	decodeBody(resp, &result)
	news := result["news"].([]any)
	if len(news) != 1 {
		t.Errorf("after delete count = %d, want 1", len(news))
	}
}

func TestAPINewsCreateMultipleAndCount(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	for i := range 15 {
		postJSON(srv.URL+"/api/v1/news", map[string]any{
			"title": fmt.Sprintf("News%d", i), "content": "C",
		})
	}

	resp, _ := http.Get(srv.URL + "/api/v1/news")
	var result map[string]any
	decodeBody(resp, &result)
	total := int(result["total"].(float64))
	if total != 15 {
		t.Errorf("total = %d, want 15", total)
	}
}

func TestAPINewsUpdateAllFields(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()
	n := createNewsItem(t, srv, "Old", "Old content")

	resp, _ := putJSON(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID), map[string]any{
		"title": "New", "content": "New content", "author": "New author",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PUT = %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp, _ = http.Get(fmt.Sprintf("%s/api/v1/news/%d", srv.URL, n.ID))
	var got entity.NewsItem
	decodeBody(resp, &got)
	if got.Title != "New" || got.Content != "New content" || got.Author != "New author" {
		t.Errorf("not updated: %+v", got)
	}
}

func TestAPINewsListVerifiesTotal(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	for i := range 8 {
		postJSON(srv.URL+"/api/v1/news", map[string]any{
			"title": fmt.Sprintf("N%d", i), "content": "C",
		})
	}

	resp, _ := http.Get(srv.URL + "/api/v1/news?page=1&page_size=3")
	var result map[string]any
	decodeBody(resp, &result)
	news := result["news"].([]any)
	if len(news) != 3 {
		t.Errorf("items = %d, want 3", len(news))
	}
	total := int(result["total"].(float64))
	if total != 8 {
		t.Errorf("total = %d, want 8", total)
	}
}
