package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	serviceMocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRouter_OrdersRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	t.Run("POST /api/user/orders - защищенный маршрут", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/user/orders", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code,
			"POST /api/user/orders должен требовать аутентификации")
	})

	t.Run("GET /api/user/orders - защищенный маршрут", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/orders", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code,
			"GET /api/user/orders должен требовать аутентификации")
	})

	t.Run("Неподдерживаемый метод для /api/user/orders", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/user/orders", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.NotEqual(t, http.StatusNotFound, rr.Code,
			"PUT /api/user/orders не должен возвращать 404")
	})
}
