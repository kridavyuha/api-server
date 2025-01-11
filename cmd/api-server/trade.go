package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type TransactionDetails struct {
	Shares int `json:"shares"`
	Price  int `json:"price"`
}

//TODO: Do we need to check whether the user registered for the league before buying the player?
// Eve though he can not buy directly from the UI without registering !

func (app *App) TransactPlayers(w http.ResponseWriter, r *http.Request) {
	fmt.Println("TransactPlayers")
	// Get transaction_type from the query params
	transactionType := r.URL.Query().Get("transaction_type")

	playerId := r.URL.Query().Get("player_id")
	if playerId == "" {
		http.Error(w, "player_id is required", http.StatusBadRequest)
		return
	}
	leagueId := r.URL.Query().Get("league_id")
	if leagueId == "" {
		http.Error(w, "league_id is required", http.StatusBadRequest)
		return
	}
	userId := r.Context().Value("user_id").(int)

	fmt.Println("Transaction Type:", transactionType, "Player ID:", playerId, "League ID:", leagueId, "User ID:", userId)

	// extract the body of the request
	// Don't believe the price from that is coming from client side
	// Always calculate the price on the server side
	// from players_{league_id} table get the price of the player_id
	var price int
	tx := app.DB.Raw("SELECT cur_price FROM players_"+leagueId+" WHERE player_id = ?", playerId).Scan(&price)
	if tx.Error != nil {
		http.Error(w, tx.Error.Error(), http.StatusInternalServerError)
		return
	}

	var transactionDetails TransactionDetails
	err := json.NewDecoder(r.Body).Decode(&transactionDetails)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	transactionDetails.Price = price

	fmt.Println("Transaction Details:", transactionDetails)

	var balance int
	err = app.DB.Raw("SELECT remaining_purse FROM purse WHERE user_id = ? AND league_id = ?", userId, leagueId).Scan(&balance).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if transactionType == "buy" {
		// Check the user's balance from the purse table
		// Calculate the total cost
		totalCost := transactionDetails.Shares * transactionDetails.Price

		// Check if the user has enough balance
		if balance < totalCost {
			http.Error(w, "insufficient balance", http.StatusBadRequest)
			return
		}
		balance = balance - totalCost
	} else if transactionType == "sell" {
		// Here check whether the user has enough shares to sell
		var shares int
		err = app.DB.Raw("SELECT shares FROM portfolio WHERE user_id = ? AND player_id = ? AND league_id = ?", userId, playerId, leagueId).Scan(&shares).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if shares < transactionDetails.Shares {
			http.Error(w, "insufficient shares", http.StatusBadRequest)
			return
		}
		balance = balance + transactionDetails.Shares*transactionDetails.Price
	}

	// Create an entry in transactions table
	err = app.DB.Exec("INSERT INTO transactions (user_id, player_id, league_id, shares, price, transaction_type, transaction_time) VALUES (?, ?, ?, ?, ?, ?, ?)", userId, playerId, leagueId, transactionDetails.Shares, transactionDetails.Price, transactionType, time.Now()).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the user's balance
	err = app.DB.Exec("UPDATE purse SET remaining_purse = ? WHERE user_id = ? and league_id = ?", balance, userId, leagueId).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the portfolio table
	// If the player is already present in the portfolio, update the shares and average price
	// If the player is not present, insert a new row
	if transactionType == "buy" {
		// Check if the player is already in the portfolio
		var result struct {
			Count    int
			Invested int
		}
		err = app.DB.Raw("SELECT COUNT(*) as count, invested FROM portfolio WHERE user_id = ? AND player_id = ? AND league_id = ? GROUP BY invested", userId, playerId, leagueId).Scan(&result).Error

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if result.Count == 0 {
			// Insert new row if player is not in the portfolio
			err = app.DB.Exec("INSERT INTO portfolio (user_id, player_id, league_id, shares, invested) VALUES (?, ?, ?, ?, ?)", userId, playerId, leagueId, transactionDetails.Shares, transactionDetails.Shares*transactionDetails.Price).Error
		} else {
			// Update the shares if player is already in the portfolio
			err = app.DB.Exec("UPDATE portfolio SET shares = shares + ?, invested = invested + ? WHERE user_id = ? AND player_id = ? AND league_id = ?", transactionDetails.Shares, transactionDetails.Price, userId, playerId, leagueId).Error
		}
	} else if transactionType == "sell" {
		// Update or delete the shares if player is already in the portfolio
		var remainingShares int
		err = app.DB.Raw("SELECT shares FROM portfolio WHERE user_id = ? AND player_id = ? AND league_id = ?", userId, playerId, leagueId).Scan(&remainingShares).Error
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		remainingShares -= transactionDetails.Shares
		if remainingShares == 0 {
			err = app.DB.Exec("DELETE FROM portfolio WHERE user_id = ? AND player_id = ? AND league_id = ?", userId, playerId, leagueId).Error
		} else {
			err = app.DB.Exec("UPDATE portfolio SET shares = ?, invested = invested - ? WHERE user_id = ? AND player_id = ? AND league_id = ?", remainingShares, transactionDetails.Price*transactionDetails.Shares, userId, playerId, leagueId).Error
		}
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type respMssge struct {
		Message string `json:"message"`
	}

	response := respMssge{Message: "Transaction successful"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Add this purchase to redis queue : so there is a record of all the transactions
	// There should be a process monitoring this queue and updating the player's price in the player table
	// As we thought theprice of a player stocks will be updated on every transaction No need of maintianing the counter then.

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