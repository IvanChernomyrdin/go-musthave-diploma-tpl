package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
}

const EncryptionKey = "32-bytes-long-key-1234567890777!"

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RunAddress, "a", "localhost:8080", "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable", "адрес подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "localhost:8081", "адрес системы расчёта начислений")

	flag.Parse()
	cfg.applyEnv()
	return cfg
}

func (cfg *Config) applyEnv() {
	if envRunAddress := os.Getenv("RUN_ADDRESS"); envRunAddress != "" {
		cfg.RunAddress = envRunAddress
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		cfg.DatabaseURI = envDatabaseURI
	}
	if envAccrualSystemAddress := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envAccrualSystemAddress != "" {
		cfg.AccrualSystemAddress = envAccrualSystemAddress
	}
}
