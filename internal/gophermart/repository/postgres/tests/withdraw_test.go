package postgres

import (
	"database/sql"
	"testing"

	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage_Withdraw(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := &postgres.PostgresStorage{DB: db}

	tests := []struct {
		name          string
		userID        int
		withdraw      models.WithdrawBalance
		setupMock     func()
		expectedError error
	}{
		{
			name:   "Successful withdrawal",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			setupMock: func() {
				// Ожидаем начало транзакции
				mock.ExpectBegin()

				// Проверка существования заказа - убрать лишние пробелы
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM orders WHERE number = \$1\s*\)`).
					WithArgs("2377225624").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

				// Расчет баланса - упростить регулярное выражение
				mock.ExpectQuery(`SELECT.*balance`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1000.0))

				// Вставка списания - упростить регулярное выражение
				mock.ExpectExec(`INSERT INTO withdrawals`).
					WithArgs(1, "2377225624", 751.0).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Коммит транзакции
				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:   "Order does not exist",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "9999999999",
				Sum:   500,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM orders WHERE number = \$1\s*\)`).
					WithArgs("9999999999").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
				mock.ExpectRollback()
			},
			expectedError: models.ErrInvalidOrderNumber,
		},
		{
			name:   "Insufficient funds",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   1000,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM orders WHERE number = \$1\s*\)`).
					WithArgs("2377225624").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT.*balance`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(500.0))
				mock.ExpectRollback()
			},
			expectedError: models.ErrLackOfFunds,
		},
		{
			name:   "Database error when checking order existence",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   500,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM orders WHERE number = \$1\s*\)`).
					WithArgs("2377225624").
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedError: sql.ErrConnDone,
		},
		{
			name:   "Database error when calculating balance",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   500,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM orders WHERE number = \$1\s*\)`).
					WithArgs("2377225624").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT.*balance`).
					WithArgs(1).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedError: sql.ErrConnDone,
		},
		{
			name:   "Database error when inserting withdrawal",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   500,
			},
			setupMock: func() {
				mock.ExpectBegin()
				mock.ExpectQuery(`SELECT EXISTS\s*\(\s*SELECT 1 FROM orders WHERE number = \$1\s*\)`).
					WithArgs("2377225624").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
				mock.ExpectQuery(`SELECT.*balance`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1000.0))
				mock.ExpectExec(`INSERT INTO withdrawals`).
					WithArgs(1, "2377225624", 500.0).
					WillReturnError(sql.ErrConnDone)
				mock.ExpectRollback()
			},
			expectedError: sql.ErrConnDone,
		},
		{
			name:   "Transaction begin error",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   500,
			},
			setupMock: func() {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := storage.Withdraw(tt.userID, tt.withdraw)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == models.ErrInvalidOrderNumber || tt.expectedError == models.ErrLackOfFunds {
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.ErrorContains(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			// Проверяем что все ожидания выполнены
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
