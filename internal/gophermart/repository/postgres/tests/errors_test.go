package postgres

import (
	"errors"
	"fmt"
	"testing"

	postgresError "go-musthave-diploma-tpl/internal/gophermart/repository/postgres"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

// запускаем errorClassifier и ClassifyPostgresError пробегаем по ошибке и ретрай или не ретрай
func TestPostgresErrorClassifier_Classify(t *testing.T) {
	classifier := postgresError.NewPostgresErrorClassifier()

	tests := []struct {
		name     string
		err      error
		expected postgresError.ErrorClassification
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: postgresError.NonRetriable,
		},
		{
			name:     "Non-Postgres error",
			err:      errors.New("some random error"),
			expected: postgresError.NonRetriable,
		},
		{
			name: "Connection exception (08 class) - Retriable",
			err: &pgconn.PgError{
				Code: "08000",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Serialization failure - Retriable",
			err: &pgconn.PgError{
				Code: "40001",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Deadlock detected - Retriable",
			err: &pgconn.PgError{
				Code: "40P01",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Unique violation - NonRetriable",
			err: &pgconn.PgError{
				Code: "23505",
			},
			expected: postgresError.NonRetriable,
		},
		{
			name: "Foreign key violation - NonRetriable",
			err: &pgconn.PgError{
				Code: "23503",
			},
			expected: postgresError.NonRetriable,
		},
		{
			name: "Check violation - NonRetriable",
			err: &pgconn.PgError{
				Code: "23514",
			},
			expected: postgresError.NonRetriable,
		},
		{
			name: "Syntax error - NonRetriable",
			err: &pgconn.PgError{
				Code: "42601",
			},
			expected: postgresError.NonRetriable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.Classify(tt.err)
			assert.Equal(t, tt.expected, result, "For error: %v", tt.err)
		})
	}
}

func TestPostgresErrorClassifier_ClassifyPostgresError(t *testing.T) {
	classifier := postgresError.NewPostgresErrorClassifier()

	tests := []struct {
		name     string
		pgErr    *pgconn.PgError
		expected postgresError.ErrorClassification
	}{
		{
			name: "Connection exceptions (08 class)",
			pgErr: &pgconn.PgError{
				Code: "08000",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Admin shutdown",
			pgErr: &pgconn.PgError{
				Code: "57P01",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Crash shutdown",
			pgErr: &pgconn.PgError{
				Code: "57P02",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Cannot connect now",
			pgErr: &pgconn.PgError{
				Code: "57P03",
			},
			expected: postgresError.Retriable,
		},
		{
			name: "Unknown error code",
			pgErr: &pgconn.PgError{
				Code: "99999",
			},
			expected: postgresError.NonRetriable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Используем reflection для тестирования приватного метода
			// или тестируем через публичный Classify
			result := classifier.Classify(tt.pgErr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// разбор пограничных случаев
func TestPostgresErrorClassifier_EdgeCases(t *testing.T) {
	classifier := postgresError.NewPostgresErrorClassifier()
	// длолжен распознать ошибку в обёртке другой ошибке
	t.Run("Wrapped Postgres error", func(t *testing.T) {

		pgErr := &pgconn.PgError{Code: "40001"}
		wrappedErr := fmt.Errorf("wrapper: %w", pgErr)

		result := classifier.Classify(wrappedErr)
		assert.Equal(t, postgresError.Retriable, result)
	})

	// должен вернуть неретрай т.к. ошибка фейковая
	t.Run("Non-PgError with same interface", func(t *testing.T) {
		// Создаем ошибку, которая не является *pgconn.PgError,
		// но имеет похожие методы
		type fakePgError struct {
			error
			code string
		}

		fakeErr := &fakePgError{
			error: errors.New("fake error"),
			code:  "40001",
		}

		result := classifier.Classify(fakeErr)
		assert.Equal(t, postgresError.NonRetriable, result)
	})
}
