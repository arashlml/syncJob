package service

import (
	"practice/internals/entity"
	"practice/internals/params"
)

type Repository interface {
	GetByID(id string) (*entity.User, error)
}

type UserService struct {
	repo Repository
}

func NewUserService(repo Repository) *UserService {
	return &UserService{repo: repo}
}

func (us *UserService) GetUserByID(userID params.UserID) (*entity.User, error) {
	user, err := us.repo.GetByID(userID.ID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
