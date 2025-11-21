package models

import (
	"time"
)

type Balance struct {
	Current   float64 `json:"current" db:"current"`
	Withdrawn float64 `json:"withdrawn" db:"sum"`
}

type WithdrawBalance struct {
	Order       string    `json:"order" db:"order_number"`
	Sum         float64   `json:"sum" db:"sum"`
	ProcessedAt time.Time `json:"processed_at" db:"processed_at"`
}
