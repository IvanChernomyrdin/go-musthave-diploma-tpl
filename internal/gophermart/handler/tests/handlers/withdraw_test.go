package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
)

func TestWithdrawHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*mocks.MockGofemartRepo)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Successful withdrawal",
			userID: "1",
			requestBody: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdraw(1, models.WithdrawBalance{
					Order: "2377225624",
					Sum:   751,
				}).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Withdrawal successful",
		},
		{
			name:   "Invalid order number",
			userID: "1",
			requestBody: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdraw(1, models.WithdrawBalance{
					Order: "2377225624",
					Sum:   751,
				}).Return(handler.ErrInvalidOrderNumber)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid order number",
		},
		{
			name:   "Insufficient funds",
			userID: "1",
			requestBody: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdraw(1, models.WithdrawBalance{
					Order: "2377225624",
					Sum:   751,
				}).Return(handler.ErrLackOfFunds)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   "Lack of funds",
		},
		{
			name:   "Internal server error",
			userID: "1",
			requestBody: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			mockSetup: func(mockRepo *mocks.MockGofemartRepo) {
				mockRepo.EXPECT().Withdraw(1, models.WithdrawBalance{
					Order: "2377225624",
					Sum:   751,
				}).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
		{
			name:           "User is not authenticated",
			userID:         "",
			requestBody:    models.WithdrawBalance{Order: "2377225624", Sum: 751},
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User is not authenticated",
		},
		{
			name:           "Invalid userID",
			userID:         "invalid",
			requestBody:    models.WithdrawBalance{Order: "2377225624", Sum: 751},
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user ID",
		},
		{
			name:           "Invalid Content-Type",
			userID:         "1",
			requestBody:    models.WithdrawBalance{Order: "2377225624", Sum: 751},
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request format",
		},
		{
			name:           "Empty order number",
			userID:         "1",
			requestBody:    models.WithdrawBalance{Order: "", Sum: 751},
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Order number is required",
		},
		{
			name:           "Sum less than or equal to 0",
			userID:         "1",
			requestBody:    models.WithdrawBalance{Order: "2377225624", Sum: 0},
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Sum must be positive",
		},
		{
			name:           "Invalid JSON",
			userID:         "1",
			requestBody:    "{invalid json",
			mockSetup:      func(mockRepo *mocks.MockGofemartRepo) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create new controller for each test
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockGofemartRepo(ctrl)
			svc := service.NewGofemartService(mockRepo)
			h := handler.NewHandler(svc)

			// Setup mock
			tt.mockSetup(mockRepo)

			// Prepare request body
			var bodyBytes []byte
			switch v := tt.requestBody.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			// Create request
			req := httptest.NewRequest("POST", "/api/user/balance/withdraw", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			if tt.name == "Invalid Content-Type" {
				req.Header.Set("Content-Type", "text/plain") // неправильный тип
			} else {
				req.Header.Set("Content-Type", "application/json") // правильный тип
			}

			// Add userID to context if exists
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// Call handler
			h.Withdraw(rr, req)

			// Check status
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Check response body
			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
