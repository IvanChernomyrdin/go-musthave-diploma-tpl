package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	serviceMocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		setupContext   func(ctx context.Context) context.Context
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Успешное получение баланса",
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
			expectedBody:   `{"current":500.5,"withdrawn":42}` + "\n",
		},
		{
			name: "User is not authenticated",
			setupContext: func(ctx context.Context) context.Context {
				return ctx // без userID
			},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User is not authenticated\n",
		},
		{
			name: "Ошибка сервиса",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, middleware.UserIDKey, "1")
			},
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(1).Return(models.Balance{}, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error\n",
		},
		{
			name: "Неверный userID в контексте",
			setupContext: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, middleware.UserIDKey, "invalid")
			},
			setupMock:      func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user ID\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/api/user/balance", nil)
			req = req.WithContext(tt.setupContext(req.Context()))

			rr := httptest.NewRecorder()

			h.GetBalance(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, tt.expectedBody, rr.Body.String())

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var balance models.Balance
				err := json.Unmarshal(rr.Body.Bytes(), &balance)
				assert.NoError(t, err)
			}
		})
	}
}
