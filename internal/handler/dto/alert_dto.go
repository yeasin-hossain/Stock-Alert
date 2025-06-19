package dto

import (
	"time"
)

type AlertStatus string
type AlertRule string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusInactive AlertStatus = "inactive"

	AlertRuleAbove AlertRule = "above"
	AlertRuleBelow AlertRule = "below"
)

type AlertCreateRequest struct {
	Name      string      `json:"name"`
	Price     float64     `json:"price"`
	Rule      AlertRule   `json:"rule"`
	StopDate  time.Time   `json:"stopDate"`
	StartDate time.Time   `json:"startDate"`
	Status    AlertStatus `json:"status"`
	UserID    string      `json:"userId"`
}

type AlertResponse struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Price     float64     `json:"price"`
	Rule      AlertRule   `json:"rule"`
	StopDate  time.Time   `json:"stopDate"`
	StartDate time.Time   `json:"startDate"`
	Status    AlertStatus `json:"status"`
	UserID    string      `json:"userId"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
