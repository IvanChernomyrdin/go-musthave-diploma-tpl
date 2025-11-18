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
	"go-musthave-diploma-tpl/internal/gophermart/models"
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
			name:        "Успешное создание заказа",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903", // Валидный номер по алгоритму Луна
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
			expectedBody:   "успешное создание заказа",
		},
		{
			name:           "Неверный Content-Type",
			userID:         "1",
			contentType:    "application/json",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "неверный формат запроса",
		},
		{
			name:           "Пользователь не аутентифицирован",
			userID:         "",
			contentType:    "text/plain",
			body:           "12345678903",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "пользователь не аутентифицирован",
		},
		{
			name:           "Пустой номер заказа",
			userID:         "1",
			contentType:    "text/plain",
			body:           "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Order number is required",
		},
		{
			name:           "Номер содержит не только цифры",
			userID:         "1",
			contentType:    "text/plain",
			body:           "123abc456",
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "неверный формат номера",
		},
		{
			name:           "Невалидный номер по алгоритму Луна",
			userID:         "1",
			contentType:    "text/plain",
			body:           "1234567890", // Невалидный номер
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "неверный формат номера",
		},
		{
			name:        "Дубликат заказа от того же пользователя",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(models.ErrDuplicateOrder)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   models.ErrDuplicateOrder.Error(),
		},
		{
			name:        "Заказ уже загружен другим пользователем",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(models.ErrOtherUserOrder)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   models.ErrOtherUserOrder.Error(),
		},
		{
			name:        "Ошибка базы данных",
			userID:      "1",
			contentType: "text/plain",
			body:        "12345678903",
			mockSetup: func() {
				mockRepo.EXPECT().CreateOrder(1, "12345678903").
					Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "внутренняя ошибка сервера",
		},
		{
			name:           "Неверный userID",
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
			// Настраиваем мок
			tt.mockSetup()

			// Создаем запрос
			req := httptest.NewRequest("POST", "/api/user/orders", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)

			// Добавляем userID в контекст если он есть
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()

			// Вызываем хендлер
			h.CreateOrder(rr, req)

			// Проверяем статус
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Проверяем тело ответа
			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
