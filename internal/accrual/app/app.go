package app

import (
	"context"
	"go-musthave-diploma-tpl/internal/accrual/config"
	"go-musthave-diploma-tpl/internal/accrual/handler"
	"go-musthave-diploma-tpl/internal/accrual/repository"
	"go-musthave-diploma-tpl/internal/accrual/router"
	"go-musthave-diploma-tpl/internal/accrual/service"
	"go-musthave-diploma-tpl/internal/accrual/storage"
	logger "go-musthave-diploma-tpl/pkg/runtime/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
)

func Run() {

	customLogger := logger.NewHTTPLogger().Logger.Sugar()
	cfg := config.Load()

	storage, err := storage.InitPostgresDB(cfg.DatabaseUrl)

	if err != nil {

		customLogger.Fatalw("failed to connect to database", "error", err)
	}
	customLogger.Infof("Миграции применены успешно")

	defer storage.Close()

	repo := repository.NewRepository(storage)
	svc := service.NewService(repo)
	handler := handler.NewHandler(svc)

	r := chi.NewRouter()

	// Create a cancellable context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	router.NewRouter(customLogger, ctx, r, handler, cfg)

	//Обновляю статус и бонусы зказов.
	go svc.Listener(ctx, customLogger, time.Duration(cfg.PollingInterval)*time.Second)

	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}
	//start server and graceful shutdown

	go func() {
		customLogger.Infof("Сервер запущен на %s", cfg.RunAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			customLogger.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-quit
	customLogger.Info("Завершение работы сервера...")

	// Cancel the context to signal the listener to stop
	cancel()

	// Create a timeout context for server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		customLogger.Fatalf("Принудительное завершение: %v", err)
	}
	select {
	case <-shutdownCtx.Done():
		if shutdownCtx.Err() == context.DeadlineExceeded {
			customLogger.Info("timeout of 30sec occurred", " time:", time.Now())
		}
	default:
		time.Sleep(10 * time.Second)
		customLogger.Info("Сервер остановлен")
	}

}
