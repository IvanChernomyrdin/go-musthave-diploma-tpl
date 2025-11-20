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

func TestRouter_Routes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Регистрация пользователя",
			method:         "POST",
			path:           "/api/user/register",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Логин пользователя",
			method:         "POST",
			path:           "/api/user/login",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Создание заказа без аутентификации",
			method:         "POST",
			path:           "/api/user/orders",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Получение заказов без аутентификации",
			method:         "GET",
			path:           "/api/user/orders",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Запрос на списание средств",
			method:         "POST",
			path:           "/api/user/balance/withdraw",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Несуществующий маршрут",
			method:         "GET",
			path:           "/api/not-found",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code,
				"Для пути %s ожидался статус %d, получен %d",
				tt.path, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRouter_RouteStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	routes := []string{
		"/api/user/register",
		"/api/user/login",
		"/api/user/orders",
	}

	for _, route := range routes {
		t.Run("Route exists: "+route, func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.NotEqual(t, http.StatusNotFound, rr.Code,
				"Маршрут %s не должен возвращать 404", route)
		})
	}
}
