package runtime

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type HTTPLogger struct {
	*zap.Logger
}

func NewHTTPLogger() *HTTPLogger {
	// Создаем папку runtime/log
	logDir := "internal/runtime/log"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("Failed to create log directory: " + err.Error())
	}

	// Настраиваем lumberjack
	logPath := filepath.Join(logDir, "http.log")
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    5,
		MaxBackups: 0,
		MaxAge:     1095, // 3 года в днях
		Compress:   true,
	}

	// Настройки формата с эмодзи
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("02.01.2006 15:04:05"), // перевод времени как в россии
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Console формат с эмодзи
	fileEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileWriter := zapcore.AddSync(lumberjackLogger)
	fileCore := zapcore.NewCore(fileEncoder, fileWriter, zap.InfoLevel)

	core := zapcore.NewTee(fileCore)

	logger := zap.New(core)

	return &HTTPLogger{Logger: logger}
}

func (logger *HTTPLogger) LogRequest(method, uri string, status, responseSize int, duration float64) {
	// Эмодзи в зависимости от статуса
	var emoji string
	switch {
	case status >= 200 && status < 300:
		emoji = "✅"
	case status >= 400 && status < 500:
		emoji = "⚠️"
	case status >= 500:
		emoji = "❌"
	default:
		emoji = "🔵"
	}

	message := emoji + " HTTP " + method + " " + uri

	logger.Info(message,
		zap.Int("status", status),
		zap.Int("size", responseSize),
		zap.Float64("duration_ms", duration),
	)
}

func (logger *HTTPLogger) Close() error {
	return logger.Sync()
}
