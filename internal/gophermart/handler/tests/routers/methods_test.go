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
		{"POST", "/api/user/orders"},
		{"GET", "/api/user/orders"},
		{"GET", "/api/user/balance"},
		{"POST", "/api/user/balance/withdraw"},
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
