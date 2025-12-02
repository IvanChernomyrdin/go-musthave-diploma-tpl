package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"go-musthave-diploma-tpl/internal/accrual/models"
	"go-musthave-diploma-tpl/internal/accrual/storage"
	"net/http"
	"strconv"
	"strings"

	"go.uber.org/zap"
)

//go:generate mockgen -source=handler.go -destination=mocks/mock.go
type Service interface {
	CreateProductReward(ctx context.Context, match string, reward float64, rewardType string) error
	RegisterNewOrder(ctx context.Context, order models.Order) (bool, error)
	GetAccrualInfo(order int64) (string, float64, bool, error)
}

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateProductReward(ctx context.Context, log *zap.SugaredLogger) http.HandlerFunc {
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
		if err := h.service.CreateProductReward(ctx, ProductReward.Match, ProductReward.Reward, ProductReward.RewardType); err != nil {
			if err == storage.ErrKeyExists {
				w.WriteHeader(http.StatusConflict)
				return
			}
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) RegisterNewOrder(ctx context.Context, log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		var buf bytes.Buffer
		var order models.Order
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(buf.Bytes(), &order); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var exist bool
		if exist, err = h.service.RegisterNewOrder(ctx, order); err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if exist {
			w.WriteHeader(http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *Handler) GetAccrualInfo(log *zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		n := pathParts[3]
		order, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		var info models.AccrualInfo
		info.Order = order
		var exist bool

		info.Status, info.Accrual, exist, err = h.service.GetAccrualInfo(order)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !exist {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		response, err := json.Marshal(info)
		if err != nil {
			log.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(response)

	}
}
