package service

import (
	"github.com/hello-api/internal/domain"
	"github.com/hello-api/internal/handler/dto"
)

type AlertService struct {
	repo domain.AlertRepository
}

func NewAlertService(repo domain.AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

func (s *AlertService) CreateAlert(alert dto.AlertCreateRequest) (*dto.AlertResponse, error) {
	return s.repo.Create(&alert)
}

func (s *AlertService) GetAlertByID(id string) (*dto.AlertResponse, error) {
	return s.repo.FindByID(id)
}

func (s *AlertService) GetAlertsByUser(userId string) ([]dto.AlertResponse, error) {
	return s.repo.FindAllByUser(userId)
}

func (s *AlertService) UpdateAlert(id string, alert dto.AlertCreateRequest) (*dto.AlertResponse, error) {
	return s.repo.Update(id, &alert)
}

func (s *AlertService) DeleteAlert(id string) error {
	return s.repo.Delete(id)
}
