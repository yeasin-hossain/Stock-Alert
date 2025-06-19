package common

import (
	"errors"
	"net/http"

	"github.com/hello-api/internal/domain"
)

// ErrorHandler is a middleware for handling errors in the API
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Call the next handler
		next.ServeHTTP(w, r)
		// Error handling is done in the handlers themselves
	})
}

// HandleError maps different error types to appropriate HTTP responses
func HandleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		RespondWithError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found")
	case errors.Is(err, domain.ErrValidation):
		RespondWithError(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, domain.ErrUnauthorized):
		RespondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized access")
	case errors.Is(err, domain.ErrForbidden):
		RespondWithError(w, http.StatusForbidden, "FORBIDDEN", "Access forbidden")
	default:
		// Log the actual error for debugging
		// logger.Error("Unexpected error", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
	}
}
