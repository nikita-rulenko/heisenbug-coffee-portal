package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nikita-rulenko/heisenbug-portal/internal/entity"
)

func paramInt64(r *http.Request, name string) (int64, error) {
	s := chi.URLParam(r, name)
	return strconv.ParseInt(s, 10, 64)
}

func queryInt(r *http.Request, name string, defaultVal int) int {
	s := r.URL.Query().Get(name)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, entity.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, entity.ErrEmptyName),
		errors.Is(err, entity.ErrEmptySlug),
		errors.Is(err, entity.ErrEmptyContent),
		errors.Is(err, entity.ErrNegativePrice),
		errors.Is(err, entity.ErrInvalidCategory),
		errors.Is(err, entity.ErrInvalidProduct),
		errors.Is(err, entity.ErrInvalidQuantity),
		errors.Is(err, entity.ErrEmptyCustomerID),
		errors.Is(err, entity.ErrEmptyOrder),
		errors.Is(err, entity.ErrInvalidStatus):
		status = http.StatusBadRequest
	case errors.Is(err, entity.ErrAlreadyExists):
		status = http.StatusConflict
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
