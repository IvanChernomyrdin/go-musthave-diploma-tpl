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
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
	h := handler.NewHandler(svc)

	router := handler.NewRouter(h, svc)

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Registration",
			method:         "POST",
			path:           "/api/user/register",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Login",
			method:         "POST",
			path:           "/api/user/login",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Create order without authentication",
			method:         "POST",
			path:           "/api/user/orders",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Get order without authentication",
			method:         "GET",
			path:           "/api/user/orders",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Request for withdrawal of funds",
			method:         "POST",
			path:           "/api/user/balance/withdraw",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Non-existent route",
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
				"Status %s expected for path %d, got %d",
				tt.path, tt.expectedStatus, rr.Code)
		})
	}
}

func TestRouter_RouteStructure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
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
				"Route %s should not return 404", route)
		})
	}
}
