package httpserver

import "errors"

var (
	ErrLoginAndPasswordRequired = errors.New("Login and password are required")
	ErrInvalidUserID            = errors.New("Invalid user ID")
	ErrInternalServerError      = errors.New("Internal server error")
	ErrInvalidJsonFormat        = errors.New("Invalid JSON format")
	ErrUserIsNotAuthenticated   = errors.New("User is not authenticated")
	ErrOrderNumberRequired      = errors.New("Order number is required")
	ErrDuplicateOrder           = errors.New("The number has already been downloaded by this user")
	ErrOtherUserOrder           = errors.New("Number uploaded by another user")
	ErrInvalidOrderNumber       = errors.New("Invalid order number")
	ErrLackOfFunds              = errors.New("Lack of funds")
	ErrInvalidNumberFormat      = errors.New("Invalid number format")
)
