package postgres

import (
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage_GetBalance(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	ps := newTestStorage(db)

	tests := []struct {
		name           string
		userID         int
		setupMock      func()
		expectedResult models.Balance
		expectError    bool
	}{
		{
			name:   "Успешное получение баланса",
			userID: 1,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"current", "withdrawn"}).
					AddRow(500.5, 42.0)
				mock.ExpectQuery(`SELECT`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedResult: models.Balance{
				Current:   500.5,
				Withdrawn: 42,
			},
			expectError: false,
		},
		{
			name:   "Баланс пустой",
			userID: 2,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"current", "withdrawn"}).
					AddRow(0, 0)
				mock.ExpectQuery(`SELECT`).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectedResult: models.Balance{
				Current:   0,
				Withdrawn: 0,
			},
			expectError: false,
		},
		{
			name:   "Ошибка базы данных",
			userID: 3,
			setupMock: func() {
				mock.ExpectQuery(`SELECT`).
					WithArgs(3).
					WillReturnError(assert.AnError)
			},
			expectedResult: models.Balance{},
			expectError:    true,
		},
		{
			name:   "Нет строк (возвращает нули)",
			userID: 4,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"current", "withdrawn"})
				mock.ExpectQuery(`SELECT`).
					WithArgs(4).
					WillReturnRows(rows)
			},
			expectedResult: models.Balance{},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := ps.GetBalance(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
