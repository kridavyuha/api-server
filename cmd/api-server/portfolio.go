package main

import (
	"backend/internals/portfolio"
	"net/http"
)

func (app *App) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(int)
	leagueId := r.URL.Query().Get("league_id")

	if leagueId == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
		return
	}

	portfolio, err := portfolio.New(app.KVStore, app.DB).GetDetailedPortfolio(userId, leagueId)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: portfolio})
}
