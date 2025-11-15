package service

type GofemartRepo interface{}

type GofemartService struct {
	repo GofemartRepo
}

func NewGofemartService(repo GofemartRepo) *GofemartService {
	return &GofemartService{repo: repo}
}
