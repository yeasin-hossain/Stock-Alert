package repository

import (
	"context"
	"errors"
	
	"github.com/hello-api/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(collection *mongo.Collection) *MongoUserRepository {
	return &MongoUserRepository{
		collection: collection,
	}
}

func (r *MongoUserRepository) FindAll() ([]domain.User, error) {
	var users []domain.User
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var user domain.User
		if err := cursor.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *MongoUserRepository) FindByID(id int) (domain.User, error) {
	var user domain.User
	err := r.collection.FindOne(context.Background(), bson.M{"id": id}).Decode(&user)
	return user, err
}

func (r *MongoUserRepository) Create(user domain.User) (domain.User, error) {
	_, err := r.collection.InsertOne(context.Background(), user)
	return user, err
}

func (r *MongoUserRepository) Update(user domain.User) (domain.User, error) {
	result, err := r.collection.UpdateOne(
		context.Background(),
		bson.M{"id": user.ID},
		bson.M{"$set": user},
	)
	if err != nil {
		return domain.User{}, err
	}
	if result.MatchedCount == 0 {
		return domain.User{}, errors.New("user not found")
	}
	return user, nil
}

func (r *MongoUserRepository) Delete(id int) error {
	result, err := r.collection.DeleteOne(context.Background(), bson.M{"id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return errors.New("user not found")
	}
	return nil
}
