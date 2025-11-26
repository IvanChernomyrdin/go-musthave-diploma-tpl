package repository

import "go-musthave-diploma-tpl/internal/accrual/models"

type Storage interface {
	CreateProductReward(match string, reward float64, rewardType string) error
	RegisterNewOrder(order int64, goods []models.Goods, status string) error
	CheckOrderExists(order int64) (bool, error)
	GetAccrualInfo(order int64) (string, float64, error)
	UpdateAccrualInfo(order int64, accrual int64, status string) error
	UpdateStatus(status string, order int64) error
	GetProductsInfo() ([]models.ProductReward, error)
	ParseMatch(match string) ([]models.ParseMatch, error)
}

type Repository struct {
	storage Storage
}

func NewRepository(storage Storage) *Repository {
	return &Repository{storage: storage}
}

func (r *Repository) CreateProductReward(match string, reward float64, rewardType string) error {
	return r.storage.CreateProductReward(match, reward, rewardType)
}

func (r *Repository) RegisterNewOrder(order int64, goods []models.Goods, status string) error {
	return r.storage.RegisterNewOrder(order, goods, status)
}

func (r *Repository) CheckOrderExists(order int64) (bool, error) {
	return r.storage.CheckOrderExists(order)
}

func (r *Repository) GetAccrualInfo(order int64) (string, float64, error) {
	return r.storage.GetAccrualInfo(order)
}

func (r *Repository) UpdateAccrualInfo(order int64, accrual int64, status string) error {
	return r.storage.UpdateAccrualInfo(order, accrual, status)
}

func (r *Repository) UpdateStatus(status string, order int64) error {
	return r.storage.UpdateStatus(status, order)
}

func (r *Repository) GetProductsInfo() ([]models.ProductReward, error) {
	return r.storage.GetProductsInfo()
}

func (r *Repository) ParseMatch(match string) ([]models.ParseMatch, error) {
	return r.storage.ParseMatch(match)
}
