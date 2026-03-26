package handler

import (
	"html/template"
	"net/http"
	"path/filepath"

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

func (h *PageHandler) Home(w http.ResponseWriter, r *http.Request) {
	products, _, _ := h.products.List(r.Context(), 0, 1, 6)
	newsItems, _, _ := h.news.List(r.Context(), 1, 3)
	categories, _ := h.categories.List(r.Context())

	h.tmpl.ExecuteTemplate(w, "home.html", map[string]any{
		"Products":   products,
		"News":       newsItems,
		"Categories": categories,
		"Title":      "Bean & Brew — Кофейный магазин",
	})
}

func (h *PageHandler) Catalog(w http.ResponseWriter, r *http.Request) {
	catID := int64(queryInt(r, "category", 0))
	page := queryInt(r, "page", 1)

	products, total, _ := h.products.List(r.Context(), catID, page, 12)
	categories, _ := h.categories.List(r.Context())

	h.tmpl.ExecuteTemplate(w, "catalog.html", map[string]any{
		"Products":       products,
		"Categories":     categories,
		"Total":          total,
		"Page":           page,
		"ActiveCategory": catID,
		"Title":          "Каталог — Bean & Brew",
	})
}

func (h *PageHandler) NewsFeed(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	items, total, _ := h.news.List(r.Context(), page, 10)

	h.tmpl.ExecuteTemplate(w, "news.html", map[string]any{
		"News":  items,
		"Total": total,
		"Page":  page,
		"Title": "Новости — Bean & Brew",
	})
}
