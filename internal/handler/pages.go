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
	templates  map[string]*template.Template
}

func NewPageHandler(
	products *usecase.ProductUseCase,
	categories *usecase.CategoryUseCase,
	news *usecase.NewsUseCase,
	templateDir string,
) *PageHandler {
	layout := filepath.Join(templateDir, "layout.html")
	partials := []string{
		filepath.Join(templateDir, "search_results.html"),
		filepath.Join(templateDir, "admin_news_item.html"),
	}

	pages := []string{
		"home.html", "catalog.html", "news.html",
		"cart.html", "checkout.html", "order_confirmed.html",
		"login.html", "admin_news.html",
	}

	templates := make(map[string]*template.Template)
	for _, page := range pages {
		files := append([]string{layout, filepath.Join(templateDir, page)}, partials...)
		templates[page] = template.Must(template.ParseFiles(files...))
	}

	return &PageHandler{
		products:   products,
		categories: categories,
		news:       news,
		templates:  templates,
	}
}

func (h *PageHandler) render(w http.ResponseWriter, page string, data map[string]any) {
	tmpl, ok := h.templates[page]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	h.render(w, "home.html", h.pageData(r, "Bean & Brew — Кофейный магазин", map[string]any{
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

	h.render(w, "catalog.html", h.pageData(r, "Каталог — Bean & Brew", map[string]any{
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

	h.render(w, "news.html", h.pageData(r, "Новости — Bean & Brew", map[string]any{
		"News":  items,
		"Total": total,
		"Page":  page,
	}))
}

func (h *PageHandler) Cart(w http.ResponseWriter, r *http.Request) {
	h.render(w, "cart.html", h.pageData(r, "Корзина — Bean & Brew", nil))
}

func (h *PageHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	h.render(w, "checkout.html", h.pageData(r, "Оформление — Bean & Brew", nil))
}

func (h *PageHandler) OrderConfirmation(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	h.render(w, "order_confirmed.html", h.pageData(r, "Заказ принят — Bean & Brew", map[string]any{
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
	h.templates["home.html"].ExecuteTemplate(w, "search_results", map[string]any{
		"Products": products,
	})
}

func (h *PageHandler) AdminNews(w http.ResponseWriter, r *http.Request) {
	items, total, _ := h.news.List(r.Context(), 1, 100)
	h.render(w, "admin_news.html", h.pageData(r, "Админка новостей — Bean & Brew", map[string]any{
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
	h.templates["admin_news.html"].ExecuteTemplate(w, "admin_news_item", item)
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

func (h *PageHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "login.html", h.pageData(r, "Вход — Bean & Brew", map[string]any{}))
}

func (h *PageHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" {
		h.render(w, "login.html", h.pageData(r, "Вход — Bean & Brew", map[string]any{
			"Error": "Введите имя пользователя",
		}))
		return
	}

	if username == adminUser && password != adminPass {
		h.render(w, "login.html", h.pageData(r, "Вход — Bean & Brew", map[string]any{
			"Error": "Неверный пароль для админа",
		}))
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    username,
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *PageHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
