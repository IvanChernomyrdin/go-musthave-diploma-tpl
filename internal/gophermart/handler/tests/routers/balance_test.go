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

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	// Создаем роутер БЕЗ middleware аутентификации для тестов
	r := chi.NewRouter()
	r.Get("/api/user/balance", h.GetBalance)

	tests := []struct {
		name           string
		setupMock      func()
		expectedStatus int
		expectedBody   models.Balance
	}{
		{
			name: "Успешное получение баланса",
			setupMock: func() {
				mockRepo.EXPECT().GetBalance(1).Return(models.Balance{
					Current:   500.5,
					Withdrawn: 42,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: models.Balance{
				Current:   500.5,
				Withdrawn: 42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			req := httptest.NewRequest("GET", "/api/user/balance", nil)
			// Добавляем userID в контекст как это делает middleware
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var balance models.Balance
				err := json.NewDecoder(rr.Body).Decode(&balance)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, balance)
			}
		})
	}
}
