package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
)

func TestGetOrdersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		userID         string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
		checkJSON      bool
		expectedLen    int
	}{
		{
			name:   "Successful orders retrieval",
			userID: "1",
			mockSetup: func() {
				now := time.Now()
				mockRepo.EXPECT().GetOrders(1).
					Return([]models.Order{
						{
							Number:     "1234567890",
							Status:     "PROCESSED",
							Accrual:    100.5,
							UploadedAt: now,
						},
						{
							Number:     "0987654321",
							Status:     "NEW",
							Accrual:    0,
							UploadedAt: now,
						},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkJSON:      true,
			expectedLen:    2,
		},
		{
			name:   "No orders",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().GetOrders(1).
					Return([]models.Order{}, nil)
			},
			// хендлер возвращает 200 и пустой массив []
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			checkJSON:      true,
			expectedLen:    0,
		},
		{
			name:   "Database error",
			userID: "1",
			mockSetup: func() {
				mockRepo.EXPECT().GetOrders(1).
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			name:           handler.ErrUserIsNotAuthenticated.Error(),
			userID:         "",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   handler.ErrUserIsNotAuthenticated.Error(),
		},
		{
			name:      "Invalid userID",
			userID:    "invalid",
			mockSetup: func() {},
			// сервис вернёт ErrInvalidUserID, а хендлер отправит 500 + "internal server error"
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("GET", "/api/user/orders", nil)

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			h.GetOrders(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}

			if tt.checkJSON {
				var orders []models.Order
				if err := json.Unmarshal(rr.Body.Bytes(), &orders); err != nil {
					t.Errorf("handler returned invalid JSON: %v", err)
				}
				if len(orders) != tt.expectedLen {
					t.Errorf("expected %d orders, got %d", tt.expectedLen, len(orders))
				}
			}
		})
	}
}
