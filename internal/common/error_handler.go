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
// Supports custom error messages by wrapping errors with context (use errors.New or fmt.Errorf)
func HandleError(w http.ResponseWriter, err error) {
	var code, message string
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		code = "NOT_FOUND"
		message = getCustomOrDefaultMessage(err, "Resource not found")
		RespondWithError(w, http.StatusNotFound, code, message)
	case errors.Is(err, domain.ErrValidation):
		code = "VALIDATION_ERROR"
		message = getCustomOrDefaultMessage(err, "Validation error")
		RespondWithError(w, http.StatusBadRequest, code, message)
	case errors.Is(err, domain.ErrUserAlreadyExit):
		code = "USER_ALREADY_EXISTS"
		message = getCustomOrDefaultMessage(err, "User already exists")
		RespondWithError(w, http.StatusConflict, code, message)
	case errors.Is(err, domain.ErrUnauthorized):
		code = "UNAUTHORIZED"
		message = getCustomOrDefaultMessage(err, "Unauthorized access")
		RespondWithError(w, http.StatusUnauthorized, code, message)
	case errors.Is(err, domain.ErrForbidden):
		code = "FORBIDDEN"
		message = getCustomOrDefaultMessage(err, "Access forbidden")
		RespondWithError(w, http.StatusForbidden, code, message)
	default:
		// Log the actual error for debugging
		// logger.Error("Unexpected error", "error", err)
		code = "INTERNAL_ERROR"
		message = getCustomOrDefaultMessage(err, "An unexpected error occurred")
		RespondWithError(w, http.StatusInternalServerError, code, message)
	}
}

// Generalized error message mapping for domain errors
var errorMessageMap = map[error]string{
	domain.ErrUserNotFound:    "Resource not found",
	domain.ErrValidation:      "Validation error",
	domain.ErrUserAlreadyExit: "User already exists",
	domain.ErrUnauthorized:    "Unauthorized access",
	domain.ErrForbidden:       "Access forbidden",
	domain.ErrInternal:        "An unexpected error occurred",
}

// getCustomOrDefaultMessage returns the custom error message if it differs from the base error, otherwise returns the default from the map
func getCustomOrDefaultMessage(err error, def string) string {
	if err == nil {
		return def
	}
	msg := err.Error()
	// Try to get a general message from the map
	for base, gen := range errorMessageMap {
		if errors.Is(err, base) {
			if msg != gen && msg != "" {
				return msg // custom message
			}
			return gen // general message
		}
	}
	if msg != def && msg != "" {
		return msg
	}
	return def
}
