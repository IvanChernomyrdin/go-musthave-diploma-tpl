package router

import (
	"context"
	"go-musthave-diploma-tpl/internal/accrual/config"
	"go-musthave-diploma-tpl/internal/accrual/handler"
	"go-musthave-diploma-tpl/internal/accrual/middleware"
	"time"

	"github.com/go-chi/chi"
	"go.uber.org/zap"
)

func NewRouter(log *zap.SugaredLogger, ctx context.Context, r *chi.Mux, handler *handler.Handler, cfg *config.Config) {
	r.Use(middleware.LoggerMiddleware())
	r.Group(func(r chi.Router) {
		r.Use(middleware.LimitRequestsMiddleware(cfg.MaxRequests, time.Duration(cfg.Timeout)*time.Second))
		r.Get("/api/orders/{number}", handler.GetAccrualInfo(log))
	})
	r.Post("/api/goods", handler.CreateProductReward(ctx, log))
	r.Post("/api/orders", handler.RegisterNewOrder(ctx, log))
}
