package router

import (
	"go-musthave-diploma-tpl/internal/accrual/config"
	"go-musthave-diploma-tpl/internal/accrual/handler"
	"go-musthave-diploma-tpl/internal/accrual/middleware"
	"time"

	"github.com/go-chi/chi"
)

func NewRouter(r *chi.Mux, handler *handler.Handler, cfg config.Config) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.LimitRequestsMiddleware(cfg.MaxRequests, time.Duration(cfg.Timeout)*time.Second))
		r.Get("/api/orders/{number}", handler.GetAccrualInfo())
	})
	r.Post("/api/goods", handler.CreateProductReward())
	r.Post("/api/orders", handler.RegisterNewOrder())
}
