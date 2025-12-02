package postgres

import (
	"testing"

	"github.com/stretchr/testify/assert"

	postgres "go-musthave-diploma-tpl/internal/gophermart/repository/postgres"
)

// проверка хэш пароля
func TestHashPassword(t *testing.T) {
	password := "testpassword"
	hash := postgres.HashPassword(password)

	assert.Len(t, hash, 64)
	assert.NotEqual(t, password, hash)

	// Проверяем детерминированность
	hash2 := postgres.HashPassword(password)
	assert.Equal(t, hash, hash2)
}
