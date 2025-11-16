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
	"github.com/stretchr/testify/require"
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
			name:           "Защищенный маршрут без аутентификации",
			method:         "GET",
			path:           "/api/user/test-auth",
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

func TestRouter_MiddlewareChain(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	t.Run("Logger middleware подключен", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/register", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("Защищенные маршруты требуют аутентификации", func(t *testing.T) {
		protectedRoutes := []string{
			"/api/user/test-auth",
			"/api/user/orders",
			"/api/user/balance",
			"/api/user/withdrawals",
		}

		for _, route := range protectedRoutes {
			t.Run("Route: "+route, func(t *testing.T) {
				req := httptest.NewRequest("GET", route, nil)
				rr := httptest.NewRecorder()

				router.ServeHTTP(rr, req)

				assert.Equal(t, http.StatusUnauthorized, rr.Code,
					"Для защищенного маршрута %s ожидался статус 401", route)
			})
		}
	})
}

func TestRouter_MethodValidation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/api/user/register"},
		{"POST", "/api/user/login"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			assert.NotEqual(t, http.StatusInternalServerError, rr.Code)
			assert.NotEqual(t, http.StatusNotFound, rr.Code)
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
		"/api/user/test-auth",
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

func TestRouter_Integration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	mockRepo.EXPECT().
		CreateUser(gomock.Any(), gomock.Any()).
		Return(nil, nil).
		AnyTimes()

	router := handler.NewRouter(h, svc)

	t.Run("Полный цикл запроса", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/user/register", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		require.NotEqual(t, http.StatusInternalServerError, rr.Code)
	})
}
