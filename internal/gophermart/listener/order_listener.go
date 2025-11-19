// internal/gophermart/listener/order_listener.go
package listener

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/lib/pq"
)

type OrderListener struct {
	dbURI                string
	logger               *zap.SugaredLogger
	workers              *WorkerPool
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
	var err error
	ol.db, err = sql.Open("postgres", ol.dbURI)
	if err != nil {
		ol.logger.Fatalf("Failed to open database: %v", err)
	}

	// Используем NewWorkerPool
	ol.workers = NewWorkerPool(10, 100, 1000, ol.processOrder, ol.logger)
	ol.workers.Start(ctx)

	go ol.loadExistingOrders(ctx)
	go ol.listenNotifications(ctx)
}

func (ol *OrderListener) loadExistingOrders(ctx context.Context) {
	ol.logger.Info("Loading existing orders with status NEW")

	offset := 0
	batchSize := 100

	for {
		select {
		case <-ctx.Done():
			ol.logger.Info("Stopping existing orders loading")
			return
		default:
			jobs, err := ol.fetchOrdersBatch(ctx, offset, batchSize)
			if err != nil {
				ol.logger.Errorf("Failed to fetch orders batch: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if len(jobs) == 0 {
				ol.logger.Info("Finished loading existing orders")
				return
			}

			ol.workers.SubmitBatch(jobs)
			ol.logger.Infof("Loaded batch of %d existing orders, total processed: %d", len(jobs), offset+len(jobs))

			offset += len(jobs)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (ol *OrderListener) fetchOrdersBatch(ctx context.Context, offset, limit int) ([]Job, error) {
	query := `
		SELECT id, user_id, number, status, created_at 
		FROM orders 
		WHERE status = $1 
		ORDER BY created_at ASC 
		LIMIT $2 OFFSET $3`

	rows, err := ol.db.QueryContext(ctx, query, "NEW", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var job Job
		err := rows.Scan(&job.OrderID, &job.UserID, &job.Number, &job.Status, &job.CreatedAt)
		if err != nil {
			ol.logger.Errorf("Failed to scan order: %v", err)
			continue
		}
		job.Attempt = 0
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (ol *OrderListener) listenNotifications(ctx context.Context) {
	ol.logger.Info("Starting PostgreSQL notifications listener")

	for {
		select {
		case <-ctx.Done():
			ol.logger.Info("Notifications listener stopped")
			return
		default:
			if err := ol.listen(ctx); err != nil {
				ol.logger.Errorf("Notification listener error: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (ol *OrderListener) listen(ctx context.Context) error {
	listener := pq.NewListener(ol.dbURI, 10*time.Second, time.Minute, ol.handleListenerEvent)
	defer listener.Close()

	err := listener.Listen("new_orders")
	if err != nil {
		return fmt.Errorf("failed to listen on channel: %w", err)
	}

	ol.logger.Info("Listening for notifications on channel 'new_orders'")

	for {
		select {
		case <-ctx.Done():
			ol.logger.Info("Stopping notification listener")
			return nil
		case notification, ok := <-listener.Notify:
			if !ok {
				ol.logger.Info("Notification channel closed")
				return fmt.Errorf("notification channel closed")
			}

			if notification != nil {
				ol.logger.Infof("Received notification: %s", notification.Extra)
				ol.handleNotification(notification.Extra)
			}
		case <-time.After(15 * time.Second):
			go func() {
				if err := listener.Ping(); err != nil {
					ol.logger.Errorf("Database ping failed: %v", err)
				}
			}()
		}
	}
}

func (ol *OrderListener) handleListenerEvent(event pq.ListenerEventType, err error) {
	switch event {
	case pq.ListenerEventConnected:
		ol.logger.Info("PostgreSQL listener connected")
	case pq.ListenerEventDisconnected:
		ol.logger.Warn("PostgreSQL listener disconnected")
	case pq.ListenerEventReconnected:
		ol.logger.Info("PostgreSQL listener reconnected")
	case pq.ListenerEventConnectionAttemptFailed:
		ol.logger.Errorf("PostgreSQL listener connection attempt failed: %v", err)
	}
}

func (ol *OrderListener) handleNotification(payload string) {
	var job Job
	if err := json.Unmarshal([]byte(payload), &job); err != nil {
		ol.logger.Errorf("Failed to unmarshal notification: %v", err)
		return
	}

	job.Attempt = 0
	job.CreatedAt = time.Now()

	ol.workers.Submit(job)
	ol.logger.Debugf("Notification job submitted to buffer, current buffer size: %d",
		ol.workers.GetBufferSize())
}

func (ol *OrderListener) processOrder(ctx context.Context, job Job) error {
	ol.logger.Infof("Processing order: ID=%d, Number=%s, UserID=%d, Attempt=%d",
		job.OrderID, job.Number, job.UserID, job.Attempt)

	// Циклически опрашиваем сервис пока не получим финальный статус
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("processing cancelled for order %s", job.Number)
		default:
			// 1. Делаем запрос к внешнему сервису
			accrualResult, err := ol.queryAccrualService(ctx, job.Number)
			if err != nil {
				return fmt.Errorf("failed to query accrual service: %w", err)
			}

			// 2. Обновляем статус в БД
			err = ol.updateOrderStatus(ctx, job.OrderID, accrualResult.Status, accrualResult.Accrual)
			if err != nil {
				return fmt.Errorf("failed to update order status: %w", err)
			}

			ol.logger.Infof("Order %s status updated to: %s", job.Number, accrualResult.Status)

			// 3. Проверяем финальный ли статус
			if ol.isFinalStatus(accrualResult.Status) {
				ol.logger.Infof("Order %s processing completed with status: %s", job.Number, accrualResult.Status)
				return nil
			}

			// 4. Если статус не финальный, ждем и повторяем запрос
			waitTime := ol.getWaitTime(accrualResult.Status)
			ol.logger.Debugf("Order %s status is %s, waiting %v before next check",
				job.Number, accrualResult.Status, waitTime)

			select {
			case <-ctx.Done():
				return fmt.Errorf("processing cancelled while waiting for order %s", job.Number)
			case <-time.After(waitTime):
				// Продолжаем цикл
			}
		}
	}
}

// isFinalStatus проверяет является ли статус финальным
func (ol *OrderListener) isFinalStatus(status string) bool {
	return status == "PROCESSED" || status == "INVALID"
}

// getWaitTime возвращает время ожидания между запросами
func (ol *OrderListener) getWaitTime(status string) time.Duration {
	switch status {
	case "REGISTERED":
		return 1 * time.Second
	case "PROCESSING":
		return 5 * time.Second
	default:
		return 5 * time.Second
	}
}

// isPermanentError проверяет является ли ошибка постоянной
func (ol *OrderListener) isPermanentError(err error) bool {
	errorStr := err.Error()
	return strings.Contains(errorStr, "order not registered") ||
		strings.Contains(errorStr, "unexpected status code")
}

func (ol *OrderListener) queryAccrualService(ctx context.Context, orderNumber string) (*AccrualResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	accrualAddress := ol.accrualSystemAddress
	if !strings.HasPrefix(accrualAddress, "http://") && !strings.HasPrefix(accrualAddress, "https://") {
		accrualAddress = "http://" + accrualAddress
	}

	url := fmt.Sprintf("%s/api/orders/%s", accrualAddress, orderNumber)
	ol.logger.Debugf("Sending request to accrual service: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var result AccrualResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		ol.logger.Debugf("Received accrual response: %+v", result)
		return &result, nil

	case http.StatusNoContent:
		ol.logger.Warnf("Order %s not registered in accrual system", orderNumber)
		return nil, fmt.Errorf("order not registered in accrual system")

	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		retrySeconds := 60
		if retryAfter != "" {
			if seconds, err := strconv.Atoi(retryAfter); err == nil {
				retrySeconds = seconds
			}
		}
		ol.logger.Warnf("Rate limit exceeded, retry after %d seconds", retrySeconds)
		return nil, fmt.Errorf("rate limit exceeded, retry after %d seconds", retrySeconds)

	case http.StatusInternalServerError:
		ol.logger.Errorf("Accrual service internal error for order %s", orderNumber)
		return nil, fmt.Errorf("accrual service internal error")

	default:
		ol.logger.Errorf("Unexpected status code %d for order %s", resp.StatusCode, orderNumber)
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func (ol *OrderListener) updateOrderStatus(ctx context.Context, orderID int, status string, accrual float64) error {
	query := `
		UPDATE orders 
		SET status = $1, accrual = $2, updated_at = NOW()
		WHERE id = $3`

	result, err := ol.db.ExecContext(ctx, query, status, accrual, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found: %d", orderID)
	}

	return nil
}

func (ol *OrderListener) Stop() {
	if ol.workers != nil {
		ol.workers.Stop()
	}
	if ol.db != nil {
		ol.db.Close()
	}
}
