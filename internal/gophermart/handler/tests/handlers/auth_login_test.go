package tests

import (
	"bytes"
	"database/sql"
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
	svc := service.NewGofemartService(mockRepo, "http://localhost:8081")
	h := handler.NewHandler(svc)

	tests := []struct {
		name           string
		payload        interface{}
		contentType    string
		mockSetup      func()
		expectedStatus int
		expectedBody   string
		checkResponse  func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "Successful login",
			payload: models.RegisterRequest{
				Login:    "testuser",
				Password: "correctpassword",
			},
			contentType: "application/json",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLoginAndPassword("testuser", "correctpassword").
					Return(&models.User{
						ID:    1,
						Login: "testuser",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				cookieHeader := rr.Header().Get("Set-Cookie")
				if cookieHeader == "" {
					t.Fatalf("expected Set-Cookie header to be set")
				}
				if !bytes.Contains([]byte(cookieHeader), []byte("userID")) {
					t.Fatalf("expected Set-Cookie to contain userID, got %s", cookieHeader)
				}
			},
		},
		{
			name: "Invalid login or password (wrong password)",
			payload: models.RegisterRequest{
				Login:    "testuser",
				Password: "wrongpassword",
			},
			contentType: "application/json",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLoginAndPassword("testuser", "wrongpassword").
					Return(nil, sql.ErrNoRows)
			},
			// СЕРВИС возвращает ошибку, которую хендлер трактует как internal error
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			// аналогично
			name: "Invalid login or password (nonexistent user)",
			payload: models.RegisterRequest{
				Login:    "nonexistent",
				Password: "password",
			},
			contentType: "application/json",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLoginAndPassword("nonexistent", "password").
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
		{
			name: "Empty login",
			payload: models.RegisterRequest{
				Login:    "",
				Password: "password",
			},
			contentType:    "application/json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrLoginAndPasswordRequired.Error(),
		},
		{
			name: "Empty password",
			payload: models.RegisterRequest{
				Login:    "testuser",
				Password: "",
			},
			contentType:    "application/json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrLoginAndPasswordRequired.Error(),
		},
		{
			name:           "Invalid JSON",
			payload:        `{"login": "testuser", "password":`, // обрубленный JSON
			contentType:    "application/json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   handler.ErrInvalidJSONFormat.Error(),
		},
		{
			name: "Wrong content type",
			payload: models.RegisterRequest{
				Login:    "testuser",
				Password: "password",
			},
			contentType:    "text/plain",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "content-type must be application/json",
		},
		{
			name: "Database error",
			payload: models.RegisterRequest{
				Login:    "testuser",
				Password: "password",
			},
			contentType: "application/json",
			mockSetup: func() {
				mockRepo.EXPECT().
					GetUserByLoginAndPassword("testuser", "password").
					Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   handler.ErrInternalServerError.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch p := tt.payload.(type) {
			case string:
				bodyBytes = []byte(p)
			default:
				var err error
				bodyBytes, err = json.Marshal(p)
				if err != nil {
					t.Fatalf("failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			h.Login(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedBody != "" && !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBody)) {
				t.Errorf("handler returned unexpected body: got %q want to contain %q", rr.Body.String(), tt.expectedBody)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
