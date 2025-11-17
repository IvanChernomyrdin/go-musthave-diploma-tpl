package postgres

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	postgres "go-musthave-diploma-tpl/internal/gophermart/repository/postgres"
)

// Вспомогательная функция для создания PostgresStorage с mock DB
func newTestStorage(db *sql.DB) *postgres.PostgresStorage {
	return postgres.NewWithDB(db)
}

// удачное создание
func TestPostgresStorage_CreateUser_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	// вычисляем реальный хеш
	expectedHash := postgres.HashPassword("password123")
	createdAt := time.Now()

	// Mock: проверяем что пользователь не существует
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE login = $1`)).
		WithArgs("newuser").
		WillReturnError(sql.ErrNoRows) // Пользователь не существует

	// Mock: ожидаем успешное создание пользователя
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id, login, password_hash, created_at`)).
		WithArgs("newuser", expectedHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
			AddRow(1, "newuser", expectedHash, createdAt))

	// Act - выполняем тестируемый метод
	user, err := storage.CreateUser("newuser", "password123")

	// сравниваем результаты ожиданий с реальными
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "newuser", user.Login)

	// Проверяем что все ожидания выполнены
	assert.NoError(t, mock.ExpectationsWereMet())
}

// создание, пользователь уже существует
func TestPostgresStorage_CreateUser_LoginExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	createdAt := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE login = $1`)).
		WithArgs("existinguser").
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
			AddRow(1, "existinguser", "hash", createdAt))

	user, err := storage.CreateUser("existinguser", "password123")

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, "login already exists", err.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

// удачная авторизация по логину и паролю
func TestPostgresStorage_GetUserByLoginAndPassword_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	expectedHash := postgres.HashPassword("correctpassword")
	createdAt := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE login = $1 AND password_hash = $2`)).
		WithArgs("testuser", expectedHash).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
			AddRow(1, "testuser", expectedHash, createdAt))

	user, err := storage.GetUserByLoginAndPassword("testuser", "correctpassword")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// неудачная авторизация, не существует
func TestPostgresStorage_GetUserByLoginAndPassword_WrongPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	wrongPasswordHash := postgres.HashPassword("wrongpassword")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE login = $1 AND password_hash = $2`)).
		WithArgs("testuser", wrongPasswordHash).
		WillReturnError(sql.ErrNoRows)

	user, err := storage.GetUserByLoginAndPassword("testuser", "wrongpassword")

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// не найден пользователь по логину
func TestPostgresStorage_GetUserByLogin_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE login = $1`)).
		WithArgs("nonexistent").
		WillReturnError(sql.ErrNoRows)

	user, err := storage.GetUserByLogin("nonexistent")

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// найден пользователь по логину
func TestPostgresStorage_GetUserByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	createdAt := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE id = $1`)).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "login", "password_hash", "created_at"}).
			AddRow(1, "testuser", "hash", createdAt))

	user, err := storage.GetUserByID(1)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, 1, user.ID)
	assert.Equal(t, "testuser", user.Login)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// не найден пользователь по ID
func TestPostgresStorage_GetUserByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, login, password_hash, created_at FROM users WHERE id = $1`)).
		WithArgs(999).
		WillReturnError(sql.ErrNoRows)

	user, err := storage.GetUserByID(999)

	assert.NoError(t, err)
	assert.Nil(t, user)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// проверка хэш пароля
func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hash := postgres.HashPassword(password)

	assert.Len(t, hash, 64)            // SHA256 produces 64 символа
	assert.NotEqual(t, password, hash) // не должен совпадать

	// Проверяем детерминированность
	hash2 := postgres.HashPassword(password)
	assert.Equal(t, hash, hash2) // должен совпадать
}
