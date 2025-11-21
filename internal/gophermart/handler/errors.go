package httpserver

import "errors"

var (
	ErrLoginAndPasswordRequired = errors.New("login and password are required")
	ErrInvalidUserID            = errors.New("invalid user ID")
	ErrInternalServerError      = errors.New("internal server error")
	ErrInvalidJsonFormat        = errors.New("invalid JSON format")
	ErrUserIsNotAuthenticated   = errors.New("user is not authenticated")
	ErrOrderNumberRequired      = errors.New("order number is required")
	ErrDuplicateOrder           = errors.New("the number has already been downloaded by this user")
	ErrOtherUserOrder           = errors.New("number uploaded by another user")
	ErrInvalidOrderNumber       = errors.New("invalid order number")
	ErrLackOfFunds              = errors.New("lack of funds")
	ErrInvalidNumberFormat      = errors.New("invalid number format")
	ErrInvalidLoginOrPassword   = errors.New("invalid login or password")
	ErrInvalidRequestFormat     = errors.New("invalid request format")
	ErrLoginAlreadyExists       = errors.New("login already exists")
)
