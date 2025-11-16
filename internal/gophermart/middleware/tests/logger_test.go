package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middlewareDir "go-musthave-diploma-tpl/internal/gophermart/middleware"
)

func TestLoggerMiddleware_Integration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		handler        http.HandlerFunc
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "Успешный GET запрос",
			method: "GET",
			path:   "/test",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name:   "Запрос с ошибкой 404",
			method: "GET",
			path:   "/not-found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not Found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Not Found",
		},
		{
			name:   "POST запрос с созданием ресурса",
			method: "POST",
			path:   "/api/resource",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte("created"))
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "created",
		},
		{
			name:   "Запрос без явного статуса",
			method: "GET",
			path:   "/default",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Default OK"))
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "Default OK",
		},
		{
			name:   "Пустой ответ 204",
			method: "DELETE",
			path:   "/resource/1",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ПОДГОТОВКА
			middleware := middlewareDir.LoggerMiddleware()
			handler := middleware(tt.handler)
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rr := httptest.NewRecorder()

			// ДЕЙСТВИЕ
			handler.ServeHTTP(rr, req)

			// ПРОВЕРКА
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("Expected body '%s', got '%s'", tt.expectedBody, rr.Body.String())
			}
		})
	}
}

func TestLoggerMiddleware_Concurrent(t *testing.T) {
	middleware := middlewareDir.LoggerMiddleware()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			req := httptest.NewRequest("GET", "/concurrent", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Goroutine %d: Expected status 200, got %d", id, rr.Code)
			}
			if rr.Body.String() != "OK" {
				t.Errorf("Goroutine %d: Expected body 'OK', got '%s'", id, rr.Body.String())
			}

			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestLoggerMiddleware_DifferentPaths(t *testing.T) {
	middleware := middlewareDir.LoggerMiddleware()

	testCases := []struct {
		path string
	}{
		{"/"},
		{"/api/users"},
		{"/api/orders/123"},
		{"/static/css/style.css"},
		{"/api/v1/long/path/with/multiple/segments"},
	}

	for _, tc := range testCases {
		t.Run("Path: "+tc.path, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			}))

			req := httptest.NewRequest("GET", tc.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("For path %s: Expected status 200, got %d", tc.path, rr.Code)
			}
		})
	}
}

func TestLoggerMiddleware_WithDelay(t *testing.T) {
	middleware := middlewareDir.LoggerMiddleware()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Delayed response"))
	}))

	req := httptest.NewRequest("GET", "/slow", nil)
	rr := httptest.NewRecorder()

	start := time.Now()
	handler.ServeHTTP(rr, req)
	duration := time.Since(start)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	if duration < 10*time.Millisecond {
		t.Errorf("Expected duration > 10ms, got %v", duration)
	}
}

func TestLoggerMiddleware_DifferentMethods(t *testing.T) {
	middleware := middlewareDir.LoggerMiddleware()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run("Method: "+method, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(method + " OK"))
			}))

			req := httptest.NewRequest(method, "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("For method %s: Expected status 200, got %d", method, rr.Code)
			}
		})
	}
}

func TestLoggerMiddleware_ResponseSize(t *testing.T) {
	middleware := middlewareDir.LoggerMiddleware()

	testCases := []struct {
		name         string
		responseBody string
		expectedSize int
	}{
		{"Empty response", "", 0},
		{"Short response", "OK", 2},
		{"Long response", "This is a longer response body", 30},
		{"Unicode response", "Привет мир!", 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tc.responseBody))
			}))

			req := httptest.NewRequest("GET", "/size-test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rr.Code)
			}

			if len(rr.Body.Bytes()) != tc.expectedSize {
				t.Errorf("Expected body size %d, got %d", tc.expectedSize, len(rr.Body.Bytes()))
			}
		})
	}
}
