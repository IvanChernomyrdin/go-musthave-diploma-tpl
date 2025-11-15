package httpserver

import (
	service "go-musthave-diploma-tpl/internal/service"
)

type Handler struct {
	svc service.GofemartRepo
}

func NewHandler(svc service.GofemartRepo) *Handler { return &Handler{svc: svc} }
