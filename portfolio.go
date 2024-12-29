package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Portfolio struct {
	PlayerId   string `json:"player_id"`
	Shares     int    `json:"shares"`
	CurPrice   int    `json:"cur_price"`
	PlayerName string `json:"player_name"`
	TeamName   string `json:"team_name"`
}

type DetailedPortfolio struct {
	Players []Portfolio `json:"players"`
	Balance int         `json:"balance"`
}

func (app *App) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	leagueId := r.URL.Query().Get("league_id")
	userId := r.Context().Value("user_id").(int)

	fmt.Println("League ID:", leagueId, "User ID:", userId)

	// Get the portfolio of the user from the DB
	var portfolio []Portfolio
	tx := app.DB.Raw("SELECT player_id, shares FROM portfolio WHERE user_id = ? AND league_id = ?", userId, leagueId).Scan(&portfolio)
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Using LeagueID and PlayerID, get the player name and current price
	for i, player := range portfolio {
		var curPrice int
		tx = app.DB.Raw("SELECT cur_price FROM players_"+leagueId+" WHERE player_id = ?", player.PlayerId).Scan(&curPrice)
		if tx.Error != nil {
			http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
			return
		}
		var playerInfo struct {
			PlayerName string
			Team       string
		}
		app.DB.Raw("SELECT player_name,team FROM players WHERE player_id = ?", player.PlayerId).Scan(&playerInfo)
		portfolio[i].CurPrice = curPrice
		portfolio[i].PlayerName = playerInfo.PlayerName
		portfolio[i].TeamName = playerInfo.Team
	}

	// Get the remaining purse balance
	var balance int
	tx = app.DB.Raw("SELECT remaining_purse FROM purse WHERE user_id = ? AND league_id = ?", userId, leagueId).Scan(&balance)
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Prepare the response
	response := DetailedPortfolio{
		Players: portfolio,
		Balance: balance,
	}

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
