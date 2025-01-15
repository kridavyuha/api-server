package leagues

// League Table structure
type Leagues struct {
	LeagueID        string `json:"league_id" gorm:"primaryKey;not null"`
	MatchID         string `json:"match_id" gorm:"not null"`
	EntryFee        int    `json:"entry_fee" gorm:"not null"`
	Capacity        int    `json:"capacity" gorm:"default:100;not null"`
	Registered      int    `json:"registered" gorm:"default:0;not null"`
	UsersRegistered string `json:"users_registered" gorm:"default:'';not null"`
	LeagueStatus    string `json:"league_status" gorm:"default:'not started';not null"`
}

type League struct {
	LeagueID        string `json:"league_id"`
	MatchID         string `json:"match_id"`
	EntryFee        int    `json:"entry_fee"`
	Capacity        int    `json:"capacity"`
	Registered      int    `json:"registered"`
	UsersRegistered string `json:"users_registered"`
	LeagueStatus    string `json:"league_status"`
	TeamA           string `json:"team_a"`
	TeamB           string `json:"team_b"`
	IsRegistered    bool   `json:"is_registered"`
}

type CreateLeagueRequestBody struct {
	MatchID  string `json:"match_id"`
	Capacity int    `json:"capacity"`
	EntryFee int    `json:"entry_fee"`
}

type Fixture struct {
	MatchID string `json:"match_id"`
	TeamA   string `json:"team_a"`
	TeamB   string `json:"team_b"`
}

type PlayerDetails struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Squad struct {
	Team    string          `json:"team"`
	Players []PlayerDetails `json:"players"`
	Id      int             `json:"id"`
}
