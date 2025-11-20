package postgres

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	db "go-musthave-diploma-tpl/internal/gophermart/config/db"
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
	// проверяем что такого логина нет (используем GetUserByLogin)
	existingUser, err := ps.GetUserByLogin(login)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("login already exists")
	}

	hashedPassword := HashPassword(password)

	var user models.User
	// добавляем
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
            RETURNING user_id
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
	err := ps.DB.QueryRow(query, userID, orderNumber, models.OrderStatusNew).Scan(&result)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	switch result {
	case "inserted":
		return nil
	case "duplicate":
		return models.ErrDuplicateOrder
	case "conflict":
		return models.ErrOtherUserOrder
	case "not_found":
		return fmt.Errorf("order not found after conflict")
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
	query := `SELECT
		(SELECT COALESCE(SUM(accrual), 0) FROM orders WHERE user_id = $1 AND status = 'PROCESSED') as current,
		(SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1) as withdrawn`

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
	// работаем через транзакцию
	tx, err := ps.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// проверяем что заказ  есть по номеру
	var exists bool
	err = tx.QueryRow(`SELECT EXISTS(
									SELECT 1 
									FROM orders 
									WHERE number = $1)
						`, withdraw.Order).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return models.ErrInvalidOrderNumber
	}
	// получаем баланс пользователя
	var balance float64
	err = tx.QueryRow(`SELECT 
        					COALESCE((
            					SELECT SUM(accrual) FROM orders WHERE user_id = $1
        					), 0) - 
        					COALESCE((
            					SELECT SUM(sum) FROM withdrawals WHERE user_id = $1
        					), 0) AS balance`, userID).Scan(&balance)
	if err != nil {
		return err
	}
	// проверяем хватает ли баллов
	if balance-withdraw.Sum < 0 {
		return models.ErrLackOfFunds
	}

	// если хватает записываем новое списание
	_, err = tx.Exec(`INSERT INTO withdrawals
							(user_id, order_number, sum)
						VALUES
							($1, $2, $3)`, userID, withdraw.Order, withdraw.Sum)
	if err != nil {
		return err
	}

	// коммитим
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
