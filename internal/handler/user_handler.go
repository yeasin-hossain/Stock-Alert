package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"github.com/hello-api/internal/common"
	"github.com/hello-api/internal/domain"
	"github.com/hello-api/internal/handler/dto"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		common.RespondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to fetch users")
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, users)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID format")
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		common.HandleError(w, err)
		return
	}
	
	if user == nil {
		common.RespondWithError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, user)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request dto.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}

	// Validate request (basic validation)
	if request.Name == "" || request.Email == "" {
		validationErr := fmt.Errorf("%w: name and email are required", domain.ErrValidation)
		common.HandleError(w, validationErr)
		return
	}

	createdUser, err := h.userService.CreateUser(request)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.RespondWithSuccess(w, http.StatusCreated, createdUser)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID format")
		return
	}

	var request dto.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request format")
		return
	}
	
	// Check if at least one field is provided
	if request.Name == "" && request.Email == "" {
		validationErr := fmt.Errorf("%w: at least one field (name or email) must be provided", domain.ErrValidation)
		common.HandleError(w, validationErr)
		return
	}
	
	updatedUser, err := h.userService.UpdateUser(id, request)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	common.RespondWithSuccess(w, http.StatusOK, updatedUser)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.RespondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid user ID format")
		return
	}

	err = h.userService.DeleteUser(id)
	if err != nil {
		common.HandleError(w, err)
		return
	}

	// For DELETE operations, use NoContent with an empty success response
	common.RespondWithSuccess(w, http.StatusNoContent, nil)
}
