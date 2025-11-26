package handler

import (
	"bytes"
	"encoding/json"
	"go-musthave-diploma-tpl/internal/accrual/models"
	"go-musthave-diploma-tpl/internal/accrual/storage"
	"net/http"
	"strconv"
)

//go:generate mockgen -source=handler.go -destination=mocks/mock.go
type Service interface {
	CreateProductReward(match string, reward float64, rewardType string) error
	RegisterNewOrder(order models.Order) (bool, error)
	GetAccrualInfo(order int64) (string, float64, bool, error)
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

func (h *Handler) RegisterNewOrder() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		var buf bytes.Buffer
		var Order models.Order
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(buf.Bytes(), &Order); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var exist bool
		if exist, err = h.service.RegisterNewOrder(Order); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exist {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) GetAccrualInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		n := r.URL.Query().Get("number")
		var info models.AccrualInfo
		order, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		info.Order = order
		var exist bool

		info.Status, info.Accrual, exist, err = h.service.GetAccrualInfo(order)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !exist {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		response, err := json.Marshal(info)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)

	}
}
