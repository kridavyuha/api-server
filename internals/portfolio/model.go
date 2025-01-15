package portfolio

type Portfolio struct {
	PlayerId   string `json:"player_id"`
	Shares     int    `json:"shares"`
	Invested   int    `json:"invested"`
	CurPrice   int    `json:"cur_price"`
	PlayerName string `json:"player_name"`
	TeamName   string `json:"team_name"`
}

type DetailedPortfolio struct {
	Players []Portfolio `json:"players"`
	Balance int         `json:"balance"`
}
