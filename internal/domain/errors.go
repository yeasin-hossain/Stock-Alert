package domain

import "errors"

var (
	// ErrUserNotFound is returned when a user is not found
	ErrUserNotFound = errors.New("user not found")
	
	// if user already exists
	ErrUserAlreadyExit = errors.New("user Already exit")
	
	// ErrValidation is returned when input validation fails
	ErrValidation = errors.New("validation error")
	
	// ErrUnauthorized is returned when a request lacks valid authentication
	ErrUnauthorized = errors.New("unauthorized")
	
	// ErrForbidden is returned when a request is not allowed
	ErrForbidden = errors.New("forbidden")
	
	// ErrInternal is returned when an unexpected internal error occurs
	ErrInternal = errors.New("internal server error")
)
