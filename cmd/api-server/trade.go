package main

import (
	"fmt"
	"net/http"

	"github.com/kridavyuha/api-server/internals/trade"

	"github.com/gorilla/websocket"
)

//TODO: Do need to check whether the user registered for the league before buying the player?
// Eve though he can not buy directly from the UI without registering !

func (app *App) TransactPlayers(w http.ResponseWriter, r *http.Request) {
	// Get transaction_type from the query params
	transactionType := r.URL.Query().Get("transaction_type")
	leagueId := r.URL.Query().Get("league_id")
	playerId := r.URL.Query().Get("player_id")
	userId := r.Context().Value("user_id").(int)

	if playerId == "" || transactionType == "" || leagueId == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "player_id, league_id and transaction_type are required"})
	}

	var transactionDetails trade.TransactionDetails
	err := getBody(r, &transactionDetails)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "Invalid request body"})
		return
	}

	err = trade.New(app.KVStore, app.DB, app.MQConn).Transaction(transactionType, playerId, leagueId, userId, transactionDetails)

	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "Transaction successful"}})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WSDetails struct {
	MatchID  string
	LeagueID string
}

// TODO: Assuming we are only having 1 match pool for now
// We can also avoid lock i guess
func (app *App) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Websocket connection established")
	matchID := r.URL.Query().Get("match_id")
	if matchID == "" {
		http.Error(w, "match_id is required", http.StatusBadRequest)
		return
	}
	leagueID := r.URL.Query().Get("league_id")
	if leagueID == "" {
		http.Error(w, "league_id is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	// defer the connection close and remove the client from the list
	defer func() {
		conn.Close()
		app.ClientsM.Lock()
		delete(app.WS, conn)
		app.ClientsM.Unlock()
	}()

	wsDetails := WSDetails{
		MatchID:  matchID,
		LeagueID: leagueID,
	}

	app.WS[conn] = wsDetails

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (app *App) Trade(w http.ResponseWriter, r *http.Request) {

	userID := r.Context().Value("user_id").(int)

	league_id := r.URL.Query().Get("league_id")
	if league_id == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "league_id is required"})
	}

	playerDetails, err := trade.New(app.KVStore, app.DB, app.MQConn).GetPlayerDetails(league_id, userID)

	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: playerDetails})
}
