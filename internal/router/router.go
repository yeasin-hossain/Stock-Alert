package router

import (
	"github.com/gorilla/mux"
	"github.com/hello-api/internal/db"
	"github.com/hello-api/internal/domain"
	"github.com/hello-api/internal/handler"
	"github.com/hello-api/internal/repository"
	"github.com/hello-api/internal/service"
)

func InitializeRoutes() *mux.Router {
	r := mux.NewRouter()

	// Initialize dependencies using interfaces for better decoupling
	userCollection := db.GetCollection("users")
	
	// Repository layer
	var userRepository domain.UserRepository
	userRepository = repository.NewMongoUserRepository(userCollection)
	
	// Service layer
	var userService domain.UserService
	userService = service.NewUserService(userRepository)
	
	// Handler layer
	userHandler := handler.NewUserHandler(userService)

	// User routes
	r.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	r.HandleFunc("/users/{id:[0-9]+}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id:[0-9]+}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id:[0-9]+}", userHandler.DeleteUser).Methods("DELETE")

	return r
}
