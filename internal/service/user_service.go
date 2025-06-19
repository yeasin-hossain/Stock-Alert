package service

import (
	"github.com/hello-api/internal/domain"
)

type UserService struct {
	repo domain.UserRepository
}

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) GetAllUsers() ([]domain.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) GetUserByID(id int) (domain.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) CreateUser(user domain.User) (domain.User, error) {
	// Here you could add business logic, validation, etc.
	return s.repo.Create(user)
}

func (s *UserService) UpdateUser(user domain.User) (domain.User, error) {
	// Here you could add business logic, validation, etc.
	return s.repo.Update(user)
}

func (s *UserService) DeleteUser(id int) error {
	return s.repo.Delete(id)
}
