package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	middlewareDir "go-musthave-diploma-tpl/internal/gophermart/middleware"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	serviceMocks "go-musthave-diploma-tpl/internal/gophermart/service/mocks"

	"github.com/golang/mock/gomock"
)

func TestCookieMiddleware_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	gofemartService := service.NewGofemartService(mockRepo)

	mockRepo.EXPECT().
		GetUserByID(123).
		Return(&models.User{ID: 123, Login: "testuser"}, nil)

	middleware := middlewareDir.AccessCookieMiddleware(gofemartService)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := middlewareDir.GetUserID(r.Context())
		if userID != "123" {
			t.Errorf("expected userID 123, got %s", userID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	encrypted, _ := middlewareDir.Encrypt("123")
	req.AddCookie(&http.Cookie{Name: "userID", Value: encrypted})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestCookieMiddleware_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	gofemartService := service.NewGofemartService(mockRepo)

	mockRepo.EXPECT().
		GetUserByID(999).
		Return(nil, nil)

	middleware := middlewareDir.AccessCookieMiddleware(gofemartService)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when user not found")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	encrypted, _ := middlewareDir.Encrypt("999")
	req.AddCookie(&http.Cookie{Name: "userID", Value: encrypted})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestCookieMiddleware_NoCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	gofemartService := service.NewGofemartService(mockRepo)

	middleware := middlewareDir.AccessCookieMiddleware(gofemartService)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when no cookie")
	}))

	req := httptest.NewRequest("GET", "/", nil)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestCookieMiddleware_InvalidEncryption(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := serviceMocks.NewMockGofemartRepo(ctrl)
	gofemartService := service.NewGofemartService(mockRepo)

	middleware := middlewareDir.AccessCookieMiddleware(gofemartService)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid encryption")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "userID", Value: "invalid-encrypted-data"})

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestSetEncryptedCookie(t *testing.T) {
	rr := httptest.NewRecorder()
	defer rr.Result().Body.Close()

	middlewareDir.SetEncryptedCookie(rr, "123")

	if rr.Header().Get("Set-Cookie") == "" {
		t.Error("cookie should be set in response header")
	}
}

func TestGetUserID(t *testing.T) {
	t.Run("Successfully retrieved userID from context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), middlewareDir.UserIDKey, "123")

		userID, err := middlewareDir.GetUserID(ctx)

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if userID != "123" {
			t.Errorf("expected userID '123', got '%s'", userID)
		}
	})

	t.Run("Error when userID is missing from context", func(t *testing.T) {
		ctx := context.Background()

		_, err := middlewareDir.GetUserID(ctx)

		if err == nil {
			t.Error("expected error when userID not in context")
		}
	})
}
