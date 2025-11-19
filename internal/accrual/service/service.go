package service

type Repository interface {
	CreateProductReward(match string, reward float64, rewardType string) error
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
