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
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(mockRepo *mocks.MockGofemartRepo)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful withdrawals retrieval",
			userID: "1",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
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
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdrawals(1).Return([]models.WithdrawBalance{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:   handler.ErrUserIsNotAuthenticated.Error(),
			userID: "",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				// репозиторий не вызывается
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   handler.ErrUserIsNotAuthenticated.Error(),
		},
		{
			name:   "Invalid userID",
			userID: "invalid",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				// userIDint будет 0 из-за strconv.Atoi("invalid")
				mockRepo.EXPECT().Withdrawals(0).Return(nil, assert.AnError)
			},
			// хендлер любую ошибку мапит в 500 + "internal server error"
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			name:   "Database error",
			userID: "1",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdrawals(1).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// отдельный контроллер и мок для каждого кейса
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockGofemartRepo(ctrl)
			svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
			h := handler.NewHandler(svc)

			tt.mockSetup(mockRepo)

			req := httptest.NewRequest("GET", "/api/user/withdrawals", nil)

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			h.Withdrawals(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}

			// успешный JSON-ответ проверяем только для 200 OK
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
