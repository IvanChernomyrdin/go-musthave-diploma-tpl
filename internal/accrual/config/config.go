package config

import (
	"flag"
	"os"
)

type Config struct {
	RUN_ADDRESS  string
	DATABASE_URI string
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.RUN_ADDRESS, "a", "localhost:8082", "адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DATABASE_URI, "d", "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable", "адрес подключения к базе данных")

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
}
