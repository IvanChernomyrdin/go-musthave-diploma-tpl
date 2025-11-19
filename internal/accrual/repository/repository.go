package repository

type Storage interface {
	CreateProductReward(match string, reward float64, rewardType string) error
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
