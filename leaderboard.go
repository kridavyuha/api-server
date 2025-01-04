package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type score struct {
	UserId   int    `json:"user_id`
	UserName string `json:"user_name"`
	Points   int    `json:"points"`
}

func (app *App) GetLeaderboard(w http.ResponseWriter, r *http.Request) {

	leagueId := r.URL.Query().Get("league_id")
	fmt.Println("League ID: ", leagueId)

	// First step -> Get all the registered users of this league from the leagues table using the league_id
	var participatingUsers string
	scores := make([]score, 0)
	tx := app.DB.Raw("SELECT users_registered FROM leagues WHERE league_id = ?", leagueId).Scan(&participatingUsers)
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
	}
	// Second step -> Go through each user, calculate their points by 1) fetching the remaining purse from the `purse` table 2) getting the portfolio of the user & calculate the sum of points.
	userIds := strings.Split(participatingUsers, ",")
	for _, userId := range userIds {
		score := &score{}
		userid, _ := strconv.Atoi(userId)
		resp, err := getUserPortfolio(app, leagueId, userid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		score.UserId = userid
		score.Points = resp.Balance
		// Now go through each player from this user's portfolio and get their worth
		for _, player := range resp.Players {
			score.Points += player.CurPrice * player.Shares
		}
		app.DB.Raw("SELECT user_name FROM users WHERE user_id = ?", userid).Scan(&score.UserName)
		scores = append(scores, *score)
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Points > scores[j].Points
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}
