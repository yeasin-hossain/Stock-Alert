package entity

import (
	"time"
	
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserEntity represents the user as stored in the database
type UserEntity struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    int               `bson:"user_id"`
	Name      string            `bson:"name"`
	Email     string            `bson:"email"`
	CreatedAt time.Time         `bson:"created_at"`
	UpdatedAt time.Time         `bson:"updated_at"`
}
