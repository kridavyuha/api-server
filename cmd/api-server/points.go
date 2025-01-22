package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kridavyuha/api-server/internals/trade"

	"github.com/gorilla/websocket"
)

type BallByBall struct {
	MatchID     string         `json:"matchId"`
	Player      map[string]int `json:"players"`
	IsCompleted bool           `json:"isCompleted"`
}

// We get the points from the generator and update the points in the DB
// We also send the points to the clients connected to the websocket
// We need to update the time sereies data for the player in redis cache.
func (app *App) BallPicker(data []byte) {

	fmt.Println(app.WS)
	var ballDetails BallByBall
	if !json.Valid(data) {
		fmt.Println("Invalid JSON data")
		return
	}
	err := json.Unmarshal(data, &ballDetails)
	if err != nil {
		fmt.Println("Error unmarshalling data:", err)
		return
	}

	// Check Match status:
	if ballDetails.IsCompleted {
		// Match is completed, close the websocket connections
		//TODO: render frontend accordingly once the match is completed
		app.ClientsM.Lock()
		for conn := range app.WS {
			err := conn.WriteMessage(websocket.TextMessage, []byte("Match completed"))
			if err != nil {
				fmt.Println("Error writing to client:", err)
			}
			conn.Close()
			// remove the client from the list
		}
		app.ClientsM.Unlock()
	}

	app.ClientsM.Lock()
	for conn, val := range app.WS {
		//TODO: can we implement this through go routines ?
		// check the match_id with that of the client

		if val.MatchID == ballDetails.MatchID {
			err := conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				conn.Close()
			}
		}
	}
	app.ClientsM.Unlock()

	// What if we have multiple leagues and we need to update points for all of them?
	// here we will be having seperate points table for each league
	// so we need to update points for each league seperately

	// get all the leagues_id for this respective match_id
	fmt.Println("MatchID:", ballDetails.MatchID)
	var leagues []string
	app.DB.Raw("SELECT league_id FROM leagues WHERE match_id = ?", ballDetails.MatchID).Scan(&leagues)

	// Write to DB here
	for _, league := range leagues {
		tableName := "players_" + league
		//TODO: Update the points for each player in the league in seperate	go routine
		for playerID, points := range ballDetails.Player {
			fmt.Println("PlayerID:", playerID, "Points:", points, "Table name: ", tableName)
			tx := app.DB.Exec("UPDATE "+tableName+" SET cur_price = cur_price + ? WHERE player_id = ?", points, playerID)
			if tx.Error != nil {
				// Log the error instead of writing to the response writer
				// as we are in a goroutine and cannot write to the response writer
				fmt.Println("Error updating points for league:", league, "player:", playerID, "error:", tx.Error)
			}
		}
	}

	for _, league := range leagues {
		for playerID, points := range ballDetails.Player {
			key := "players_" + league + "_" + playerID
			// This way if we miss any points in entry to redis cache , we may get wrong points while cal from prev points.
			// DB points will be correct but redis cache points will be wrong.
			// TODO: Can we have a cron job to update the redis cache points from DB points? Seems like this whole job of running it ball by ball is heavy.
			lastEntry, err := app.KVStore.LIndex(key, -1)
			if err != nil {
				fmt.Println("Error fetching last entry from redis cache for player:", playerID, "error:", err)
			}

			var lastPoints int
			if lastEntry != "" {
				var timestamp int
				_, err = fmt.Sscanf(lastEntry, "%d,%d", &timestamp, &lastPoints)
				if err != nil {
					fmt.Println("Error parsing last entry from redis cache for player:", playerID, "error:", err)
					continue
				}
			}

			points += lastPoints
			timestamp := time.Now().Unix()
			value := fmt.Sprintf("%d,%d", timestamp, points)
			err = app.KVStore.RPush(key, value)
			if err != nil {
				fmt.Println("Error writing to redis cache for player:", playerID, "error:", err)
			}
		}
	}

	// Write to redis cache

	// w.Write([]byte("Points received"))
}

// Get points for a player from redis cache for a league
// This is generally used to show the graph representation of points for a player in a league
func (app *App) GetPointsPlayerWise(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	playerID := r.URL.Query().Get("player_id")
	if leagueID == "" || playerID == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "required params missing"})

	}

	points, err := trade.New(app.KVStore, app.DB).GetTimeseriesPlayerPoints(playerID, leagueID)

	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"data": points}})
}
