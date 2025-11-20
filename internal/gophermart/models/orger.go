package models

import (
	"errors"
	"time"
)

type Order struct {
	UID        int       `json:"-" db:"uid"`
	UserID     int       `json:"-" db:"user_id"`
	Number     string    `json:"number" db:"number"`
	Status     string    `json:"status" db:"status"`
	Accrual    float64   `json:"accrual,omitempty" db:"accrual"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
}

// статусы заказов
const (
	OrderStatusNew        = "NEW"
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

var (
	ErrDuplicateOrder = errors.New("the number has already been downloaded by this user")
	ErrOtherUserOrder = errors.New("number uploaded by another user")

	ErrInvalidOrderNumber = errors.New("invalid order number")
	ErrLackOfFunds        = errors.New("lack of funds")
)
