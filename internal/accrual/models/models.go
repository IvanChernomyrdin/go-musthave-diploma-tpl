package models

type ProductReward struct {
	Match      string  `json:"match"`
	Reward     float64 `json:"reward"`
	RewardType string  `json:"reward_type"`
}

type Order struct {
	Order string  `json:"order"`
	Goods []Goods `json:"goods"`
}

type Goods struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

const (
	Registered  = "REGISTERED"
	Invalid     = "INVALID"
	Progressing = "PROCESSING"
	Processed   = "PROCESSED"
)

type AccrualInfo struct {
	Order   int64   `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

type ParseMatch struct {
	Order int64   `json:"order"`
	Price float64 `json:"price"`
}
