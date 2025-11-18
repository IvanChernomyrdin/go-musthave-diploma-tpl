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
	ErrDuplicateOrder = errors.New("номер уже был загружен этим пользователем")
	ErrOtherUserOrder = errors.New("номер загружен другим пользователем")
)
