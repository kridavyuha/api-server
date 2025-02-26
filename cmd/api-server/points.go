package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/kridavyuha/api-server/internals/leagues"
	"github.com/kridavyuha/api-server/internals/trade"

	"github.com/gorilla/websocket"
)

// data that is being sent to websocket clients should be distunguished for
// perffactor and coreprocessor

type PerfDetails struct {
	PerfFactor  map[string]int `json:"perf_factor"`
	MatchID     string         `json:"match_id"`
	IsCompleted bool           `json:"isCompleted"`
}

type CoreDetails struct {
	//TODO: This might be chaned to map[string]float64
	CoreFactor map[string]string `json:"core_factor"`
	LeagueId   string            `json:"league_id"`
}

type PriceUpdate struct {
	IsPerf      bool        `json:"is_perf"`
	PerfDetails PerfDetails `json:"perf_details"`
	IsCore      bool        `json:"is_core"`
	CoreDetails CoreDetails `json:"core_details"`
}

// We get the points from the generator and update the points in the DB
// We also send the points to the clients connected to the websocket
// We need to update the time sereies data for the player in redis cache.
func (app *App) BallPicker(data []byte) {

	var PerfDetails PerfDetails
	if !json.Valid(data) {
		fmt.Println("Invalid JSON data")
		return
	}
	err := json.Unmarshal(data, &PerfDetails)
	if err != nil {
		fmt.Println("Error unmarshalling data:", err)
		return
	}

	// Check Match status:
	if PerfDetails.IsCompleted {
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

		// create a priceupdate struct
		priceUpdate := PriceUpdate{
			IsPerf:      true,
			PerfDetails: PerfDetails,
			IsCore:      false,
		}

		data, err := json.Marshal(priceUpdate)
		if err != nil {
			fmt.Println("Error marshalling data:", err)
			return
		}

		if val.MatchID == PerfDetails.MatchID {
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
	var leagues []string
	app.DB.Raw("SELECT league_id FROM leagues WHERE match_id = ?", PerfDetails.MatchID).Scan(&leagues)

	// Write to DB here
	for _, league := range leagues {
		tableName := "players_" + league
		//TODO: Update the points for each player in the league in seperate	go routine
		for playerID, points := range PerfDetails.PerfFactor {
			key := "players_" + league + "_" + playerID
			err := app.DB.Exec("UPDATE "+tableName+" SET cur_price = cur_price + ?, updated_at = ? WHERE player_id = ?", float64(points), time.Now(), playerID).Error
			if err != nil {
				fmt.Println("Error updating points for player:", playerID, "error:", err)
			}
			lastEntry, err := app.KVStore.LIndex(key, -1)
			if err != nil {
				fmt.Println("Error fetching last entry from redis cache for player:", playerID, "error:", err)
			}

			var lastPoints float64
			if lastEntry != "" {
				val := strings.Split(lastEntry, ",")
				lastPoints, err = strconv.ParseFloat(val[1], 64)
				if err != nil {
					fmt.Println("Error parsing points from redis cache for player:", playerID, "error:", err)
				}
			} else {
				//TODO: Load the base price from DB
			}

			newPoints := float64(points) + lastPoints

			now := time.Now()

			timestamp := now.Format("2006-01-02 15:04:05.000000-07")

			value := fmt.Sprintf("%s,%.2f", timestamp, newPoints)
			err = app.KVStore.RPush(key, value)
			if err != nil {
				fmt.Println("Error writing to redis cache for player:", playerID, "error:", err)
			}
			// also update in players_<league_id>
			app.KVStore.HSet("players_"+league, playerID, newPoints)
		}
	}
}

func (app *App) HandleUpdateCorePrices(Leagues []leagues.League) {
	// loop over the websocket clients and send the updated prices
	for _, league := range Leagues {
		go func(league leagues.League) {
			// get all the players for this league

			_, err := app.KVStore.HGet("players_"+league.LeagueID, "cache_safe")
			if err != nil {
				switch {
				case err == redis.Nil:
					// get the data from the DB
					var players []struct {
						PlayerID string  `json:"player_id"`
						CurPrice float64 `json:"cur_price"`
					}
					err := app.DB.Raw("SELECT player_id, cur_price FROM players_" + league.LeagueID).Scan(&players).Error
					if err != nil {
						fmt.Println("Error fetching players for league:", league.LeagueID, "error:", err)
						return
					}

					for _, player := range players {
						// put this data in the redis cache
						err := app.KVStore.HSet("players_"+league.LeagueID, player.PlayerID, player.CurPrice)
						if err != nil {
							fmt.Println("Error writing to redis cache for league:", league.LeagueID, "error:", err)
							return
						}
					}
					app.KVStore.HSet("players_"+league.LeagueID, "cache_safe", true)

				case err != nil:
					fmt.Println("Error fetching cache_safe for league:", league.LeagueID, "error:", err)
					return
				}
			}

			// get the data from the redis cache
			players, err := app.KVStore.HGetAll("players_" + league.LeagueID)
			if err != nil {
				fmt.Println("Error fetching players from redis cache for league:", league.LeagueID, "error:", err)
				return
			}

			// fmt.Println("Players:", players)
			// Create Core details struct
			coreDetails := CoreDetails{
				CoreFactor: players,
				LeagueId:   league.LeagueID,
			}

			priceUpdate := PriceUpdate{
				IsPerf:      false,
				IsCore:      true,
				CoreDetails: coreDetails,
			}

			data, err := json.Marshal(priceUpdate)
			if err != nil {
				fmt.Println("Error marshalling data:", err)
				return
			}

			app.ClientsM.Lock()
			for conn, val := range app.WS {
				if val.LeagueID == league.LeagueID {
					// check the match_id with that of the client
					// send the updated price to the client
					err := conn.WriteMessage(websocket.TextMessage, data)
					if err != nil {
						conn.Close()
					}
				}
			}
			app.ClientsM.Unlock()
		}(league)
	}

}

// Get points for a player from redis cache for a league
// This is generally used to show the graph representation of points for a player in a league
func (app *App) GetPointsPlayerWise(w http.ResponseWriter, r *http.Request) {
	leagueID := r.URL.Query().Get("league_id")
	playerID := r.URL.Query().Get("player_id")
	if leagueID == "" || playerID == "" {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "required params missing"})

	}

	points, err := trade.New(app.KVStore, app.DB, app.MQConn).GetTimeseriesPlayerPoints(playerID, leagueID)

	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"data": points}})
}
