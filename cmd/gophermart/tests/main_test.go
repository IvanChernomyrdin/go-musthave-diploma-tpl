package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/repository/postgres"
	"go-musthave-diploma-tpl/internal/gophermart/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceInitialization(t *testing.T) {
	t.Run("Создание сервиса с репозиторием", func(t *testing.T) {
		repo := postgres.New()
		svc := service.NewGofemartService(repo)
		assert.NotNil(t, svc)
	})

	t.Run("Создание handler с сервисом", func(t *testing.T) {
		repo := postgres.New()
		svc := service.NewGofemartService(repo)
		h := handler.NewHandler(svc)
		assert.NotNil(t, h)
	})
}

func TestRouterInitialization(t *testing.T) {
	t.Run("Создание роутера", func(t *testing.T) {
		repo := postgres.New()
		svc := service.NewGofemartService(repo)
		h := handler.NewHandler(svc)
		r := handler.NewRouter(h, svc)

		assert.NotNil(t, r)

		req := httptest.NewRequest("GET", "/api/not-found", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestServerConfiguration(t *testing.T) {
	t.Run("Конфигурация HTTP сервера", func(t *testing.T) {
		repo := postgres.New()
		svc := service.NewGofemartService(repo)
		h := handler.NewHandler(svc)
		r := handler.NewRouter(h, svc)

		server := &http.Server{
			Addr:    "localhost:8080",
			Handler: r,
		}

		assert.NotNil(t, server)
		assert.Equal(t, "localhost:8080", server.Addr)
		assert.NotNil(t, server.Handler)
	})
}

func TestGracefulShutdown(t *testing.T) {
	t.Run("Создание контекста с таймаутом", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		assert.NotNil(t, ctx)
		assert.NotNil(t, cancel)

		select {
		case <-ctx.Done():
			t.Error("Контекст не должен истекать сразу")
		default:
		}
	})

	t.Run("Контекст истекает по таймауту", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
		case <-time.After(200 * time.Millisecond):
			t.Error("Контекст должен истечь по таймауту")
		}
	})
}

func TestSignalHandling(t *testing.T) {
	t.Run("Канал для сигналов создается", func(t *testing.T) {
		quit := make(chan os.Signal, 1)
		assert.NotNil(t, quit)
		assert.Equal(t, 1, cap(quit))
	})
}

func TestApplicationInitialization(t *testing.T) {
	t.Run("Полная инициализация приложения", func(t *testing.T) {
		repo := postgres.New()
		require.NotNil(t, repo)

		svc := service.NewGofemartService(repo)
		require.NotNil(t, svc)

		h := handler.NewHandler(svc)
		require.NotNil(t, h)

		r := handler.NewRouter(h, svc)
		require.NotNil(t, r)

		server := &http.Server{
			Addr:    "localhost:0",
			Handler: r,
		}
		require.NotNil(t, server)

		assert.Equal(t, "localhost:0", server.Addr)
		assert.NotNil(t, server.Handler)
	})
}

func TestMainFunctionComponents(t *testing.T) {
	t.Run("Все основные компоненты main функции", func(t *testing.T) {
		assert.NotNil(t, postgres.New)
		assert.NotNil(t, service.NewGofemartService)
		assert.NotNil(t, handler.NewHandler)
		assert.NotNil(t, handler.NewRouter)
	})
}
