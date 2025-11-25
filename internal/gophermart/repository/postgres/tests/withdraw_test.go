package postgres

import (
	"database/sql"
	"testing"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresStorage_Withdraw(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		withdraw      models.WithdrawBalance
		setupMock     func(mock sqlmock.Sqlmock)
		expectedError error
	}{
		{
			name:   "Successful withdrawal",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   751,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				// расчёт баланса
				mock.ExpectQuery(`SELECT.*balance`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(1000.0))

				// запись списания
				mock.ExpectExec(`INSERT INTO withdrawals`).
					WithArgs(1, "2377225624", 751.0).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
			expectedError: nil,
		},
		{
			name:   "Insufficient funds",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   1000,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

				// баланс меньше суммы списания
				mock.ExpectQuery(`SELECT.*balance`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"balance"}).AddRow(500.0))

				mock.ExpectRollback()
			},
			expectedError: handler.ErrLackOfFunds,
		},
		{
			name:   "Database error when calculating balance",
			userID: 1,
			withdraw: models.WithdrawBalance{
				Order: "2377225624",
				Sum:   500,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()

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
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin().WillReturnError(sql.ErrConnDone)
			},
			expectedError: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error creating mock: %v", err)
			}
			defer db.Close()

			storage := &postgres.PostgresStorage{DB: db}

			tt.setupMock(mock)

			err = storage.Withdraw(tt.userID, tt.withdraw)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == handler.ErrLackOfFunds {
					assert.Equal(t, tt.expectedError, err)
				} else {
					assert.ErrorContains(t, err, tt.expectedError.Error())
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
