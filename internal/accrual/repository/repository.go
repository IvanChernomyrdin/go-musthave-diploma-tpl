package repository

import (
	"context"
	"go-musthave-diploma-tpl/internal/accrual/models"
)

type Storage interface {
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

type Repository struct {
	storage Storage
}

func NewRepository(storage Storage) *Repository {
	return &Repository{storage: storage}
}

func (r *Repository) CreateProductReward(ctx context.Context, match string, reward float64, rewardType string) error {
	return r.storage.CreateProductReward(ctx, match, reward, rewardType)
}

func (r *Repository) RegisterNewOrder(ctx context.Context, order int64, goods []models.Goods, status string) error {
	return r.storage.RegisterNewOrder(ctx, order, goods, status)
}

func (r *Repository) CheckOrderExists(order int64) (bool, error) {
	return r.storage.CheckOrderExists(order)
}

func (r *Repository) GetAccrualInfo(order int64) (string, float64, error) {
	return r.storage.GetAccrualInfo(order)
}

func (r *Repository) UpdateAccrualInfo(ctx context.Context, order int64, accrual float64, status string) error {
	return r.storage.UpdateAccrualInfo(ctx, order, accrual, status)
}

func (r *Repository) UpdateStatus(ctx context.Context, status string, order int64) error {
	return r.storage.UpdateStatus(ctx, status, order)
}

func (r *Repository) GetProductsInfo() ([]models.ProductReward, error) {
	return r.storage.GetProductsInfo()
}

func (r *Repository) ParseMatch(match string) ([]models.ParseMatch, error) {
	return r.storage.ParseMatch(match)
}

func (r *Repository) GetUnprocessedOrders() ([]int64, error) {
	return r.storage.GetUnprocessedOrders()
}
