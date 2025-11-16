package main

import (
	"context"
	config "go-musthave-diploma-tpl/internal/gophermart/config"
	db "go-musthave-diploma-tpl/internal/gophermart/config/db"
	chiRouter "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/repository/postgres"
	logger "go-musthave-diploma-tpl/internal/gophermart/runtime/logger"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// конфиг
	cfg := config.Load()
	// логгер
	castomLogger := logger.NewHTTPLogger().Logger.Sugar()
	// бд и миграции
	if err := db.Init(cfg.DATABASE_URI); err != nil {
		castomLogger.Fatalf("PostgreSQL недоступна: %v", err)
	} else {
		castomLogger.Infof("Миграции применены успешно")
	}
	// chi роутер
	repo := postgres.New()
	//что должен реализовывать этот сервис
	svc := service.NewGofemartService(repo)
	//инициализируем хандлеры
	h := chiRouter.NewHandler(svc)
	//инициализируем роуты
	r := chiRouter.NewRouter(h, svc)

	//создаём серве
	server := &http.Server{
		Addr:    cfg.RUN_ADDRESS,
		Handler: r,
	}
	//start server and graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		castomLogger.Infof("Сервер запущен на %s", cfg.RUN_ADDRESS)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			castomLogger.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	<-quit
	castomLogger.Info("Завершение работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		castomLogger.Fatalf("Принудительное завершение: %v", err)
	}
	if err := logger.NewHTTPLogger().Close(); err != nil {
		castomLogger.Fatalf("Логгер не завершил работу: %v", err)
	}
	castomLogger.Info("Сервер остановлен")
}
