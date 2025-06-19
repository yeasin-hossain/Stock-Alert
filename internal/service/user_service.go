package service

import (
	"strings"
	"time"
	
	"github.com/hello-api/internal/domain"
	"github.com/hello-api/internal/handler/dto"
	"github.com/hello-api/internal/repository/entity"
)

type UserService struct {
	repo domain.UserRepository
}

// Ensure UserServiceImpl implements UserService
var _ domain.UserService = (*UserService)(nil)

func NewUserService(repo domain.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

// mapEntityToDTO converts a user entity to a user DTO
func mapEntityToDTO(userEntity *entity.UserEntity) dto.UserResponse {
	return dto.UserResponse{
		ID:        userEntity.ID.Hex(),
		UserID:    userEntity.UserID,
		Name:      userEntity.Name,
		Email:     userEntity.Email,
		CreatedAt: userEntity.CreatedAt,
		UpdatedAt: userEntity.UpdatedAt,
	}
}

// GetAllUsers retrieves all users and returns them as DTOs
func (s *UserService) GetAllUsers() ([]dto.UserResponse, error) {
	userEntities, err := s.repo.FindAll()
	if err != nil {
		return nil, err
	}
	
	var userDTOs []dto.UserResponse
	for _, entity := range userEntities {
		userDTOs = append(userDTOs, mapEntityToDTO(&entity))
	}
	
	return userDTOs, nil
}

// GetUserByID retrieves a user by ID and returns it as a DTO
func (s *UserService) GetUserByID(id string) (*dto.UserResponse, error) {
	userEntity, err := s.repo.FindByObjectID(id)
	if err != nil {
		return nil, err
	}
	if userEntity == nil {
		return nil, nil
	}
	response := mapEntityToDTO(userEntity)
	return &response, nil
}

// CreateUser creates a new user from a DTO and returns a response DTO
func (s *UserService) CreateUser(userDTO dto.UserCreateRequest) (*dto.UserResponse, error) {
	// Validate required fields
	if userDTO.Name == "" || userDTO.Email == "" || userDTO.UserID == "" {
		return nil, domain.ErrValidation
	}
	userID := strings.ToLower(userDTO.UserID)
	// Efficiently check if userId exists in DB
	existing, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrValidation // UserID already exists
	}
	// Create entity from DTO
	userEntity := &entity.UserEntity{
		UserID: userID,
		Name:  userDTO.Name,
		Email: userDTO.Email,
	}
	
	// Save to repository
	createdEntity, err := s.repo.Create(userEntity)
	if err != nil {
		return nil, err
	}
	
	// Convert back to DTO
	response := mapEntityToDTO(createdEntity)
	return &response, nil
}

// UpdateUser updates an existing user from a DTO and returns a response DTO
func (s *UserService) UpdateUser(id string, userDTO dto.UserUpdateRequest) (*dto.UserResponse, error) {
	// First, get the existing user
	existingEntity, err := s.repo.FindByObjectID(id)
	if err != nil {
		return nil, err
	}
	if existingEntity == nil {
		return nil, domain.ErrUserNotFound
	}

	// Update only the provided fields
	if userDTO.Name != "" {
		existingEntity.Name = userDTO.Name
	}
	if userDTO.Email != "" {
		existingEntity.Email = userDTO.Email
	}
	
	existingEntity.UpdatedAt = time.Now()

	// Save to repository
	updatedEntity, err := s.repo.Update(existingEntity)
	if err != nil {
		return nil, err
	}
	
	// Convert back to DTO
	response := mapEntityToDTO(updatedEntity)
	return &response, nil
}

// DeleteUser deletes a user by ID
func (s *UserService) DeleteUser(id string) error {
	// You could add additional business logic here
	// For example, check if the user has related data before deleting
	return s.repo.DeleteByObjectID(id)
}
