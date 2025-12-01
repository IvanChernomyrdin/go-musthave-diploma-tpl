package config

import (
	"flag"
	"os"
	"strings"
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
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "http://localhost:8081", "адрес системы расчёта начислений")

	flag.Parse()

	cfg.applyEnv()

	if !strings.HasPrefix(cfg.AccrualSystemAddress, "http://") &&
		!strings.HasPrefix(cfg.AccrualSystemAddress, "https://") {
		cfg.AccrualSystemAddress = "http://" + cfg.AccrualSystemAddress
	}

	return cfg
}

func (cfg *Config) applyEnv() {
	if v := os.Getenv("RUN_ADDRESS"); v != "" {
		cfg.RunAddress = v
	}
	if v := os.Getenv("DATABASE_URI"); v != "" {
		cfg.DatabaseURI = v
	}
	if v := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); v != "" {
		cfg.AccrualSystemAddress = v
	}
}
