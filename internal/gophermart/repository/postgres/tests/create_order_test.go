package postgres

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
)

// Ошибка при сканировании строк
func TestPostgresStorage_GetOrders_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1

	// возвращаем некорректные данные (строка вместо числа для accrual)
	rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
		AddRow("1234567890", "PROCESSED", "not_a_number", time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`)).
		WithArgs(userID).
		WillReturnRows(rows)

	orders, err := storage.GetOrders(userID)

	assert.Error(t, err)
	assert.Nil(t, orders)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Успешное создание заказа
func TestPostgresStorage_CreateOrder_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1
	orderNumber := "12345678903"

	mock.ExpectQuery(regexp.QuoteMeta(`
        WITH inserted AS (
            INSERT INTO orders (user_id, number, status) 
            VALUES ($1, $2, $3)
            ON CONFLICT (number) DO NOTHING
            RETURNING uid
        ),
        existing AS (
            SELECT user_id FROM orders WHERE number = $2
        )
        SELECT 
            CASE 
                WHEN EXISTS (SELECT 1 FROM inserted) THEN 'inserted'::text
                WHEN EXISTS (SELECT 1 FROM existing WHERE user_id = $1) THEN 'duplicate'::text
                WHEN EXISTS (SELECT 1 FROM existing) THEN 'conflict'::text
                ELSE 'not_found'::text
            END as result`)).
		WithArgs(userID, orderNumber, models.OrderStatusNew).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow("inserted"))

	err = storage.CreateOrder(userID, orderNumber)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Дубликат заказа от того же пользователя
func TestPostgresStorage_CreateOrder_Duplicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1
	orderNumber := "12345678903"

	mock.ExpectQuery(regexp.QuoteMeta(`
        WITH inserted AS (
            INSERT INTO orders (user_id, number, status) 
            VALUES ($1, $2, $3)
            ON CONFLICT (number) DO NOTHING
            RETURNING uid
        ),
        existing AS (
            SELECT user_id FROM orders WHERE number = $2
        )
        SELECT 
            CASE 
                WHEN EXISTS (SELECT 1 FROM inserted) THEN 'inserted'::text
                WHEN EXISTS (SELECT 1 FROM existing WHERE user_id = $1) THEN 'duplicate'::text
                WHEN EXISTS (SELECT 1 FROM existing) THEN 'conflict'::text
                ELSE 'not_found'::text
            END as result`)).
		WithArgs(userID, orderNumber, models.OrderStatusNew).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow("duplicate"))

	err = storage.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Equal(t, handler.ErrDuplicateOrder, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Конфликт - заказ уже существует у другого пользователя
func TestPostgresStorage_CreateOrder_Conflict(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1
	orderNumber := "12345678903"

	mock.ExpectQuery(regexp.QuoteMeta(`
        WITH inserted AS (
            INSERT INTO orders (user_id, number, status) 
            VALUES ($1, $2, $3)
            ON CONFLICT (number) DO NOTHING
            RETURNING uid
        ),
        existing AS (
            SELECT user_id FROM orders WHERE number = $2
        )
        SELECT 
            CASE 
                WHEN EXISTS (SELECT 1 FROM inserted) THEN 'inserted'::text
                WHEN EXISTS (SELECT 1 FROM existing WHERE user_id = $1) THEN 'duplicate'::text
                WHEN EXISTS (SELECT 1 FROM existing) THEN 'conflict'::text
                ELSE 'not_found'::text
            END as result`)).
		WithArgs(userID, orderNumber, models.OrderStatusNew).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow("conflict"))

	err = storage.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Equal(t, handler.ErrOtherUserOrder, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Ошибка базы данных при создании заказа
func TestPostgresStorage_CreateOrder_DatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1
	orderNumber := "12345678903"

	mock.ExpectQuery(regexp.QuoteMeta(`
        WITH inserted AS (
            INSERT INTO orders (user_id, number, status) 
            VALUES ($1, $2, $3)
            ON CONFLICT (number) DO NOTHING
            RETURNING uid
        ),
        existing AS (
            SELECT user_id FROM orders WHERE number = $2
        )
        SELECT 
            CASE 
                WHEN EXISTS (SELECT 1 FROM inserted) THEN 'inserted'::text
                WHEN EXISTS (SELECT 1 FROM existing WHERE user_id = $1) THEN 'duplicate'::text
                WHEN EXISTS (SELECT 1 FROM existing) THEN 'conflict'::text
                ELSE 'not_found'::text
            END as result`)).
		WithArgs(userID, orderNumber, models.OrderStatusNew).
		WillReturnError(fmt.Errorf("database error"))

	err = storage.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create order")
	assert.NoError(t, mock.ExpectationsWereMet())
}

// Неожиданный результат
func TestPostgresStorage_CreateOrder_UnexpectedResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer db.Close()

	storage := newTestStorage(db)

	userID := 1
	orderNumber := "12345678903"

	mock.ExpectQuery(regexp.QuoteMeta(`
        WITH inserted AS (
            INSERT INTO orders (user_id, number, status) 
            VALUES ($1, $2, $3)
            ON CONFLICT (number) DO NOTHING
            RETURNING uid
        ),
        existing AS (
            SELECT user_id FROM orders WHERE number = $2
        )
        SELECT 
            CASE 
                WHEN EXISTS (SELECT 1 FROM inserted) THEN 'inserted'::text
                WHEN EXISTS (SELECT 1 FROM existing WHERE user_id = $1) THEN 'duplicate'::text
                WHEN EXISTS (SELECT 1 FROM existing) THEN 'conflict'::text
                ELSE 'not_found'::text
            END as result`)).
		WithArgs(userID, orderNumber, models.OrderStatusNew).
		WillReturnRows(sqlmock.NewRows([]string{"result"}).AddRow("unknown"))

	err = storage.CreateOrder(userID, orderNumber)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected result")
	assert.NoError(t, mock.ExpectationsWereMet())
}
