package main

import (
	"net/http"

	"github.com/kridavyuha/api-server/internals/portfolio"
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

func (app *App) GetActivePortfolios(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value("user_id").(int)

	portfolios, err := portfolio.New(app.KVStore, app.DB).GetActivePortfolios(userId)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: portfolios})
}
