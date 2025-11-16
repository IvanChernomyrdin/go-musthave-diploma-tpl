package postgres

import (
	"database/sql"
	"go-musthave-diploma-tpl/internal/config/db"
	"time"
)

// тут будет реализация роутов

type RetryConfig struct {
	MaxAttempts  int
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Second,
	}
}

type PostgresStorage struct {
	db              *sql.DB
	retryConfig     RetryConfig
	errorClassifier *PostgresErrorClassifier
}

func New() *PostgresStorage {
	return &PostgresStorage{
		db:              db.GetDB(),
		retryConfig:     DefaultRetryConfig(),
		errorClassifier: NewPostgresErrorClassifier(),
	}
}
