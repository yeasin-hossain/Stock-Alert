package common

import (
	"encoding/json"
	"net/http"
)

// Response represents a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorData  `json:"error,omitempty"`
}

// ErrorData represents error information in the API response
type ErrorData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewSuccessResponse creates a new success response with data
func NewSuccessResponse(data interface{}) Response {
	return Response{
		Success: true,
		Data:    data,
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(code string, message string) Response {
	return Response{
		Success: false,
		Error: &ErrorData{
			Code:    code,
			Message: message,
		},
	}
}

// RespondWithSuccess sends a success response with standard format
func RespondWithSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	response := NewSuccessResponse(data)
	RespondWithJSON(w, statusCode, response)
}

// RespondWithError sends an error response with standard format
func RespondWithError(w http.ResponseWriter, statusCode int, code string, message string) {
	response := NewErrorResponse(code, message)
	RespondWithJSON(w, statusCode, response)
}

// RespondWithJSON sends a JSON response with given status code
func RespondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}
