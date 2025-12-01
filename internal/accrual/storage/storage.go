package storage

import (
	"context"
	"database/sql"
	"fmt"
	"go-musthave-diploma-tpl/internal/accrual/models"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresDB struct {
	DB *sql.DB
}

func InitPostgresDB(databaseDSN string) (*PostgresDB, error) {
	connection := strings.Trim(databaseDSN, `"`)

	var err error
	db, err := sql.Open("pgx", connection)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("проверка подключения к БД не удалась: %v", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания драйвера миграций: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/accrual/migrations/",
		"postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания миграции для системы лояльности: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("ошибка применения миграций для системы лояльности: %v", err)
	}

	return &PostgresDB{DB: db}, nil
}

func (db *PostgresDB) CheckConnection() error {
	err := db.DB.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

func (db *PostgresDB) CreateProductReward(ctx context.Context, match string, reward float64, rewardType string) error {
	op := "path: internal/accrual/storage/CreateProductReward"
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s starts a transaction err:%w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var exists bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT * FROM products
			WHERE match = $1
		)
	`, match).Scan(&exists)
	if err != nil {
		return fmt.Errorf("%s QueryRow err:%w", op, err)
	}

	if exists {
		return ErrKeyExists
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO products (match, reward, reward_type)
		VALUES ($1, $2, $3)
	`, match, reward, rewardType)
	if err != nil {
		return fmt.Errorf("%s Exec err:%w", op, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s Commit err:%w", op, err)
	}
	return nil
}

func (db *PostgresDB) RegisterNewOrder(ctx context.Context, order int64, goods []models.Goods, status string) error {
	op := "path: internal/accrual/storage/RegisterNewOrder"

	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s starts a transaction err: %w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Build the goods array using proper PostgreSQL composite type syntax
	goodsValues := make([]string, len(goods))
	for i, good := range goods {

		desc := strings.ReplaceAll(good.Description, "'", "''")

		goodsValues[i] = fmt.Sprintf(`"(%s,%f)"`, desc, good.Price)
	}

	// Create the array literal with proper PostgreSQL syntax
	goodsArrayStr := "{" + strings.Join(goodsValues, ",") + "}"

	query := `
        INSERT INTO orders (order_id, goods, status)
        VALUES ($1, $2::goods[], $3)
    `

	_, err = tx.ExecContext(ctx, query, order, goodsArrayStr, status)
	if err != nil {
		return fmt.Errorf("%s Exec err: %w", op, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s Commit err: %w", op, err)
	}
	return nil
}

func (db *PostgresDB) CheckOrderExists(order int64) (bool, error) {
	op := "path: internal/accrual/storage/CheckOrderExists"

	var exists bool
	err := db.DB.QueryRow(`
		SELECT EXISTS (
			SELECT * FROM orders
			WHERE order_id = $1
		)
	`, order).Scan(&exists)
	if err != nil {
		return exists, fmt.Errorf("%s QueryRow err:%w", op, err)
	}

	return exists, nil
}

func (db *PostgresDB) GetAccrualInfo(order int64) (string, float64, error) {
	op := "path: internal/accrual/storage/GetAccrualInfo"
	tx, err := db.DB.Begin()
	if err != nil {
		return "", 0, fmt.Errorf("%s starts a transaction err:%w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var accrual sql.NullFloat64
	var status string
	err = tx.QueryRow(`
		SELECT accrual, status FROM orders
		WHERE order_id = $1
		`, order).Scan(&accrual, &status)
	if err != nil {
		return "", 0, fmt.Errorf("%s QueryRow err:%w", op, err)
	}
	if accrual.Valid {
		return status, accrual.Float64, nil
	}
	return status, 0, nil

}

func (db *PostgresDB) UpdateAccrualInfo(ctx context.Context, order int64, accrual float64, status string) error {
	op := "path: internal/accrual/storage/UpdateAccrualInfo"
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s starts a transaction err:%w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE orders SET accrual = $1, status = $2
		WHERE order_id = $3
	`, accrual, status, order)
	if err != nil {
		return fmt.Errorf("%s Exec err:%w", op, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s Commit err:%w", op, err)
	}
	return nil
}

func (db *PostgresDB) UpdateStatus(ctx context.Context, status string, order int64) error {
	op := "path: internal/accrual/storage/UpdateStatus"
	tx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("%s starts a transaction err:%w", op, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}

	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE orders SET status = $1
		WHERE order_id = $2
	`, status, order)
	if err != nil {
		return fmt.Errorf("%s Exec err:%w", op, err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s Commit err:%v", op, err)
	}
	return nil
}

func (db *PostgresDB) GetProductsInfo() ([]models.ProductReward, error) {
	op := "path: internal/accrual/storage/GetProductsInfo"
	var productsInfo []models.ProductReward

	rows, err := db.DB.Query(`SELECT match, reward, reward_type FROM products`)
	if err != nil {
		return productsInfo, fmt.Errorf("%s Query err:%w", op, err)
	}

	defer rows.Close()
	for rows.Next() {
		var match string
		var reward float64
		var rewardType string
		err = rows.Scan(&match, &reward, &rewardType)
		if err != nil {
			return productsInfo, fmt.Errorf("%s Scan err:%w", op, err)
		}
		productsInfo = append(productsInfo, models.ProductReward{
			Match:      match,
			Reward:     reward,
			RewardType: rewardType,
		})
	}
	if err := rows.Err(); err != nil {
		return productsInfo, fmt.Errorf("%s rows.Err():%w", op, err)
	}
	return productsInfo, nil
}

func (db *PostgresDB) ParseMatch(match string) ([]models.ParseMatch, error) {
	op := "path: internal/accrual/storage/ParseMatch"
	rows, err := db.DB.Query("SELECT order_id, (g).price FROM orders, UNNEST(goods) AS g WHERE (g).description LIKE $1 AND status NOT IN ('INVALID', 'PROCESSED')", fmt.Sprintf("%%%s%%", match))
	if err != nil {
		return []models.ParseMatch{}, fmt.Errorf("%s error executing query:%w", op, err)
	}
	defer rows.Close()
	var parseMatches []models.ParseMatch
	for rows.Next() {
		var order int64
		var price float64
		err = rows.Scan(&order, &price)
		if err != nil {
			return []models.ParseMatch{}, fmt.Errorf("%s error scanning row:%w", op, err)
		}
		parseMatches = append(parseMatches, models.ParseMatch{
			Order: order,
			Price: price,
		})
	}
	if err := rows.Err(); err != nil {
		return parseMatches, fmt.Errorf("%s rows.Err():%w", op, err)
	}
	return parseMatches, nil
}

// GetUnprocessedOrders returns all orders that are not yet processed
func (db *PostgresDB) GetUnprocessedOrders() ([]int64, error) {
	op := "path: internal/accrual/storage/GetUnprocessedOrders"
	var orderIDs []int64

	rows, err := db.DB.Query("SELECT order_id FROM orders WHERE status NOT IN ('INVALID', 'PROCESSED')")
	if err != nil {
		return orderIDs, fmt.Errorf("%s error executing query:%w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var orderID int64
		err = rows.Scan(&orderID)
		if err != nil {
			return orderIDs, fmt.Errorf("%s error scanning row:%w", op, err)
		}
		orderIDs = append(orderIDs, orderID)
	}
	if err := rows.Err(); err != nil {
		return orderIDs, fmt.Errorf("%s rows.Err():%w", op, err)
	}

	return orderIDs, nil
}
