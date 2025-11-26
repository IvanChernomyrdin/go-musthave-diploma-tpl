package service

import (
	"fmt"
	"go-musthave-diploma-tpl/internal/accrual/models"
	luhn "go-musthave-diploma-tpl/pkg"
	"log"
)

type Repository interface {
	CreateProductReward(match string, reward float64, rewardType string) error
	RegisterNewOrder(order int64, goods []models.Goods, status string) error
	CheckOrderExists(order int64) (bool, error)
	GetAccrualInfo(order int64) (string, float64, error)
	UpdateAccrualInfo(order int64, accrual float64, status string) error
	UpdateStatus(status string, order int64) error
	GetProductsInfo() ([]models.ProductReward, error)
	ParseMatch(match string) ([]models.ParseMatch, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProductReward(match string, reward float64, rewardType string) error {
	return s.repo.CreateProductReward(match, reward, rewardType)
}

func (s *Service) RegisterNewOrder(order models.Order) (bool, error) {
	var exist bool
	if exist, err := s.repo.CheckOrderExists(order.Order); err != nil {
		return exist, err
	}

	if err := s.repo.RegisterNewOrder(order.Order, order.Goods, models.Registered); err != nil {
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
		return "", 0, exist, fmt.Errorf("order is not exist")

	}
	status, accrual, err := s.repo.GetAccrualInfo(order)
	return status, accrual, exist, err
}

func (s *Service) Listener() {
	ParseMatch := make(chan models.ParseMatch)
	reward := make(chan models.ProductReward)
	go s.UpdateStatus(ParseMatch, reward)
	for {
		s.ParseMatch(ParseMatch, reward)
	}

}

// ParseMatch получаю информацию о бонусах и провожу парсиг заказов, если номер заказа не валидный, обновляю статус. Далее отправляю данные в каналы.
func (s *Service) ParseMatch(ParseMatch chan<- models.ParseMatch, reward chan<- models.ProductReward) {
	var ProductReward []models.ProductReward
	ProductReward, err := s.repo.GetProductsInfo()
	if err != nil {
		log.Println(err)
	}
	for _, product := range ProductReward {

		orders, err := s.repo.ParseMatch(product.Match)
		if err != nil {
			log.Println(err)
			continue

		}

		for _, v := range orders {
			if !luhn.ValidateLuhn(string(v.Order)) {
				s.repo.UpdateStatus(models.Invalid, v.Order)
				continue
			}
			ParseMatch <- v
			reward <- product
		}

	}
}

// UpdateStatus Обновляю статус и бонусы зказа.
func (s *Service) UpdateStatus(ParseMatch <-chan models.ParseMatch, ProductReward <-chan models.ProductReward) {
	accrual := make(chan float64)
	reward := make(chan float64)
	defer close(accrual)
	defer close(reward)
	for v := range ProductReward {
		if v.RewardType == "pt" {
			accrual <- v.Reward
		} else {
			reward <- v.Reward
		}
	}

	for v := range ParseMatch {

		select {
		case accrual := <-accrual:
			err := s.repo.UpdateAccrualInfo(v.Order, accrual, models.Processed)
			if err != nil {
				log.Println(err)
			}
		case reward := <-reward:
			accrual := v.Price / 100 * reward
			err := s.repo.UpdateAccrualInfo(v.Order, accrual, models.Processed)
			if err != nil {
				log.Println(err)
			}
		}

	}

}
