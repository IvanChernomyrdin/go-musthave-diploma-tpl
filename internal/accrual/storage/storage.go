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

type PostgresDB struct {
	DB *sql.DB
}

func InitPostgresDB(databaseDSN string) (*PostgresDB, error) {
	connection := strings.Trim(databaseDSN, `"`)

	var err error
	db, err := sql.Open("pgx", connection)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("проверка подключения к БД не удалась: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания драйвера миграций: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/accrual/migrations/",
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания миграции для системы лояльности: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("ошибка применения миграций для системы лояльности: %v", err)
	}

	return &PostgresDB{DB: db}, nil
}

func (db *PostgresDB) CheckConnection() error {
	err := db.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgresDB) CreateProductReward(match string, reward float64, rewardType string) error {
	op := "path: internal/accrual/storage/CreateProductReward"
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s starts a transaction err:%w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var exists bool
	err = tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM products
			WHERE match = $1
		)
	`, match).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s QueryRow err:%w", op, err)
	}

	if exists {
		return ErrKeyExists
	}

	_, err = tx.Exec(`
		INSERT INTO products (match, reward, reward_type)
		VALUES ($1, $2, $3)
	`, match, reward, rewardType)
	if err != nil {
		return fmt.Errorf("%s Exec err:%w", op, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s Commit err:%w", op, err)
	}
	return nil
}
