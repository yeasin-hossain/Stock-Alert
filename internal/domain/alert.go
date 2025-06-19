package domain

import "github.com/hello-api/internal/handler/dto"

// AlertRepository interface defines the contract for alert data operations
type AlertRepository interface {
	Create(alert *dto.AlertCreateRequest) (*dto.AlertResponse, error)
	FindByID(id string) (*dto.AlertResponse, error)
	FindAllByUser(userId string) ([]dto.AlertResponse, error)
	Update(id string, alert *dto.AlertCreateRequest) (*dto.AlertResponse, error)
	Delete(id string) error
}

type AlertService interface {
	CreateAlert(alert dto.AlertCreateRequest) (*dto.AlertResponse, error)
	GetAlertByID(id string) (*dto.AlertResponse, error)
	GetAlertsByUser(userId string) ([]dto.AlertResponse, error)
	UpdateAlert(id string, alert dto.AlertCreateRequest) (*dto.AlertResponse, error)
	DeleteAlert(id string) error
}
