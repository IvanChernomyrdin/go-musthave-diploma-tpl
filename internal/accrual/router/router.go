package router

import (
	"go-musthave-diploma-tpl/internal/accrual/handler"
	"go-musthave-diploma-tpl/internal/gophermart/config"

	"github.com/go-chi/chi"
)

func NewRouter(r *chi.Mux, handler *handler.Handler, cfg config.Config) {

	r.Post("/api/goods", handler.CreateProductReward())

}
