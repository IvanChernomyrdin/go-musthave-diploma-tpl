package postgres

import (
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

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

// найден пользователь по ID
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
