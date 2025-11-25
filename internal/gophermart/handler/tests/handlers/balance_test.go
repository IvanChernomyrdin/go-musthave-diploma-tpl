package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		setupContext   func(ctx context.Context) context.Context
		setupMock      func()
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Successfully get balance",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, middleware.UserIDKey, "1")
			},
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(1).Return(models.Balance{
					Current:   500.5,
					Withdrawn: 42,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"current":500.5,"withdrawn":42}`,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
			},
		},
		{
			name: "User not authenticated",
			setupContext: func(ctx context.Context) context.Context {
				return ctx // без userID
			},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"user is not authenticated"}`,
		},
		{
			name: "Database error",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, middleware.UserIDKey, "1")
			},
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(1).Return(models.Balance{}, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
		{
			name: "Zero balance",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, middleware.UserIDKey, "1")
			},
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(1).Return(models.Balance{
					Current:   0,
					Withdrawn: 0,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"current":0,"withdrawn":0}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/api/user/balance", nil)
			req = req.WithContext(tt.setupContext(req.Context()))

			rr := httptest.NewRecorder()

			h.GetBalance(rr, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			// Проверяем тело ответа
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String(), "handler returned unexpected body")
			}

			// Проверяем дополнительные условия
			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
