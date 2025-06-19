package repository

import (
	"context"
	"time"

	"github.com/hello-api/internal/handler/dto"
	"github.com/hello-api/internal/repository/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoAlertRepository struct {
	collection *mongo.Collection
}

func NewMongoAlertRepository(collection *mongo.Collection) *MongoAlertRepository {
	return &MongoAlertRepository{collection: collection}
}

func (r *MongoAlertRepository) Create(alertReq *dto.AlertCreateRequest) (*dto.AlertResponse, error) {
	alertEntity := entity.AlertEntity{
		ID:        primitive.NewObjectID().Hex(),
		Name:      alertReq.Name,
		Price:     alertReq.Price,
		Rule:      entity.AlertRule(alertReq.Rule),
		StopDate:  alertReq.StopDate,
		StartDate: alertReq.StartDate,
		Status:    entity.AlertStatus(alertReq.Status),
		UserID:    alertReq.UserID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := r.collection.InsertOne(context.Background(), alertEntity)
	if err != nil {
		return nil, err
	}
	return mapAlertEntityToDTO(&alertEntity), nil
}

func (r *MongoAlertRepository) FindByID(id string) (*dto.AlertResponse, error) {
	var alert entity.AlertEntity
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&alert)
	if err != nil {
		return nil, err
	}
	return mapAlertEntityToDTO(&alert), nil
}

func (r *MongoAlertRepository) FindAllByUser(userId string) ([]dto.AlertResponse, error) {
	var alerts []entity.AlertEntity
	cursor, err := r.collection.Find(context.Background(), bson.M{"userId": userId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())
	if err := cursor.All(context.Background(), &alerts); err != nil {
		return nil, err
	}
	var result []dto.AlertResponse
	for _, alert := range alerts {
		result = append(result, *mapAlertEntityToDTO(&alert))
	}
	return result, nil
}

func (r *MongoAlertRepository) Update(id string, alertReq *dto.AlertCreateRequest) (*dto.AlertResponse, error) {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"name":       alertReq.Name,
		"price":      alertReq.Price,
		"rule":       alertReq.Rule,
		"stopDate":   alertReq.StopDate,
		"startDate":  alertReq.StartDate,
		"status":     alertReq.Status,
		"userId":     alertReq.UserID,
		"updated_at": time.Now(),
	}}
	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *MongoAlertRepository) Delete(id string) error {
	_, err := r.collection.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

func mapAlertEntityToDTO(alert *entity.AlertEntity) *dto.AlertResponse {
	return &dto.AlertResponse{
		ID:        alert.ID,
		Name:      alert.Name,
		Price:     alert.Price,
		Rule:      dto.AlertRule(alert.Rule),
		StopDate:  alert.StopDate,
		StartDate: alert.StartDate,
		Status:    dto.AlertStatus(alert.Status),
		UserID:    alert.UserID,
		CreatedAt: alert.CreatedAt,
		UpdatedAt: alert.UpdatedAt,
	}
}
