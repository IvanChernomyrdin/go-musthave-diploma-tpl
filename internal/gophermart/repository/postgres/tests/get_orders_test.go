package postgres

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"go-musthave-diploma-tpl/internal/gophermart/models"
)

// Успешное получение заказов
func TestPostgresStorage_GetOrders_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1
	expectedOrders := []models.Order{
		{
			Number:     "1234567890",
			Status:     "PROCESSED",
			Accrual:    100.5,
			UploadedAt: time.Now().Add(-24 * time.Hour),
		},
		{
			Number:     "0987654321",
			Status:     "NEW",
			Accrual:    0,
			UploadedAt: time.Now(),
		},
	}

	// ожидаем запрос на получение заказов
	rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"})
	for _, order := range expectedOrders {
		rows.AddRow(order.Number, order.Status, order.Accrual, order.UploadedAt)
	}

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`)).
		WithArgs(userID).
		WillReturnRows(rows)

	orders, err := storage.GetOrders(userID)

	assert.NoError(t, err)
	assert.NotNil(t, orders)
	assert.Len(t, orders, 2)
	assert.Equal(t, expectedOrders[0].Number, orders[0].Number)
	assert.Equal(t, expectedOrders[0].Status, orders[0].Status)
	assert.Equal(t, expectedOrders[0].Accrual, orders[0].Accrual)
	assert.Equal(t, expectedOrders[1].Number, orders[1].Number)
	assert.Equal(t, expectedOrders[1].Status, orders[1].Status)
	assert.Equal(t, expectedOrders[1].Accrual, orders[1].Accrual)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Получение пустого списка заказов
func TestPostgresStorage_GetOrders_Empty(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1

	// ожидаем пустой результат
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`)).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}))

	orders, err := storage.GetOrders(userID)

	assert.NoError(t, err)
	assert.Empty(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Ошибка базы данных при получении заказов
func TestPostgresStorage_GetOrders_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1

	// ожидаем ошибку базы данных
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`)).
		WithArgs(userID).
		WillReturnError(fmt.Errorf("database connection failed"))

	orders, err := storage.GetOrders(userID)

	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}
