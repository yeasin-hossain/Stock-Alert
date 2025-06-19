package db

import (
	"github.com/hello-api/pkg/mongo"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"sync"
)

var (
	client     *mongodriver.Client
	clientOnce sync.Once
)

// GetClient returns a singleton MongoDB client
func GetClient() *mongodriver.Client {
	clientOnce.Do(func() {
		client = mongo.ConnectMongo()
	})
	return client
}

// GetDatabase returns the users database
func GetDatabase() *mongodriver.Database {
	return mongo.CreateDatabase(GetClient())
}

// GetCollection returns a specific collection from the users database
func GetCollection(name string) *mongodriver.Collection {
	return GetDatabase().Collection(name)
}
