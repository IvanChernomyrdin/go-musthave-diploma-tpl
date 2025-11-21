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

func TestRouter_WithdrawRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	t.Run("POST /api/user/orbalance/withdraw - protected route", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/user/balance/withdraw", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code,
			"POST /api/user/balance/withdraw must require authentication")
	})

	t.Run("Unsupported method for /api/user/balance/withdraw", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/user/balance/withdraw", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.NotEqual(t, http.StatusNotFound, rr.Code,
			"PUT /api/user/balance/withdraw should not return 404")
	})
}
