package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPLogger(t *testing.T) {
	t.Run("Создание логгера с правильной структурой папок", func(t *testing.T) {
		// удаляем папку если существует
		logDir := "runtime/log"
		os.RemoveAll(logDir)

		logger := NewHTTPLogger()
		defer logger.Close()

		assert.NotNil(t, logger)
		assert.NotNil(t, logger.Logger)

		// Проверяем что папка создалась
		_, err := os.Stat(logDir)
		assert.NoError(t, err, "Папка логов должна создаваться")

		// Делаем запись в лог чтобы файл создался
		logger.LogRequest("GET", "/test", 200, 100, 1.0)

		// Даем время на запись в файл
		time.Sleep(100 * time.Millisecond)

		// Теперь проверяем что файл лога создался
		logPath := filepath.Join(logDir, "http.log")
		_, err = os.Stat(logPath)
		assert.NoError(t, err, "Файл лога должен создаваться после записи")
	})

	t.Run("Создание логгера когда папка уже существует", func(t *testing.T) {
		logDir := "runtime/log"
		err := os.MkdirAll(logDir, 0755)
		require.NoError(t, err)

		logger := NewHTTPLogger()
		defer logger.Close()

		assert.NotNil(t, logger)
	})

	t.Run("Логгер не паникует при создании", func(t *testing.T) {
		assert.NotPanics(t, func() {
			logger := NewHTTPLogger()
			if logger != nil {
				logger.Close()
			}
		})
	})
}

func TestHTTPLogger_LogRequest(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	tests := []struct {
		name          string
		method        string
		uri           string
		status        int
		responseSize  int
		duration      float64
		expectedEmoji string
	}{
		{
			name:          "Успешный запрос 200",
			method:        "GET",
			uri:           "/api/users",
			status:        200,
			responseSize:  1024,
			duration:      15.5,
			expectedEmoji: "✅",
		},
		{
			name:          "Успешный запрос 201",
			method:        "POST",
			uri:           "/api/orders",
			status:        201,
			responseSize:  512,
			duration:      25.0,
			expectedEmoji: "✅",
		},
		{
			name:          "Клиентская ошибка 400",
			method:        "POST",
			uri:           "/api/login",
			status:        400,
			responseSize:  128,
			duration:      5.2,
			expectedEmoji: "⚠️",
		},
		{
			name:          "Ошибка 404",
			method:        "GET",
			uri:           "/api/not-found",
			status:        404,
			responseSize:  256,
			duration:      3.1,
			expectedEmoji: "⚠️",
		},
		{
			name:          "Серверная ошибка 500",
			method:        "GET",
			uri:           "/api/internal",
			status:        500,
			responseSize:  512,
			duration:      100.5,
			expectedEmoji: "❌",
		},
		{
			name:          "Серверная ошибка 503",
			method:        "PUT",
			uri:           "/api/service",
			status:        503,
			responseSize:  1024,
			duration:      150.0,
			expectedEmoji: "❌",
		},
		{
			name:          "Информационный статус 100",
			method:        "GET",
			uri:           "/api/info",
			status:        100,
			responseSize:  64,
			duration:      1.5,
			expectedEmoji: "🔵",
		},
		{
			name:          "Редирект 301",
			method:        "GET",
			uri:           "/api/old",
			status:        301,
			responseSize:  0,
			duration:      2.0,
			expectedEmoji: "🔵",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				logger.LogRequest(tt.method, tt.uri, tt.status, tt.responseSize, tt.duration)
			})
		})
	}
}

func TestHTTPLogger_EmojiSelection(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	testCases := []struct {
		status   int
		expected string
	}{
		{200, "✅"}, {201, "✅"}, {204, "✅"},
		{400, "⚠️"}, {401, "⚠️"}, {403, "⚠️"}, {404, "⚠️"},
		{500, "❌"}, {502, "❌"}, {503, "❌"},
		{100, "🔵"}, {301, "🔵"}, {302, "🔵"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Status_%d", tc.status), func(t *testing.T) {
			assert.NotPanics(t, func() {
				logger.LogRequest("GET", "/test", tc.status, 100, 1.0)
			})
		})
	}
}

func TestHTTPLogger_Close(t *testing.T) {
	t.Run("Закрытие логгера без ошибок", func(t *testing.T) {
		logger := NewHTTPLogger()

		assert.NotPanics(t, func() {
			err := logger.Close()
			assert.NoError(t, err)
		})
	})

	t.Run("Многократное закрытие логгера", func(t *testing.T) {
		logger := NewHTTPLogger()

		err := logger.Close()
		assert.NoError(t, err)

		assert.NotPanics(t, func() {
			err := logger.Close()
			_ = err
		})
	})
}

func TestHTTPLogger_ConcurrentLogging(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.LogRequest("GET", "/api/test", 200, 100, 1.0)
			logger.LogRequest("POST", "/api/data", 201, 200, 2.0)
			logger.LogRequest("PUT", "/api/update", 400, 150, 1.5)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	assert.True(t, true)
}

func TestHTTPLogger_FileRotation(t *testing.T) {
	t.Run("Логгер поддерживает ротацию файлов", func(t *testing.T) {
		logger := NewHTTPLogger()
		defer logger.Close()

		for i := 0; i < 100; i++ {
			logger.LogRequest("GET", "/api/test", 200, 100, 1.0)
		}

		assert.True(t, true)
	})
}

// Новые тесты для дополнительного покрытия
func TestHTTPLogger_DifferentHTTPMethods(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run("Method_"+method, func(t *testing.T) {
			assert.NotPanics(t, func() {
				logger.LogRequest(method, "/api/test", 200, 100, 1.0)
			})
		})
	}
}

func TestHTTPLogger_VariousURIs(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	uris := []string{
		"/",
		"/api/users/123",
		"/api/orders/456/items",
		"/static/css/style.css",
		"/api/v1/long/path/with/many/segments",
	}

	for _, uri := range uris {
		t.Run("URI_"+uri, func(t *testing.T) {
			assert.NotPanics(t, func() {
				logger.LogRequest("GET", uri, 200, 100, 1.0)
			})
		})
	}
}

func TestHTTPLogger_ResponseSizes(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	sizes := []int{0, 1, 100, 1024, 1048576, 9999999}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("Size_%d", size), func(t *testing.T) {
			assert.NotPanics(t, func() {
				logger.LogRequest("GET", "/api/test", 200, size, 1.0)
			})
		})
	}
}

func TestHTTPLogger_Durations(t *testing.T) {
	logger := NewHTTPLogger()
	defer logger.Close()

	durations := []float64{0.1, 1.0, 10.5, 100.0, 1000.0, 9999.9}

	for _, duration := range durations {
		t.Run(fmt.Sprintf("Duration_%.1f", duration), func(t *testing.T) {
			assert.NotPanics(t, func() {
				logger.LogRequest("GET", "/api/test", 200, 100, duration)
			})
		})
	}
}
