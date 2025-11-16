package postgres

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"go-musthave-diploma-tpl/internal/gophermart/config/db"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	logger "go-musthave-diploma-tpl/internal/gophermart/runtime/logger"
)

// кастомный логгер записывает в файл runtime/log
var castomLogger = logger.NewHTTPLogger().Sugar()

type PostgresStorage struct {
	db              *sql.DB
	errorClassifier *PostgresErrorClassifier
}

func New() *PostgresStorage {
	return &PostgresStorage{
		db:              db.GetDB(),
		errorClassifier: NewPostgresErrorClassifier(),
	}
}

// хешируем пароль в sha256
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// ищем по логину (для проверки существования)
func (ps *PostgresStorage) GetUserByLogin(login string) (*models.User, error) {
	var user models.User

	query := `SELECT 
					id, 
					login, 
					password_hash, 
					created_at 
				FROM users 
				WHERE login = $1`
	err := ps.db.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		castomLogger.Infof("failed to get user by login: %v", err)
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return &user, nil
}

// получение пользователя по id
func (ps *PostgresStorage) GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`

	err := ps.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		castomLogger.Infof("failed to get user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// ищем по логину и паролю
func (ps *PostgresStorage) GetUserByLoginAndPassword(login, password string) (*models.User, error) {
	var user models.User
	hashedPassword := HashPassword(password)

	query := `SELECT 
					id, 
					login, 
					password_hash, 
					created_at 
				FROM users 
				WHERE login = $1 
					AND password_hash = $2`
	err := ps.db.QueryRow(query, login, hashedPassword).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		castomLogger.Infof("failed to get user by login and password: %v", err)
		return nil, fmt.Errorf("failed to get user by login and password: %w", err)
	}

	return &user, nil
}

// создаём пользователя
func (ps *PostgresStorage) CreateUser(login, password string) (*models.User, error) {
	// проверяем что такого логина нет (используем GetUserByLogin)
	existingUser, err := ps.GetUserByLogin(login)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("login already exists")
	}

	hashedPassword := HashPassword(password)

	var user models.User
	// добавляем
	query := `INSERT INTO users (login, password_hash) 
        	  VALUES ($1, $2) 
        	  RETURNING id, login, password_hash, created_at`

	err = ps.db.QueryRow(query, login, hashedPassword).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		castomLogger.Infof("failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func NewWithDB(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{
		db:              db,
		errorClassifier: NewPostgresErrorClassifier(),
	}
}
