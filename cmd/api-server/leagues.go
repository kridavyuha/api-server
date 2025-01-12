package main

import (
	"backend/internals/leagues"
	"net/http"
)

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

func (app *App) DeleteLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
		return
	}

	err := leagues.New(app.KVStore, app.DB).DeleteLeague(leagueID)

	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "League deleted successfully"}})
}

func (app *App) StartLeague(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
		return
	}

	// Update the league status to 'started'
	err := leagues.New(app.KVStore, app.DB).StartLeague(leagueID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "League started successfully"}})
}

func (app *App) StartMatch(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
		return
	}

	// Update the match status to 'started'
	err := leagues.New(app.KVStore, app.DB).StartMatch(leagueID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "Match started successfully"}})
}
