package main

import (
	"backend/internals/leagues"
	"encoding/json"
	"fmt"
	"net/http"
)

type Player struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
}

type GetLeagueDetails struct {
	LeagueID      string             `json:"league_id"`
	MatchID       string             `json:"match_id"`
	PlayerDetails []GetPlayerDetails `json:"player_details"`
}

type GetPlayerDetails struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Team       string `json:"team"`
	ProfilePic string `json:"profile_pic"`
	CurPrice   int    `json:"cur_price"`
	LastChange string `json:"last_change"`
	Shares     int    `json:"shares"`
}

type League struct {
	LeagueID        string `json:"league_id"`
	MatchID         string `json:"match_id"`
	Capacity        int    `json:"capacity"`
	EntryFee        int    `json:"entry_fee"`
	Registered      int    `json:"registered"`
	UsersRegistered string `json:"users_registered"`
	LeagueStatus    string `json:"league_status"`
	TeamA           string `json:"team_a"`
	TeamB           string `json:"team_b"`
	IsRegistered    bool   `json:"is_registered"`
}

func (app *App) CreateLeague(w http.ResponseWriter, r *http.Request) {

	var league leagues.CreateLeagueRequestBody
	err := getBody(r, &league)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "Invalid request body"})
	}

	err = leagues.New(app.KVStore, app.DB).CreateLeague(league)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: err.Error()})
	}

	sendResponse(w, httpResp{Status: http.StatusCreated, Data: map[string]interface{}{"message": "League created successfully"}})
}

func (app *App) GetLeagues(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from the context
	userID := r.Context().Value("user_id").(int)

	leagues, err := leagues.New(app.KVStore, app.DB).GetLeagues(userID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: leagues})
}

func (app *App) RegisterLeague(w http.ResponseWriter, r *http.Request) {

	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
	}
	userID := r.Context().Value("user_id").(int)

	err := leagues.New(app.KVStore, app.DB).RegisterToLeague(userID, leagueID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "Registered successfully"}})
}

// this need to be moved to trade.go
func (app *App) GetLeague(w http.ResponseWriter, r *http.Request) {
	matchID := r.URL.Query().Get("league_id")
	if matchID == "" {
		http.Error(w, "match_id is required", http.StatusBadRequest)
		return
	}

	// Create a table name using the match_id
	tableName := "players_" + matchID

	// Get user_id
	userID := r.Context().Value("user_id").(int)

	// Get all players from the table
	var playerDetails []GetPlayerDetails

	query := `
	SELECT p.player_id, p.player_name, p.team, pl.cur_price, pl.last_change
	FROM players p
	JOIN ` + tableName + ` pl ON p.player_id = pl.player_id;`

	err := app.DB.Raw(query).Scan(&playerDetails).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Player Details: ", playerDetails)

	sharesQuery := `
		SELECT player_id, shares
		FROM portfolio
		WHERE league_id = ? AND user_id = ?;`

	var sharesData []struct {
		PlayerID string `json:"player_id"`
		Shares   int    `json:"shares"`
	}

	err = app.DB.Raw(sharesQuery, matchID, userID).Scan(&sharesData).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sharesMap := make(map[string]int)
	for _, share := range sharesData {
		sharesMap[share.PlayerID] = share.Shares
	}

	for i, player := range playerDetails {
		playerDetails[i].Shares = sharesMap[player.PlayerID]
	}

	// Return the players
	json.NewEncoder(w).Encode(playerDetails)
}

func (app *App) DeleteLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		http.Error(w, "league_id is required", http.StatusBadRequest)
		return
	}

	// Delete the league from the leagues table
	err := app.DB.Exec("DELETE FROM leagues WHERE league_id = ?", leagueID).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete the table_{match_id} from the database
	err = app.DB.Exec("DROP TABLE players_" + leagueID).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := respMssge{
		Message: "League deleted successfully",
	}

	json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusNoContent)
}

func (app *App) StartLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		http.Error(w, "league_id is required", http.StatusBadRequest)
		return
	}

	// Update the league status to 'started'
	err := app.DB.Exec("UPDATE leagues SET league_status = 'opened' WHERE league_id = ?", leagueID).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("League opened successfully"))
	w.WriteHeader(http.StatusOK)
}

func (app *App) StartMatch(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		http.Error(w, "league_id is required", http.StatusBadRequest)
		return
	}
	var matchID string
	err := app.DB.Raw("SELECT match_id FROM leagues WHERE league_id = ?", leagueID).Scan(&matchID).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make a Get Request to endpoint http://localhost:8081/scores?match_id={match_id} to get the scores of the match
	resp, err := http.Get("http://localhost:8081/scores?match_id=" + matchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to trigger webhook", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Match started"))

}
