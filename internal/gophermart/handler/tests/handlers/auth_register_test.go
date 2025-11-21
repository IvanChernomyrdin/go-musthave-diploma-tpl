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

func TestRegisterHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockGofemartRepo(ctrl)
	svc := service.NewGofemartService(mockRepo)
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		payload        interface{}
		mockSetup      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Sucessfully register",
			payload: map[string]string{
				"login":    "newuser",
				"password": "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser("newuser", "password123").
					Return(&models.User{
						ID:    1,
						Login: "newuser",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User successfully registered",
		},
		{
			name: "Login already taken",
			payload: map[string]string{
				"login":    "existinguser",
				"password": "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser("existinguser", "password123").
					Return(nil, fmt.Errorf("login already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "login already taken",
		},
		{
			name: "Empty login",
			payload: map[string]string{
				"login":    "",
				"password": "password123",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrLoginAndPasswordRequired.Error(),
		},
		{
			name: "Empty password",
			payload: map[string]string{
				"login":    "user",
				"password": "",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrLoginAndPasswordRequired.Error(),
		},
		{
			name: "Database error",
			payload: map[string]string{
				"login":    "testuser",
				"password": "password123",
			},
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser("testuser", "password123").
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			name:           "Неверный JSON",
			payload:        "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrInvalidJSONFormat.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.mockSetup()

			// Подготавливаем тело запроса
			var body []byte
			var err error

			switch v := tt.payload.(type) {
			case string:
				body = []byte(v)
			default:
				body, err = json.Marshal(v)
				if err != nil {
					t.Fatalf("Failed to marshal payload: %v", err)
				}
			}

			// Создаем запрос
			req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			// Вызываем хендлер
			h.Register(rr, req)

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
