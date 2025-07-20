package handler

import (
	"net/http"

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
		switch r.Method {
		case http.MethodPost:
			h.upload(w, r)
		case http.MethodDelete:
			h.delete(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	filesDir := "files/"
	mux.Handle("/files/", http.StripPrefix("/files/", http.FileServer(http.Dir(filesDir))))

	return h.RequireTokenMiddleware(mux)
}
