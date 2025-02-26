package cache

import "time"

type PlayerInLeague struct {
	PlayerID    string    `json:"player_id"`
	CurPrice    float64   `json:"cur_price"`
	BasePrice   float64   `json:"base_price"`
	LastUpdated time.Time `json:"last_updated"`
	Status      string    `json:"status"`
}

type Portfolio struct {
	PlayerId   string  `json:"player_id"`
	Shares     int     `json:"shares"`
	AvgPrice   float64 `json:"avg_price"`
	CurPrice   float64 `json:"cur_price"`
	PlayerName string  `json:"player_name"`
	TeamName   string  `json:"team_name"`
}

type PlayerMetaData struct {
	PlayerId   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Team       string `json:"team"`
}
