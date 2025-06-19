package router

import (
	"github.com/gorilla/mux"
	"github.com/hello-api/internal/handler"
	"github.com/hello-api/internal/repository"
	"github.com/hello-api/internal/service"
	"github.com/hello-api/internal/db"
)

func InitializeRoutes() *mux.Router {
	r := mux.NewRouter()

	// Initialize dependencies
	userCollection := db.GetCollection("users")
	userRepository := repository.NewMongoUserRepository(userCollection)
	userService := service.NewUserService(userRepository)
	userHandler := handler.NewUserHandler(userService)

	// User routes
	r.HandleFunc("/users", userHandler.GetUsers).Methods("GET")
	r.HandleFunc("/users/{id:[0-9]+}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id:[0-9]+}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id:[0-9]+}", userHandler.DeleteUser).Methods("DELETE")

	return r
}
