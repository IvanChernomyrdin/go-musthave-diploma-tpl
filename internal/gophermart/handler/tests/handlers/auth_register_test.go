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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_Register(t *testing.T) {
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
			name: "Successful registration",
			payload: map[string]string{
				"login":    "newuser",
				"password": "password123",
			},
			contentType: "application/json",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser("newuser", "password123").
					Return(&models.User{
						ID:    1,
						Login: "newuser",
					}, nil)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Contains(t, rr.Header().Get("Set-Cookie"), "userID")
			},
		},
		{
			name: "Login already exists",
			payload: map[string]string{
				"login":    "existinguser",
				"password": "password123",
			},
			contentType: "application/json",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser("existinguser", "password123").
					Return(nil, fmt.Errorf("login already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"login already taken"}`,
		},
		{
			name: "Empty login and password",
			payload: map[string]string{
				"login":    "",
				"password": "",
			},
			contentType:    "application/json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"login and password are required"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			var bodyBytes []byte
			switch payload := tt.payload.(type) {
			case string:
				bodyBytes = []byte(payload)
			default:
				var err error
				bodyBytes, err = json.Marshal(payload)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/api/user/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()

			h.Register(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, rr.Body.String(), "handler returned unexpected body")
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}
