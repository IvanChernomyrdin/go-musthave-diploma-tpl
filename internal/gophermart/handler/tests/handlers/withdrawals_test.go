package tests

import (
	"bytes"
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

func TestHandler_Withdrawals(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(mockRepo *mocks.MockGofemartRepo)
		expectedStatus int
		expectedBody   string
		checkJSON      bool
	}{
		{
			name:   "Успешное получение списка списаний",
			userID: "1",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				expectedWithdrawals := []models.WithdrawBalance{
					{
						Order:       "2377225624",
						Sum:         751.50,
						ProcessedAt: time.Now().Add(-24 * time.Hour),
					},
					{
						Order:       "49927398716",
						Sum:         500.25,
						ProcessedAt: time.Now().Add(-12 * time.Hour),
					},
				}
				mockRepo.EXPECT().Withdrawals(1).Return(expectedWithdrawals, nil)
			},
			expectedStatus: http.StatusOK,
			checkJSON:      true,
		},
		{
			name:   "Нет списаний",
			userID: "1",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdrawals(1).Return([]models.WithdrawBalance{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Пользователь не аутентифицирован",
			userID:         "",
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   handler.ErrUserIsNotAuthenticated.Error(),
		},
		{
			name:   "Неверный userID",
			userID: "invalid",
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				// userIDint будет 0
				mockRepo.EXPECT().Withdrawals(0).Return(nil, assert.AnError)
			},
			// хендлер мапит любую ошибку в 500 + "internal server error"
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			name:   "Ошибка базы данных",
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
			// отдельный контроллер и мок на каждый под-тест
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

			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}

			if tt.checkJSON {
				var withdrawals []models.WithdrawBalance
				err := json.Unmarshal(rr.Body.Bytes(), &withdrawals)
				assert.NoError(t, err)
				assert.Len(t, withdrawals, 2)
				assert.Equal(t, "2377225624", withdrawals[0].Order)
				assert.Equal(t, 751.50, withdrawals[0].Sum)
				assert.Equal(t, "49927398716", withdrawals[1].Order)
				assert.Equal(t, 500.25, withdrawals[1].Sum)
			}
		})
	}
}
