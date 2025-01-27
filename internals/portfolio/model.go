package portfolio

type Portfolio struct {
	PlayerId   string  `json:"player_id"`
	Shares     int     `json:"shares"`
	AvgPrice   float64 `json:"avg_price"`
	CurPrice   float64 `json:"cur_price"`
	PlayerName string  `json:"player_name"`
	TeamName   string  `json:"team_name"`
}

type DetailedPortfolio struct {
	Players []Portfolio `json:"players"`
	Balance float64     `json:"balance"`
}
