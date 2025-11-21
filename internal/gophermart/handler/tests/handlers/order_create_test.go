package tests

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
)

func TestCreateOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		userID         string
		contentType    string
		body           string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Successful order creation",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "Successful order creation",
		},
		{
			name:           "Invalid Content-Type",
			userID:         "1",
			contentType:    "application/json",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request format",
		},
		{
			name:           "User is not authenticated",
			userID:         "",
			contentType:    "text/plain",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "User is not authenticated",
		},
		{
			name:           "Empty order number",
			userID:         "1",
			contentType:    "text/plain",
			body:           "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Order number is required",
		},
		{
			name:           "Number contains non-digit characters",
			userID:         "1",
			contentType:    "text/plain",
			body:           "123abc456",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid number format",
		},
		{
			name:           "Invalid Luhn number",
			userID:         "1",
			contentType:    "text/plain",
			body:           "1234567890",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "Invalid number format",
		},
		{
			name:        "Duplicate order from same user",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(handler.ErrDuplicateOrder)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   handler.ErrDuplicateOrder.Error(),
		},
		{
			name:        "Order already uploaded by another user",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(handler.ErrOtherUserOrder)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   handler.ErrOtherUserOrder.Error(),
		},
		{
			name:        "Database error",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
		{
			name:           "Invalid userID",
			userID:         "invalid",
			contentType:    "text/plain",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Invalid user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			h.CreateOrder(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
