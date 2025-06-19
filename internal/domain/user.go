package domain

import (
	"github.com/hello-api/internal/handler/dto"
	"github.com/hello-api/internal/repository/entity"
)

// UserRepository interface defines the contract for user data operations
type UserRepository interface {
	FindAll() ([]entity.UserEntity, error)
	FindByObjectID(id string) (*entity.UserEntity, error)
	Create(user *entity.UserEntity) (*entity.UserEntity, error)
	Update(user *entity.UserEntity) (*entity.UserEntity, error)
	DeleteByObjectID(id string) error
}

// UserService defines the contract for the user service
type UserService interface {
	GetAllUsers() ([]dto.UserResponse, error)
	GetUserByID(id string) (*dto.UserResponse, error)
	CreateUser(user dto.UserCreateRequest) (*dto.UserResponse, error)
	UpdateUser(id string, user dto.UserUpdateRequest) (*dto.UserResponse, error)
	DeleteUser(id string) error
}
