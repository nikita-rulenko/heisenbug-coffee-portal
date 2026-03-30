package handler

import (
	"context"
	"net/http"
)

type ctxKey string

const (
	usernameKey ctxKey = "username"
	isAdminKey  ctxKey = "isAdmin"
	cookieName         = "bb_session"
	adminUser          = "admin"
	adminPass          = "admin"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var username string
		var isAdmin bool
		if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
			username = c.Value
			isAdmin = username == adminUser
		}
		ctx := context.WithValue(r.Context(), usernameKey, username)
		ctx = context.WithValue(ctx, isAdminKey, isAdmin)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !IsAdmin(r.Context()) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GetUsername(ctx context.Context) string {
	if v, ok := ctx.Value(usernameKey).(string); ok {
		return v
	}
	return ""
}

func IsAdmin(ctx context.Context) bool {
	if v, ok := ctx.Value(isAdminKey).(bool); ok {
		return v
	}
	return false
}

func (h *PageHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	h.tmpl.ExecuteTemplate(w, "login.html", h.pageData(r, "Вход — Bean & Brew", map[string]any{}))
}

func (h *PageHandler) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" {
		h.tmpl.ExecuteTemplate(w, "login.html", h.pageData(r, "Вход — Bean & Brew", map[string]any{
			"Error": "Введите имя пользователя",
		}))
		return
	}

	if username == adminUser && password != adminPass {
		h.tmpl.ExecuteTemplate(w, "login.html", h.pageData(r, "Вход — Bean & Brew", map[string]any{
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
