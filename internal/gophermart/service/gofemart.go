// internal/service/service.go
package service

import (
	"fmt"
	"go-musthave-diploma-tpl/internal/gophermart/models"
)

// GofemartRepo - интерфейс репозитория
type GofemartRepo interface {
	GetUserByLoginAndPassword(login, password string) (*models.User, error)
	CreateUser(login, password string) (*models.User, error)
	GetUserByID(id int) (*models.User, error)
	// создание и проверка заказа
	CreateOrder(userID int, orderNumber string) error
	// получение заказов по пользвователю
	GetOrders(userID int) ([]models.Order, error)
}

// GofemartService - сервис с бизнес-логикой
type GofemartService struct {
	repo GofemartRepo
}

func NewGofemartService(repo GofemartRepo) *GofemartService {
	return &GofemartService{repo: repo}
}

func (s *GofemartService) RegisterUser(login, password string) (*models.User, error) {
	if login == "" || password == "" {
		return nil, fmt.Errorf("login and password are required")
	}

	user, err := s.repo.CreateUser(login, password)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *GofemartService) LoginUser(login, password string) (*models.User, error) {
	user, err := s.repo.GetUserByLoginAndPassword(login, password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("invalid login or password")
	}

	return user, nil
}

func (s *GofemartService) GetUserByID(userID int) (*models.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("invalid user ID")
	}
	return s.repo.GetUserByID(userID)
}

// CreateOrder - создание нового заказа
func (s *GofemartService) CreateOrder(userID int, orderNumber string) error {
	if userID <= 0 {
		return fmt.Errorf("invalid user ID")
	}

	if orderNumber == "" {
		return fmt.Errorf("order number is required")
	}

	return s.repo.CreateOrder(userID, orderNumber)
}

func (s *GofemartService) GetOrders(userID int) ([]models.Order, error) {
	return s.repo.GetOrders(userID)
}
