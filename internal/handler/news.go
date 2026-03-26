package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
	"github.com/nikita-rulenko/heisenbug-portal/internal/usecase"
)

type NewsHandler struct {
	uc *usecase.NewsUseCase
}

func NewNewsHandler(uc *usecase.NewsUseCase) *NewsHandler {
	return &NewsHandler{uc: uc}
}

func (h *NewsHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Post("/", h.Create)
	r.Get("/{id}", h.GetByID)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h *NewsHandler) List(w http.ResponseWriter, r *http.Request) {
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "page_size", 10)

	items, total, err := h.uc.List(r.Context(), page, pageSize)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"news":  items,
		"total": total,
		"page":  page,
	})
}

func (h *NewsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var n entity.NewsItem
	if err := decodeJSON(r, &n); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.uc.Create(r.Context(), &n); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (h *NewsHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	n, err := h.uc.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *NewsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var n entity.NewsItem
	if err := decodeJSON(r, &n); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	n.ID = id
	if err := h.uc.Update(r.Context(), &n); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, n)
}

func (h *NewsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := paramInt64(r, "id")
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.uc.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
