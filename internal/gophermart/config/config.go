package config

import (
	"flag"
	"os"
)

type Config struct {
	RUN_ADDRESS            string
	DATABASE_URI           string
	ACCRUAL_SYSTEM_ADDRESS string
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RUN_ADDRESS, "a", "localhost:8081", "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DATABASE_URI, "d", "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable", "адрес подключения к базе данных")
	flag.StringVar(&cfg.ACCRUAL_SYSTEM_ADDRESS, "r", "localhost:8082", "адрес системы расчёта начислений")

	flag.Parse()
	cfg.applyEnv()
	return cfg
}

func (cfg *Config) applyEnv() {
	if envRUN_ADDRESS := os.Getenv("RUN_ADDRESS"); envRUN_ADDRESS != "" {
		cfg.RUN_ADDRESS = envRUN_ADDRESS
	}
	if envDATABASE_URI := os.Getenv("DATABASE_URI"); envDATABASE_URI != "" {
		cfg.DATABASE_URI = envDATABASE_URI
	}
	if envACCRUAL_SYSTEM_ADDRESS := os.Getenv("ACCRUAL_SYSTEM_ADDRESS"); envACCRUAL_SYSTEM_ADDRESS != "" {
		cfg.ACCRUAL_SYSTEM_ADDRESS = envACCRUAL_SYSTEM_ADDRESS
	}
}
