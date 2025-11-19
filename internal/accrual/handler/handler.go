package handler

import (
	"bytes"
	"encoding/json"
	"go-musthave-diploma-tpl/internal/accrual/models"
	"go-musthave-diploma-tpl/internal/accrual/storage"
	"net/http"
)

type Service interface {
	CreateProductReward(match string, reward float64, rewardType string) error
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateProductReward() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		var buf bytes.Buffer
		var ProductReward models.ProductReward
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(buf.Bytes(), &ProductReward); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := h.service.CreateProductReward(ProductReward.Match, ProductReward.Reward, ProductReward.RewardType); err != nil {
			if err == storage.ErrKeyExists {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
