// internal/service/service.go
package service

import (
	"fmt"
	"go-musthave-diploma-tpl/internal/gophermart/models"
)

// GofemartRepo - интерфейс репозитория
type GofemartRepo interface {
	// получение пользователя по логину и паролю
	GetUserByLoginAndPassword(login, password string) (*models.User, error)
	// создание пользователя
	CreateUser(login, password string) (*models.User, error)
	// получаем пользователя по ID
	GetUserByID(id int) (*models.User, error)
	// создание и проверка заказа
	CreateOrder(userID int, orderNumber string) error
	// получение заказов по пользвователю
	GetOrders(userID int) ([]models.Order, error)
	// получение баланса
	GetBalance(userID int) (models.Balance, error)
	// запрос на списание средств
	Withdraw(userID int, withdraw models.WithdrawBalance) error
	// получение списка информации о выводе средств
	Withdrawals(userID int) ([]models.WithdrawBalance, error)
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
	if login == "" || password == "" {
		return nil, fmt.Errorf("login and password are required")
	}

	user, err := s.repo.GetUserByLoginAndPassword(login, password)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, fmt.Errorf("Invalid login or password")
	}

	return user, nil
}

func (s *GofemartService) GetUserByID(userID int) (*models.User, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("Invalid user ID")
	}
	return s.repo.GetUserByID(userID)
}

// CreateOrder - создание нового заказа
func (s *GofemartService) CreateOrder(userID int, orderNumber string) error {
	if userID <= 0 {
		return fmt.Errorf("Invalid user ID")
	}

	if orderNumber == "" {
		return fmt.Errorf("Order number is required")
	}

	return s.repo.CreateOrder(userID, orderNumber)
}

func (s *GofemartService) GetOrders(userID int) ([]models.Order, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("Invalid user ID")
	}
	return s.repo.GetOrders(userID)
}

func (s *GofemartService) GetBalance(userID int) (models.Balance, error) {
	if userID <= 0 {
		return models.Balance{}, fmt.Errorf("Invalid user ID")
	}
	return s.repo.GetBalance(userID)
}

func (s *GofemartService) Withdraw(userID int, withdraw models.WithdrawBalance) error {
	return s.repo.Withdraw(userID, withdraw)
}

func (s *GofemartService) Withdrawals(userID int) ([]models.WithdrawBalance, error) {
	return s.repo.Withdrawals(userID)
}
