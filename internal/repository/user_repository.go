package repository

import (
	"context"
	"errors"
	"time"
	
	"github.com/hello-api/internal/repository/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(collection *mongo.Collection) *MongoUserRepository {
	return &MongoUserRepository{
		collection: collection,
	}
}

// FindAll retrieves all user entities
func (r *MongoUserRepository) FindAll() ([]entity.UserEntity, error) {
	var userEntities []entity.UserEntity
	
	opts := options.Find().SetSort(bson.D{{Key: "user_id", Value: 1}})
	cursor, err := r.collection.Find(context.Background(), bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &userEntities); err != nil {
		return nil, err
	}
	
	return userEntities, nil
}

// FindByID retrieves a user entity by ID
func (r *MongoUserRepository) FindByID(id int) (*entity.UserEntity, error) {
	var userEntity entity.UserEntity
	err := r.collection.FindOne(context.Background(), bson.M{"user_id": id}).Decode(&userEntity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found, but not an error
		}
		return nil, err
	}
	
	return &userEntity, nil
}

// Create inserts a new user entity
func (r *MongoUserRepository) Create(userEntity *entity.UserEntity) (*entity.UserEntity, error) {
	// Find the maximum user ID and increment
	var maxUserEntity entity.UserEntity
	opts := options.FindOne().SetSort(bson.D{{Key: "user_id", Value: -1}})
	err := r.collection.FindOne(context.Background(), bson.M{}, opts).Decode(&maxUserEntity)
	
	newUserID := 1
	if err == nil {
		// If we found a user, increment the ID
		newUserID = maxUserEntity.UserID + 1
	} else if err != mongo.ErrNoDocuments {
		// If there was an error other than "not found"
		return nil, err
	}
	
	// Set the new User ID, created_at and updated_at
	userEntity.UserID = newUserID
	userEntity.CreatedAt = time.Now()
	userEntity.UpdatedAt = time.Now()
	
	// Ensure we have a new ID
	userEntity.ID = primitive.NewObjectID()
	
	res, err := r.collection.InsertOne(context.Background(), userEntity)
	if err != nil {
		return nil, err
	}
	
	// Set the newly generated ID
	userEntity.ID = res.InsertedID.(primitive.ObjectID)
	
	return userEntity, nil
}

// Update updates an existing user entity
func (r *MongoUserRepository) Update(userEntity *entity.UserEntity) (*entity.UserEntity, error) {
	// Find the existing user
	existingEntity, err := r.FindByID(userEntity.UserID)
	if err != nil {
		return nil, err
	}
	if existingEntity == nil {
		return nil, errors.New("user not found")
	}
	
	// Preserve creation date and ID
	userEntity.CreatedAt = existingEntity.CreatedAt
	userEntity.ID = existingEntity.ID
	userEntity.UpdatedAt = time.Now()
	
	filter := bson.M{"user_id": userEntity.UserID}
	update := bson.M{"$set": userEntity}
	
	_, err = r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}
	
	return userEntity, nil
}

// Delete removes a user entity by ID
func (r *MongoUserRepository) Delete(id int) error {
	result, err := r.collection.DeleteOne(context.Background(), bson.M{"user_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}
