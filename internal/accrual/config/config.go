package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

var (
	runAddress  string
	databaseUrl string
	maxRequests int
	timeout     int
)

type Config struct {
	RunAddress  string `env:"RUN_ADDRESS"`
	DatabaseUrl string `env:"DATABASE_URI"`
	MaxRequests int    `env:"MAX_REQUESTS"`
	Timeout     int    `env:"TIMEOUT"`
}

func Load() *Config {
	cfg := &Config{}

	flag.StringVar(&runAddress, "a", "localhost:8082", "адрес и порт запуска сервиса")
	flag.StringVar(&databaseUrl, "d", "postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable", "адрес подключения к базе данных")
	flag.IntVar(&maxRequests, "m", 100, "максимальное количество запросов")
	flag.IntVar(&timeout, "t", 10, "таймаут в секундах")
	flag.Parse()
	cfg.applyEnv()
	return cfg
}

func (cfg *Config) applyEnv() {
	env.Parse(&cfg)
	if cfg.RunAddress == "" {
		cfg.RunAddress = runAddress
	}
	if cfg.DatabaseUrl == "" {
		cfg.DatabaseUrl = databaseUrl
	}
	if cfg.MaxRequests == 0 {

		cfg.MaxRequests = maxRequests
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = timeout
	}

}
