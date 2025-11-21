package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestWithdrawalsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful withdrawals retrieval",
			userID: "1",
			mockSetup: func() {
				expectedWithdrawals := []models.WithdrawBalance{
					{
						Order:       "2377225624",
						Sum:         751,
						ProcessedAt: time.Now().Add(-24 * time.Hour),
					},
					{
						Order:       "49927398716",
						Sum:         500,
						ProcessedAt: time.Now().Add(-12 * time.Hour),
					},
				}
				mockRepo.EXPECT().Withdrawals(1).Return(expectedWithdrawals, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "No withdrawals",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().Withdrawals(1).Return([]models.WithdrawBalance{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   "User is not authenticated",
			userID: "",
			mockSetup: func() {
				// No mock expectations
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User is not authenticated",
		},
		{
			name:   "Invalid userID",
			userID: "invalid",
			mockSetup: func() {
				// No mock expectations
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user ID",
		},
		{
			name:   "Database error",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().Withdrawals(1).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.mockSetup()

			// Создаем запрос
			req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)

			// Добавляем userID в контекст если он есть
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// Вызываем хендлер
			h.Withdrawals(rr, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Проверяем тело ответа для случаев с ошибками
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}

			// Проверяем JSON ответ для успешного случая
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				var withdrawals []models.WithdrawBalance
				err := json.Unmarshal(rr.Body.Bytes(), &withdrawals)
				assert.NoError(t, err)
				assert.NotEmpty(t, withdrawals)
			}
		})
	}
}

// Test for router integration
func TestRouter_WithdrawalsRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)
	router := handler.NewRouter(h, svc)

	t.Run("GET /api/user/withdrawals - protected route", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authentication required")
	})

	t.Run("GET /api/user/withdrawals - unsupported method", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/user/withdrawals", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.NotEqual(t, http.StatusNotFound, rr.Code)
	})
}
