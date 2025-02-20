package notification

import "time"

type Notification struct {
	Id          int       `json:"id"`
	Entity      string    `json:"entity"`
	Actor       int       `json:"actor"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type Transaction struct {
	PlayerName      string    `json:"player_name"`
	Shares          int       `json:"shares"`
	Price           float64   `json:"price"`
	TransactionTime time.Time `json:"transaction_time"`
}
