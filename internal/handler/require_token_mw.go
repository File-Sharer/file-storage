package handler

import (
	"net/http"
	"os"
)

func (h *Handler) RequireTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Internal-Token")
		if token != os.Getenv("X_INTERNAL_TOKEN") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
