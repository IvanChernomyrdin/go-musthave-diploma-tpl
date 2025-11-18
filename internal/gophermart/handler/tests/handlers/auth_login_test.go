package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	mocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
)

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		payload        map[string]string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Успешный логин",
			payload: map[string]string{
				"login":    "testuser",
				"password": "correctpassword",
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByLoginAndPassword("testuser", "correctpassword").
					Return(&models.User{
						ID:    1,
						Login: "testuser",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Successfully logged in",
		},
		{
			name: "Неверный пароль",
			payload: map[string]string{
				"login":    "testuser",
				"password": "wrongpassword",
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByLoginAndPassword("testuser", "wrongpassword").
					Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid login or password",
		},
		{
			name: "Несуществующий пользователь",
			payload: map[string]string{
				"login":    "nonexistent",
				"password": "password",
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByLoginAndPassword("nonexistent", "password").
					Return(nil, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Invalid login or password",
		},
		{
			name: "Пустой логин",
			payload: map[string]string{
				"login":    "",
				"password": "password",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Login and password are required",
		},
		{
			name: "Ошибка базы данных",
			payload: map[string]string{
				"login":    "testuser",
				"password": "password",
			},
			mockSetup: func() {
				mockRepo.EXPECT().GetUserByLoginAndPassword("testuser", "password").
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.mockSetup()

			// Подготавливаем запрос
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			req := httptest.NewRequest("POST", "/api/user/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// Вызываем хендлер
			h.Login(rr, req)

			// Проверяем статус
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// Проверяем тело ответа
			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
			}

			// Для успешного логина проверяем куку
			if tt.expectedStatus == http.StatusOK {
				cookies := rr.Result().Cookies()
				if len(cookies) == 0 {
					t.Error("handler should set cookie on successful login")
				}
			}
		})
	}
}
