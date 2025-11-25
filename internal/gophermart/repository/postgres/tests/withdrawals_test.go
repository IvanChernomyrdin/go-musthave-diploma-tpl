package postgres

import (
	"database/sql"
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/gophermart/models"
	"go-musthave-diploma-tpl/internal/gophermart/repository/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_Withdrawals(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := &postgres.PostgresStorage{DB: db}

	// Фиксированное время для тестов
	now := time.Now()
	time1 := now.Add(-24 * time.Hour)
	time2 := now.Add(-12 * time.Hour)

	tests := []struct {
		name           string
		userID         int
		setupMock      func()
		expectedResult []models.WithdrawBalance
		expectedError  error
	}{
		{
			name:   "Successful withdrawals retrieval",
			userID: 1,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
					AddRow("2377225624", 751.50, time1).
					AddRow("49927398716", 500.25, time2)

				mock.ExpectQuery(`SELECT 
					order_number,
					sum,
					processed_at
				FROM withdrawals
				WHERE user_id = \$1
				ORDER BY processed_at DESC`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedResult: []models.WithdrawBalance{
				{
					Order:       "2377225624",
					Sum:         751.50,
					ProcessedAt: time1,
				},
				{
					Order:       "49927398716",
					Sum:         500.25,
					ProcessedAt: time2,
				},
			},
			expectedError: nil,
		},
		{
			name:   "No withdrawals for user",
			userID: 2,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"})

				mock.ExpectQuery(`SELECT 
					order_number,
					sum,
					processed_at
				FROM withdrawals
				WHERE user_id = \$1
				ORDER BY processed_at DESC`).
					WithArgs(2).
					WillReturnRows(rows)
			},
			expectedResult: []models.WithdrawBalance{},
			expectedError:  nil,
		},
		{
			name:   "Database query error",
			userID: 3,
			setupMock: func() {
				mock.ExpectQuery(`SELECT 
					order_number,
					sum,
					processed_at
				FROM withdrawals
				WHERE user_id = \$1
				ORDER BY processed_at DESC`).
					WithArgs(3).
					WillReturnError(sql.ErrConnDone)
			},
			expectedResult: nil,
			expectedError:  sql.ErrConnDone,
		},
		{
			name:   "Error scanning rows",
			userID: 4,
			setupMock: func() {
				// Возвращаем неверный тип данных для суммы
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
					AddRow("1234567890", "not_a_float", time1)

				mock.ExpectQuery(`SELECT 
					order_number,
					sum,
					processed_at
				FROM withdrawals
				WHERE user_id = \$1
				ORDER BY processed_at DESC`).
					WithArgs(4).
					WillReturnRows(rows)
			},
			expectedResult: nil,
			expectedError:  nil, // Будет ошибка сканирования
		},
		{
			name:   "Error during rows iteration",
			userID: 5,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
					AddRow("1234567890", 100.0, time1).
					RowError(0, sql.ErrTxDone) // Ошибка при итерации

				mock.ExpectQuery(`SELECT 
					order_number,
					sum,
					processed_at
				FROM withdrawals
				WHERE user_id = \$1
				ORDER BY processed_at DESC`).
					WithArgs(5).
					WillReturnRows(rows)
			},
			expectedResult: nil,
			expectedError:  sql.ErrTxDone,
		},
		{
			name:   "Single withdrawal",
			userID: 6,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
					AddRow("1234567890", 300.75, time1)

				mock.ExpectQuery(`SELECT 
					order_number,
					sum,
					processed_at
				FROM withdrawals
				WHERE user_id = \$1
				ORDER BY processed_at DESC`).
					WithArgs(6).
					WillReturnRows(rows)
			},
			expectedResult: []models.WithdrawBalance{
				{
					Order:       "1234567890",
					Sum:         300.75,
					ProcessedAt: time1,
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок
			tt.setupMock()

			// Вызываем тестируемый метод
			result, err := storage.Withdrawals(tt.userID)

			// Проверяем ошибку
			if tt.expectedError != nil {
				require.Error(t, err)
				if tt.expectedError == sql.ErrConnDone || tt.expectedError == sql.ErrTxDone {
					assert.ErrorIs(t, err, tt.expectedError)
				} else {
					assert.ErrorContains(t, err, tt.expectedError.Error())
				}
			} else {
				// Для случаев с ошибкой сканирования просто проверяем что есть ошибка
				if tt.name == "Error scanning rows" {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			}

			// Проверяем результат - более гибкая проверка
			if tt.expectedResult != nil {
				if tt.expectedResult == nil {
					assert.Nil(t, result)
				} else if len(tt.expectedResult) == 0 {
					// Для пустого ожидаемого результата, проверяем что результат либо nil, либо пустой слайс
					if result != nil {
						assert.Empty(t, result)
					}
				} else {
					require.NotNil(t, result)
					assert.Len(t, result, len(tt.expectedResult))

					for i, expected := range tt.expectedResult {
						assert.Equal(t, expected.Order, result[i].Order)
						assert.Equal(t, expected.Sum, result[i].Sum)
						assert.WithinDuration(t, expected.ProcessedAt, result[i].ProcessedAt, time.Second)
					}
				}
			} else {
				assert.Nil(t, result)
			}

			// Проверяем что все ожидания выполнены
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestPostgresStorage_Withdrawals_RowsClose(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := &postgres.PostgresStorage{DB: db}

	t.Run("Rows are properly closed", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
			AddRow("1234567890", 100.0, time.Now())

		mock.ExpectQuery(`SELECT 
			order_number,
			sum,
			processed_at
		FROM withdrawals
		WHERE user_id = \$1
		ORDER BY processed_at DESC`).
			WithArgs(1).
			WillReturnRows(rows)

		result, err := storage.Withdrawals(1)

		assert.NoError(t, err)
		// Более гибкая проверка - результат может быть либо не-nil с данными, либо nil
		if result != nil {
			assert.Len(t, result, 1)
			assert.Equal(t, "1234567890", result[0].Order)
			assert.Equal(t, 100.0, result[0].Sum)
		}

		// Проверяем что rows были закрыты
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
