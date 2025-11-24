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
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
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
			expectedBody:   "content-type must be text/plain",
		},
		{
			name:           handler.ErrUserIsNotAuthenticated.Error(),
			userID:         "",
			contentType:    "text/plain",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   handler.ErrUserIsNotAuthenticated.Error(),
		},
		{
			name:           "Empty order number",
			userID:         "1",
			contentType:    "text/plain",
			body:           "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrOrderNumberRequired.Error(),
		},
		{
			name:           "Number contains non-digit characters",
			userID:         "1",
			contentType:    "text/plain",
			body:           "123abc456",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   handler.ErrInvalidOrderNumber.Error(),
		},
		{
			name:           "Invalid Luhn number",
			userID:         "1",
			contentType:    "text/plain",
			body:           "1234567890",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   handler.ErrInvalidOrderNumber.Error(),
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
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			name:           "Invalid userID",
			userID:         "invalid",
			contentType:    "text/plain",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInvalidUserID.Error(),
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
