package mongo

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() *mongo.Client {
	// Hardcoded MongoDB URI for development
	mongoURI := "mongodb://localhost:27017/dev_db"
	log.Printf("Using hardcoded MongoDB URI: %s", mongoURI)
	
	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(context.Background(), nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")
	return client
}

func CreateDatabase(client *mongo.Client) *mongo.Database {
	return client.Database("users")
}
