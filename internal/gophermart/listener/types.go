package listener

import "time"

type Job struct {
	OrderID   int       `json:"order_id"`
	UserID    int       `json:"user_id"`
	Number    string    `json:"number"`
	Status    string    `json:"status"`
	Attempt   int       `json:"attempt"`
	CreatedAt time.Time `json:"created_at"`
}

type AccrualResponse struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}
