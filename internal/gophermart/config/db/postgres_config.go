package storage

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var DB *sql.DB

func Init(databaseDSN string) error {
	connection := getConnect(databaseDSN)

	var err error
	DB, err = sql.Open("pgx", connection)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %v", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("проверка подключения к БД не удалась: %v", err)
	}

	driver, err := postgres.WithInstance(DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("ошибка создания драйвера миграций: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/gophermart/migrations/",
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("ошибка создания миграции для гофемарта: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("ошибка применения миграций для гофемарта: %v", err)
	}

	return nil
}

func getConnect(connectionFlag string) string {
	if connectionFlag != "" {
		return strings.Trim(connectionFlag, `"`)
	}
	return ""
}
