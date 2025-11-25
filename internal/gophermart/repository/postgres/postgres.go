package postgres

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	db "go-musthave-diploma-tpl/internal/gophermart/config/db"
	handler "go-musthave-diploma-tpl/internal/gophermart/handler"
	"go-musthave-diploma-tpl/internal/gophermart/models"
	logger "go-musthave-diploma-tpl/pkg/runtime/logger"
)

var castomLogger = logger.NewHTTPLogger().Sugar()

type PostgresStorage struct {
	DB              *sql.DB
	errorClassifier *PostgresErrorClassifier
}

func New() *PostgresStorage {
	return &PostgresStorage{
		DB:              db.DB,
		errorClassifier: NewPostgresErrorClassifier(),
	}
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (ps *PostgresStorage) GetUserByLogin(login string) (*models.User, error) {
	var user models.User

	query := `SELECT 
					id, 
					login, 
					password_hash, 
					created_at 
				FROM users 
				WHERE login = $1`
	err := ps.DB.QueryRow(query, login).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by login: %w", err)
	}

	return &user, nil
}

func (ps *PostgresStorage) GetUserByID(id int) (*models.User, error) {
	var user models.User
	query := `SELECT id, login, password_hash, created_at FROM users WHERE id = $1`

	err := ps.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		castomLogger.Infof("failed to get user by ID: %v", err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

func (ps *PostgresStorage) GetUserByLoginAndPassword(login, password string) (*models.User, error) {
	var user models.User
	hashedPassword := HashPassword(password)

	query := `SELECT 
					id, 
					login, 
					password_hash, 
					created_at 
				FROM users 
				WHERE login = $1 
					AND password_hash = $2`
	err := ps.DB.QueryRow(query, login, hashedPassword).Scan(&user.ID, &user.Login, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		castomLogger.Infof("failed to get user by login and password: %v", err)
		return nil, fmt.Errorf("failed to get user by login and password: %w", err)
	}

	return &user, nil
}

func (ps *PostgresStorage) CreateUser(login, password string) (*models.User, error) {
	existingUser, err := ps.GetUserByLogin(login)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		return nil, errors.New("login already exists")
	}

	hashedPassword := HashPassword(password)

	var user models.User
	query := `INSERT INTO users (login, password_hash) 
              VALUES ($1, $2) 
              RETURNING id, login, password_hash, created_at`

	err = ps.DB.QueryRow(query, login, hashedPassword).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		castomLogger.Infof("failed to create user: %v", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}

func (ps *PostgresStorage) CreateOrder(userID int, orderNumber string) error {
	query := `
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
            END as result`

	var result string
	if err := ps.DB.QueryRow(query, userID, orderNumber, models.OrderStatusNew).Scan(&result); err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	switch result {
	case "inserted":
		// триггер сам отправит notify
		return nil
	case "duplicate":
		return handler.ErrDuplicateOrder
	case "conflict":
		return handler.ErrOtherUserOrder
	default:
		return fmt.Errorf("unexpected result: %s", result)
	}
}

func (ps *PostgresStorage) GetOrders(userID int) ([]models.Order, error) {
	rows, err := ps.DB.Query(`
        SELECT number, status, accrual, uploaded_at 
        FROM orders WHERE user_id = $1 
        ORDER BY uploaded_at DESC`, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (ps *PostgresStorage) GetBalance(userID int) (models.Balance, error) {
	var balance models.Balance

	query := `
        SELECT
            COALESCE((
                SELECT SUM(accrual)
                FROM orders
                WHERE user_id = $1 AND status = 'PROCESSED'
            ), 0)
            -
            COALESCE((
                SELECT SUM(sum)
                FROM withdrawals
                WHERE user_id = $1
            ), 0) AS current,
            COALESCE((
                SELECT SUM(sum)
                FROM withdrawals
                WHERE user_id = $1
            ), 0) AS withdrawn
    `

	err := ps.DB.QueryRow(query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err == sql.ErrNoRows {
		return models.Balance{Current: 0, Withdrawn: 0}, nil
	}
	if err != nil {
		return models.Balance{}, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

func (ps *PostgresStorage) Withdraw(userID int, withdraw models.WithdrawBalance) error {
	tx, err := ps.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// считаем баланс как в GetBalance
	var balance float64
	err = tx.QueryRow(`
        SELECT 
            COALESCE((
                SELECT SUM(accrual) 
                FROM orders 
                WHERE user_id = $1 AND status = 'PROCESSED'
            ), 0) 
            - COALESCE((
                SELECT SUM(sum) 
                FROM withdrawals 
                WHERE user_id = $1
            ), 0) AS balance
    `, userID).Scan(&balance)
	if err != nil {
		return err
	}

	if balance-withdraw.Sum < 0 {
		return handler.ErrLackOfFunds
	}

	// просто пишем факт списания, без проверки, что заказ существует в orders
	_, err = tx.Exec(`
        INSERT INTO withdrawals (user_id, order_number, sum)
        VALUES ($1, $2, $3)
    `, userID, withdraw.Order, withdraw.Sum)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (ps *PostgresStorage) Withdrawals(userID int) ([]models.WithdrawBalance, error) {
	rows, err := ps.DB.Query(`	SELECT 
									order_number,
									sum,
									processed_at
								FROM withdrawals
								WHERE user_id = $1
								ORDER BY processed_at DESC
							`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []models.WithdrawBalance
	for rows.Next() {
		var w models.WithdrawBalance
		if err := rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}
