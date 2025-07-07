package handler

import (
	"net/http"
	"os"

	"github.com/File-Sharer/file-storage/internal/service"
)

type Handler struct {
	services *service.Service
}

func New(services *service.Service) *Handler {
	return &Handler{
		services: services,
	}
}

func (h *Handler) Init() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("X-Internal-Token")
		if os.Getenv("INTERNAL_TOKEN") != token {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodPost:
			h.upload(w, r)
		case http.MethodDelete:
			h.delete(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	publicDir := "public/"
	mux.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir(publicDir))))

	return h.RequireTokenMiddleware(mux)
}
