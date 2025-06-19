package entity

import (
	"time"
)

// AlertStatus and AlertRule enums
type AlertStatus string
type AlertRule string

const (
	AlertStatusActive   AlertStatus = "active"
	AlertStatusInactive AlertStatus = "inactive"

	AlertRuleAbove AlertRule = "above"
	AlertRuleBelow AlertRule = "below"
)

// AlertEntity represents the alert as stored in the database
type AlertEntity struct {
	ID        string      `bson:"_id,omitempty" json:"id"`
	Name      string      `bson:"name" json:"name"`
	Price     float64     `bson:"price" json:"price"`
	Rule      AlertRule   `bson:"rule" json:"rule"`
	StopDate  time.Time   `bson:"stopDate" json:"stopDate"`
	StartDate time.Time   `bson:"startDate" json:"startDate"`
	Status    AlertStatus `bson:"status" json:"status"`
	UserID    string      `bson:"userId" json:"userId"`
	CreatedAt time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time   `bson:"updated_at" json:"updated_at"`
}
