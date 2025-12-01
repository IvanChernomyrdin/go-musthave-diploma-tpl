package service

import (
	"context"
	"fmt"
	"go-musthave-diploma-tpl/internal/accrual/models"
	luhn "go-musthave-diploma-tpl/pkg"
	"strconv"
	"time"

	"go.uber.org/zap"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go -package=mock_service
type Repository interface {
	CreateProductReward(ctx context.Context, match string, reward float64, rewardType string) error
	RegisterNewOrder(ctx context.Context, order int64, goods []models.Goods, status string) error
	CheckOrderExists(order int64) (bool, error)
	GetAccrualInfo(order int64) (string, float64, error)
	UpdateAccrualInfo(ctx context.Context, order int64, accrual float64, status string) error
	UpdateStatus(ctx context.Context, status string, order int64) error
	GetProductsInfo() ([]models.ProductReward, error)
	ParseMatch(match string) ([]models.ParseMatch, error)
	GetUnprocessedOrders() ([]int64, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateProductReward(ctx context.Context, match string, reward float64, rewardType string) error {
	return s.repo.CreateProductReward(ctx, match, reward, rewardType)
}

func (s *Service) RegisterNewOrder(ctx context.Context, order models.Order) (bool, error) {
	var exist bool
	number, err := strconv.ParseInt(order.Order, 10, 64)
	if err != nil {
		return exist, err
	}
	exist, err = s.repo.CheckOrderExists(number)
	if err != nil {
		return exist, err
	}

	if err := s.repo.RegisterNewOrder(ctx, number, order.Goods, models.Registered); err != nil {
		return exist, err
	}
	return exist, nil
}

func (s *Service) GetAccrualInfo(order int64) (string, float64, bool, error) {
	exist, err := s.repo.CheckOrderExists(order)
	if err != nil {
		return "", 0, exist, err
	}
	if !exist {
		return "", 0, exist, nil
	}
	status, accrual, err := s.repo.GetAccrualInfo(order)
	return status, accrual, exist, err
}

// Listener запускает процесс обработки заказов
func (s *Service) Listener(ctx context.Context, log *zap.SugaredLogger, pollingInterval time.Duration) {
	log.Info("Listener started")

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Listener stopped")
			return
		case <-ticker.C:
			if err := s.processOrders(ctx, log); err != nil {
				log.Errorf("Error processing orders: %v", err)
			}
		}
	}
}

// processOrders обрабатывает все заказы
func (s *Service) processOrders(ctx context.Context, log *zap.SugaredLogger) error {

	products, err := s.repo.GetProductsInfo()
	if err != nil {
		return fmt.Errorf("failed to get products info: %w", err)
	}

	processedOrderIDs := make(map[int64]bool)

	for _, product := range products {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		orders, err := s.repo.ParseMatch(product.Match)
		if err != nil {
			log.Errorf("Failed to parse match %s: %v", product.Match, err)
			continue
		}

		orderMap := make(map[int64][]models.ParseMatch)
		for _, order := range orders {
			orderMap[order.Order] = append(orderMap[order.Order], order)
			processedOrderIDs[order.Order] = true
		}

		for orderID, orderItems := range orderMap {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			if !luhn.ValidateLuhn(strconv.FormatInt(orderID, 10)) && !luhn.ValidateLuhn(strconv.FormatInt(orderID, 10)) {
				if err := s.repo.UpdateStatus(ctx, models.Invalid, orderID); err != nil {
					log.Errorf("Failed to update status for order %d: %v", orderID, err)
				}
				continue
			}

			if err := s.updateOrderAccrual(ctx, log, orderID, orderItems, product); err != nil {
				log.Errorf("Failed to update accrual for order %d: %v", orderID, err)
			}
		}
	}

	// Получение всех заказов, которые не соответствуют ни одному правилу начисления бонусов и обновление их статуса на PROCESSED с нулевым начислением бонусов.
	unprocessedOrders, err := s.repo.GetUnprocessedOrders()
	if err != nil {
		log.Errorf("Failed to get unprocessed orders: %v", err)

	} else {
		// Обновление статусов для заказов, которые не соответствуют ни одному правилу начисления бонусов
		for _, orderID := range unprocessedOrders {
			if !processedOrderIDs[orderID] {
				if err := s.repo.UpdateAccrualInfo(ctx, orderID, 0, models.Processed); err != nil {
					log.Errorf("Failed to update accrual for order %d: %v", orderID, err)
				} else {
					log.Infof("Updated order %d to PROCESSED with zero accrual (no matching products)", orderID)
				}
			}
		}
	}

	return nil
}

// updateOrderAccrual обновляет начисление бонусов для заказа
func (s *Service) updateOrderAccrual(ctx context.Context, log *zap.SugaredLogger, orderID int64, orderItems []models.ParseMatch, product models.ProductReward) error {
	var totalAccrual float64

	// Общая сумма начислений для всех товаров в заказе
	for _, item := range orderItems {
		var accrual float64
		if product.RewardType == "pt" {

			accrual = item.Price * product.Reward / 100
		} else {

			accrual = product.Reward
		}
		totalAccrual += accrual
	}

	// Обновление информации о начислениях для заказа
	if err := s.repo.UpdateAccrualInfo(ctx, orderID, totalAccrual, models.Processed); err != nil {
		return fmt.Errorf("failed to update accrual info for order %d: %w", orderID, err)
	}

	log.Infof("Updated accrual for order %d: %.2f", orderID, totalAccrual)
	return nil
}
