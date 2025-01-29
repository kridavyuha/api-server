package trade

type TransactionDetails struct {
	Shares             int     `json:"shares"`
	PlayerCurrentPrice float64 `json:"price"`
}

type GetPlayerDetails struct {
	PlayerID   string  `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Team       string  `json:"team"`
	ProfilePic string  `json:"profile_pic"`
	BasePrice  float64 `json:"base_price"`
	CurPrice   float64 `json:"cur_price"`
	LastChange string  `json:"last_change"`
	Shares     int     `json:"shares"`
}
