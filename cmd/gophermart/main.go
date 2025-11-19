package main

import (
	"context"
	config "go-musthave-diploma-tpl/internal/gophermart/config"
	db "go-musthave-diploma-tpl/internal/gophermart/config/db"
	chiRouter "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/listener"
	"go-musthave-diploma-tpl/internal/gophermart/repository/postgres"
	"go-musthave-diploma-tpl/internal/gophermart/service"
	logger "go-musthave-diploma-tpl/pkg/runtime/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	// конфиг
	cfg := config.Load()
	// логгер
	customLogger := logger.NewHTTPLogger().Logger.Sugar()
	// бд и миграции
	if err := db.Init(cfg.DatabaseURI); err != nil {
		customLogger.Fatalf("PostgreSQL недоступна: %v", zap.Error(err))
	} else {
		customLogger.Infof("Миграции применены успешно")
	}
	// chi роутер
	repo := postgres.New()
	//что должен реализовывать этот сервис
	svc := service.NewGofemartService(repo)
	//инициализируем хандлеры
	h := chiRouter.NewHandler(svc)
	//инициализируем роуты
	r := chiRouter.NewRouter(h, svc)

	// запускаем слушателя
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ПЕРЕДАЕМ АДРЕС СЕРВИСА НАЧИСЛЕНИЙ В LISTENER
	orderListener := listener.NewOrderListener(cfg.DatabaseURI, cfg.AccrualSystemAddress, customLogger)
	orderListener.Start(ctx)

	//создаём серве
	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}
	//start server and graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		customLogger.Infof("Сервер запущен на %s", cfg.RunAddress)
		customLogger.Infof("Accrual system address: %s", cfg.AccrualSystemAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			customLogger.Fatalf("Ошибка сервера: %v", zap.Error(err))
		}
	}()

	<-quit
	customLogger.Info("Завершение работы сервера...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		customLogger.Fatalf("Принудительное завершение: %v", zap.Error(err))
	}
	if err := logger.NewHTTPLogger().Close(); err != nil {
		customLogger.Fatalf("Логгер не завершил работу: %v", zap.Error(err))
	}
	customLogger.Info("Сервер остановлен")
}
