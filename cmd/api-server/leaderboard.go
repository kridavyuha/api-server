package main

import (
	"net/http"

	"github.com/kridavyuha/api-server/internals/leaderboard"
)

func (app *App) GetLeaderboard(w http.ResponseWriter, r *http.Request) {

	leagueId := r.URL.Query().Get("league_id")

	if leagueId == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
		return
	}

	scores, err := leaderboard.New(app.KVStore, app.DB).GetLeaderboard(leagueId)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: scores})
}
