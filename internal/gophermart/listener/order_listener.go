package listener

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type OrderListener struct {
	dbURI                string
	logger               *zap.SugaredLogger
	db                   *sql.DB
	accrualSystemAddress string
}

func NewOrderListener(dbURI, accrualSystemAddress string, logger *zap.SugaredLogger) *OrderListener {
	return &OrderListener{
		dbURI:                dbURI,
		accrualSystemAddress: accrualSystemAddress,
		logger:               logger,
	}
}

func (ol *OrderListener) Start(ctx context.Context) {
	ol.logger.Info("=== ORDER LISTENER STARTING ===")
	ol.logger.Infof("Raw Database URI: %s", ol.dbURI)
	ol.logger.Infof("Accrual address: %s", ol.accrualSystemAddress)

	// убираем кавычки, если они есть
	dsn := strings.Trim(ol.dbURI, `"`)
	ol.logger.Infof("Sanitized Database URI: %s", dsn)

	// Очистка env чтобы pgx не подхватывал старые переменные
	os.Unsetenv("PGUSER")
	os.Unsetenv("PGPASSWORD")
	os.Unsetenv("PGDATABASE")
	os.Unsetenv("PGHOST")
	os.Unsetenv("PGPORT")

	var err error
	ol.db, err = sql.Open("pgx", dsn)
	if err != nil {
		ol.logger.Fatalf("Failed to open database: %v", err)
	}

	if err := ol.db.Ping(); err != nil {
		ol.logger.Fatalf("Failed to ping database: %v", err)
	}
	ol.logger.Info("Database connection successful")

	// грузим существующие NEW-заказы и сразу запускаем их обработку
	go ol.loadExistingOrders(ctx)

	// слушаем нотификации новых заказов
	go ol.listenNotifications(ctx)
}

// --------------------------------------------
// ЗАГРУЗКА И СЛУШАТЕЛЬ НОВЫХ ЗАКАЗОВ
// --------------------------------------------

func (ol *OrderListener) loadExistingOrders(ctx context.Context) {
	ol.logger.Info("Loading existing NEW orders...")

	rows, err := ol.db.QueryContext(ctx,
		`SELECT uid, user_id, number, status, uploaded_at 
         FROM orders WHERE status='NEW' ORDER BY uploaded_at ASC`)
	if err != nil {
		ol.logger.Errorf("Load failed: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var job Job
		if err := rows.Scan(&job.OrderID, &job.UserID, &job.Number, &job.Status, &job.CreatedAt); err != nil {
			ol.logger.Errorf("Row scan failed: %v", err)
			continue
		}
		job.Attempt = 0

		// ВАЖНО: запускаем обработку заказа в отдельной горутине, без WorkerPool
		go func(j Job) {
			if err := ol.processOrder(ctx, j); err != nil {
				ol.logger.Errorf("Failed to process existing order %s: %v", j.Number, err)
			}
		}(job)

		count++
	}

	ol.logger.Infof("Existing orders loaded: %d", count)
}

func (ol *OrderListener) listenNotifications(ctx context.Context) {
	cfg, err := pgxpool.ParseConfig(strings.Trim(ol.dbURI, `"`))
	if err != nil {
		ol.logger.Errorf("Failed to parse pgx config: %v", err)
		return
	}

	conn, err := pgx.ConnectConfig(ctx, cfg.ConnConfig)
	if err != nil {
		ol.logger.Errorf("Failed to connect to Postgres for LISTEN: %v", err)
		return
	}
	defer conn.Close(ctx)

	if _, err := conn.Exec(ctx, "LISTEN new_orders"); err != nil {
		ol.logger.Errorf("Failed LISTEN: %v", err)
		return
	}

	ol.logger.Info("Listening for new_orders notifications")

	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			// если контекст отменён — выходим
			if ctx.Err() != nil {
				ol.logger.Infof("Listen context canceled, stop listening")
				return
			}
			ol.logger.Warnf("WaitForNotification error, retrying in 5s: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		var job Job
		if err := json.Unmarshal([]byte(n.Payload), &job); err != nil {
			ol.logger.Warnf("Failed to parse notification payload: %v", err)
			continue
		}

		job.Attempt = 0
		ol.logger.Infof("New order notification received: %s", job.Number)

		// Снова — запускаем обработку без WorkerPool
		go func(j Job) {
			if err := ol.processOrder(ctx, j); err != nil {
				ol.logger.Errorf("Failed to process notified order %s: %v", j.Number, err)
			}
		}(job)
	}
}

// --------------------------------------------
// ОБРАБОТКА ЗАКАЗА
// --------------------------------------------

func (ol *OrderListener) processOrder(ctx context.Context, job Job) error {
	ol.logger.Infof("Start processing order %s (uid=%d)", job.Number, job.OrderID)

	for {
		select {
		case <-ctx.Done():
			ol.logger.Infof("Processing cancelled for order %s", job.Number)
			return fmt.Errorf("cancelled")
		default:
		}

		result, err := ol.queryAccrualService(ctx, job.Number)
		if err != nil {
			ol.logger.Warnf("Accrual service error for order %s: %v", job.Number, err)
			time.Sleep(2 * time.Second)
			continue
		}

		if result != nil {
			ol.logger.Infof("Accrual result for order %s: %+v", job.Number, result)
			if err := ol.updateOrderStatus(ctx, job.OrderID, result.Status, result.Accrual); err != nil {
				ol.logger.Errorf("Failed to update order %s: %v", job.Number, err)
				return err
			}

			if result.Status == "PROCESSED" || result.Status == "INVALID" {
				ol.logger.Infof("Order %s reached final status %s", job.Number, result.Status)
				return nil
			}
		}

		time.Sleep(2 * time.Second)
	}
}

// --------------------------------------------
// HTTP запрос к Accrual
// --------------------------------------------

func (ol *OrderListener) queryAccrualService(ctx context.Context, number string) (*AccrualResponse, error) {
	client := http.Client{Timeout: 20 * time.Second}

	addr := ol.accrualSystemAddress
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}

	url := fmt.Sprintf("%s/api/orders/%s", addr, number)
	ol.logger.Infof("Querying accrual service: %s", url)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var r AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			return nil, fmt.Errorf("Failed to decode accrual response: %w", err)
		}
		return &r, nil

	case http.StatusTooManyRequests:
		ra := resp.Header.Get("Retry-After")
		sec, _ := strconv.Atoi(ra)
		if sec == 0 {
			sec = 60
		}
		ol.logger.Warnf("Rate limited, retrying after %d seconds", sec)
		time.Sleep(time.Duration(sec) * time.Second)
		return nil, fmt.Errorf("rate limit")

	case http.StatusNoContent:
		ol.logger.Infof("Accrual service: order %s not yet registered", number)
		return nil, nil

	default:
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
}

// --------------------------------------------
// ОБНОВЛЕНИЕ СТАТУСА
// --------------------------------------------

func (ol *OrderListener) updateOrderStatus(ctx context.Context, uid int, status string, accrual float64) error {
	res, err := ol.db.ExecContext(ctx,
		`UPDATE orders SET status=$1, accrual=$2, uploaded_at=NOW() WHERE uid=$3`,
		status, accrual, uid)
	if err != nil {
		return fmt.Errorf("DB update failed: %w", err)
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("order not found uid=%d", uid)
	}

	ol.logger.Infof("Order %d updated: status=%s, accrual=%.2f", uid, status, accrual)
	return nil
}

// --------------------------------------------
// ОСТАНОВКА
// --------------------------------------------

func (ol *OrderListener) Stop() {
	if ol.db != nil {
		ol.db.Close()
	}
	ol.logger.Info("Order listener stopped")
}
