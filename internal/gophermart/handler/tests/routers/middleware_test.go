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

func TestRouter_MiddlewareChain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	t.Run("Logger middleware connected", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/register", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Protected routes require authentication", func(t *testing.T) {
		protectedRoutes := []string{
			"/api/user/orders",
			"/api/user/balance",
			"/api/user/balance/withdraw",
		}

		for _, route := range protectedRoutes {
			t.Run("Route: "+route, func(t *testing.T) {
				req := httptest.NewRequest("GET", route, nil)
				rr := httptest.NewRecorder()

				router.ServeHTTP(rr, req)

				assert.Equal(t, http.StatusUnauthorized, rr.Code,
					"For protected route %s expected status 401", route)
			})
		}
	})
}
