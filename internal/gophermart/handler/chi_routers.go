package httpserver

import (
	"net/http"

	middleware "go-musthave-diploma-tpl/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func NewRouter(h *Handler) http.Handler {
	r := chi.NewRouter()

	// логгер запросов
	r.Use(middleware.LoggerMiddleware())

	r.Route("/", func(r chi.Router) {})

	return r
}
