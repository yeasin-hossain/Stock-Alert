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
	r.HandleFunc("/users/{id:[a-fA-F0-9]{24}}", userHandler.GetUser).Methods("GET")
	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	r.HandleFunc("/users/{id:[a-fA-F0-9]{24}}", userHandler.UpdateUser).Methods("PUT")
	r.HandleFunc("/users/{id:[a-fA-F0-9]{24}}", userHandler.DeleteUser).Methods("DELETE")

	// Alert routes
	alertCollection := db.GetCollection("alerts")
	alertRepository := repository.NewMongoAlertRepository(alertCollection)
	alertService := service.NewAlertService(alertRepository)
	alertHandler := handler.NewAlertHandler(alertService)

	r.HandleFunc("/alerts", alertHandler.CreateAlert).Methods("POST")
	r.HandleFunc("/alerts/{id}", alertHandler.GetAlert).Methods("GET")
	r.HandleFunc("/alerts/user/{userId}", alertHandler.GetAlertsByUser).Methods("GET")
	r.HandleFunc("/alerts/{id}", alertHandler.UpdateAlert).Methods("PUT")
	r.HandleFunc("/alerts/{id}", alertHandler.DeleteAlert).Methods("DELETE")

	return r
}
