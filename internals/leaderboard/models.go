package leaderboard

type score struct {
	UserId   int     `json:"user_id"`
	UserName string  `json:"user_name"`
	Points   float64 `json:"points"`
}
