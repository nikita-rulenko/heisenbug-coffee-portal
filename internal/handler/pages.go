package handler

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/usecase"
)

type PageHandler struct {
	products   *usecase.ProductUseCase
	categories *usecase.CategoryUseCase
	news       *usecase.NewsUseCase
	tmpl       *template.Template
}

func NewPageHandler(
	products *usecase.ProductUseCase,
	categories *usecase.CategoryUseCase,
	news *usecase.NewsUseCase,
	templateDir string,
) *PageHandler {
	tmpl := template.Must(template.ParseGlob(filepath.Join(templateDir, "*.html")))
	return &PageHandler{
		products:   products,
		categories: categories,
		news:       news,
		tmpl:       tmpl,
	}
}

func (h *PageHandler) pageData(r *http.Request, title string, extra map[string]any) map[string]any {
	data := map[string]any{
		"Title":    title,
		"Username": GetUsername(r.Context()),
		"IsAdmin":  IsAdmin(r.Context()),
	}
	for k, v := range extra {
		data[k] = v
	}
	return data
}

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	products, _, _ := h.products.List(r.Context(), 0, 1, 6)
	newsItems, _, _ := h.news.List(r.Context(), 1, 3)
	categories, _ := h.categories.List(r.Context())

	h.tmpl.ExecuteTemplate(w, "home.html", h.pageData(r, "Bean & Brew — Кофейный магазин", map[string]any{
		"Products":   products,
		"News":       newsItems,
		"Categories": categories,
	}))
}

func (h *PageHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	catID := int64(queryInt(r, "category", 0))
	page := queryInt(r, "page", 1)

	products, total, _ := h.products.List(r.Context(), catID, page, 12)
	categories, _ := h.categories.List(r.Context())

	h.tmpl.ExecuteTemplate(w, "catalog.html", h.pageData(r, "Каталог — Bean & Brew", map[string]any{
		"Products":       products,
		"Categories":     categories,
		"Total":          total,
		"Page":           page,
		"ActiveCategory": catID,
	}))
}

func (h *PageHandler) NewsFeed(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	items, total, _ := h.news.List(r.Context(), page, 10)

	h.tmpl.ExecuteTemplate(w, "news.html", h.pageData(r, "Новости — Bean & Brew", map[string]any{
		"News":  items,
		"Total": total,
		"Page":  page,
	}))
}

func (h *PageHandler) Cart(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "cart.html", h.pageData(r, "Корзина — Bean & Brew", nil))
}

func (h *PageHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "checkout.html", h.pageData(r, "Оформление — Bean & Brew", nil))
}

func (h *PageHandler) OrderConfirmation(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	h.tmpl.ExecuteTemplate(w, "order_confirmed.html", h.pageData(r, "Заказ принят — Bean & Brew", map[string]any{
		"OrderID": orderID,
	}))
}

func (h *PageHandler) SearchFragment(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		w.Write([]byte(""))
		return
	}
	products, _ := h.products.Search(r.Context(), q, 8)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.ExecuteTemplate(w, "search_results", map[string]any{
		"Products": products,
	})
}

func (h *PageHandler) AdminNews(w http.ResponseWriter, r *http.Request) {
	items, total, _ := h.news.List(r.Context(), 1, 100)
	h.tmpl.ExecuteTemplate(w, "admin_news.html", h.pageData(r, "Админка новостей — Bean & Brew", map[string]any{
		"News":  items,
		"Total": total,
	}))
}

func (h *PageHandler) AdminNewsCreate(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	title := r.FormValue("title")
	content := r.FormValue("content")
	author := GetUsername(r.Context())
	if author == "" {
		author = "Админ"
	}

	item := &entity.NewsItem{
		Title:   title,
		Content: content,
		Author:  author,
	}
	if err := h.news.Create(r.Context(), item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	h.tmpl.ExecuteTemplate(w, "admin_news_item", item)
}

func (h *PageHandler) AdminNewsDelete(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	h.news.Delete(r.Context(), id)
	w.WriteHeader(http.StatusOK)
}
