package trade

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kridavyuha/api-server/pkg/kvstore"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type TradeService struct {
	KV kvstore.KVStore
	DB *gorm.DB
}

func New(kv kvstore.KVStore, db *gorm.DB) *TradeService {
	return &TradeService{
		KV: kv,
		DB: db,
	}
}

func (ts *TradeService) getPlayerPriceList(leagueId, playerId string) ([]string, error) {
	players, err := ts.KV.LRange("players_"+leagueId+"_"+playerId, 0, -1)

	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		// TODO: load the table data into cache.
		// If the player is not found in the cache, get the player details from the players table
		// Load players from the players table to the cache
		// If the player is not found in the players table, return an error
		return nil, fmt.Errorf("player not found")
	}
	return players, nil
}

func (ts *TradeService) getPurse(userId int, leagueId string) (float64, error) {
	balanceStr, err := ts.KV.Get("purse_" + strconv.Itoa(userId) + "_" + leagueId)
	if err != nil {
		if err == redis.Nil {
			// TODO: load table data into cache
			// Load the purse from the table to cache
			// if not in table return err.
		} else {
			return 0, err
		}
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (ts *TradeService) getCurPrice(league_id string, player_id string) (int, string, error) {

	playerData, err := ts.getPlayerPriceList(league_id, player_id)
	if err != nil {
		return 0, "", err
	}

	TsAndPrice := strings.Split(playerData[len(playerData)-1], ",")
	if len(TsAndPrice) != 2 {
		return 0, "", fmt.Errorf("invalid data format for price and timestamp")
	}

	price, err := strconv.Atoi(TsAndPrice[1])
	if err != nil {
		return 0, "", err
	}
	return price, TsAndPrice[0], nil
}

func (ts *TradeService) getBasePrice(league_id string, player_id string) (int, error) {
	playerData, err := ts.getPlayerPriceList(league_id, player_id)
	if err != nil {
		return 0, err
	}

	TsAndPrice := strings.Split(playerData[0], ",")
	if len(TsAndPrice) != 2 {
		return 0, fmt.Errorf("invalid data format for price and timestamp")
	}

	price, err := strconv.Atoi(TsAndPrice[1])
	if err != nil {
		return 0, err
	}
	return price, nil
}

func (ts *TradeService) Transaction(transactionType, playerId, leagueId string, userId int, transactionDetails TransactionDetails) error {

	// Here we simultaneously update the transactions, purse and portfolio tables in db and cache for consistency
	// the purchase rate will be calculated by the core service once the transaction is successful this will be sent to queue.

	players, err := ts.getPlayerPriceList(leagueId, playerId)
	if err != nil {
		return err
	}

	priceAndTs := strings.Split(players[len(players)-1], ",")
	if len(priceAndTs) != 2 {
		return fmt.Errorf("invalid data format for price and timestamp")
	}

	transactionDetails.PlayerCurrentPrice, err = strconv.ParseFloat(priceAndTs[1], 64)
	if err != nil {
		return err
	}
	// ----------------------------------------------
	// Get the user's balance from the purse table

	userRemainingPoints, err := ts.getPurse(userId, leagueId)
	if err != nil {
		return err
	}

	// ----------------------------------------------
	var ownedShares int

	// Check in portfolio_user_id_league_id hash for this player_id field
	_, err = ts.KV.HGet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, "is_cached")
	if err != nil {
		if err == redis.Nil {
			//TODO: fetch the portfolio from the table and load it into cache, along side is_cached active
		} else {
			return err
		}
	}
	sharesInvestedStr, err := ts.KV.HGet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId)

	switch {
	case err == redis.Nil:
		sharesInvestedStr = "0,0"
	case err != nil:
		return err
	}

	sharesInvested := strings.Split(sharesInvestedStr, ",")
	if len(sharesInvested) != 2 {
		return fmt.Errorf("invalid data format for shares and invested")
	}

	ownedShares, err = strconv.Atoi(sharesInvested[0])
	if err != nil {
		return err
	}

	invested, err := strconv.ParseFloat(sharesInvested[1], 64)
	if err != nil {
		return err
	}

	fmt.Printf("Shares: %d, Invested: %f\n", ownedShares, invested)

	// ----------------------------------------------

	if transactionType == "buy" {
		// Check the user's balance from the purse table
		// Calculate the total cost
		estimatedPurchaseAmount := float64(transactionDetails.Shares) * transactionDetails.PlayerCurrentPrice
		// Check if the user has enough balance
		if userRemainingPoints < estimatedPurchaseAmount {
			return fmt.Errorf("insufficient balance")
		}
		// balance = balance - totalCost
	} else if transactionType == "sell" {

		if ownedShares < transactionDetails.Shares {
			return fmt.Errorf("insufficient shares")
		}
		// balance = balance + float64(transactionDetails.Shares)*transactionDetails.PlayerCurrentPrice
	}
	return nil
}

func (ts *TradeService) GetPlayerDetails(leagueId string, userId int) ([]GetPlayerDetails, error) {

	playerDetails := make([]GetPlayerDetails, 0)

	// Get the player details from the players_{league_id} table
	// TODO: If this list size is 0 load it from table.
	players, err := ts.KV.Keys("players_" + leagueId + "*")
	if err != nil {
		return playerDetails, err
	}

	for _, player := range players {
		var p GetPlayerDetails
		playerId := strings.Split(player, "_")[2]

		price, timestamp, err := ts.getCurPrice(leagueId, playerId)

		if err != nil {
			return playerDetails, err
		}

		basePrice, err := ts.getBasePrice(leagueId, playerId)

		if err != nil {
			return playerDetails, err
		}
		p.CurPrice = price
		p.LastChange = timestamp
		p.PlayerID = playerId
		p.BasePrice = basePrice
		playerDetails = append(playerDetails, p)
	}

	// Get the player details from the players table
	for i, player := range playerDetails {
		var playerData struct {
			PlayerName string `json:"player_name"`
			Team       string `json:"team"`
			ProfilePic string `json:"profile_pic"`
		}

		err := ts.DB.Raw("SELECT player_name, team FROM players WHERE player_id = ?", player.PlayerID).Scan(&playerData).Error
		if err != nil {
			return playerDetails, err
		}

		playerDetails[i].PlayerName = playerData.PlayerName
		playerDetails[i].Team = playerData.Team
	}

	// Get the share details from the portfolio_{user_id}_{league_id} hash map
	portfolio, err := ts.KV.HGetAll("portfolio_" + strconv.Itoa(userId) + "_" + leagueId)
	if err != nil {
		return playerDetails, err
	}

	for i, player := range playerDetails {
		if shares, ok := portfolio[player.PlayerID]; ok {
			sharesInvested := strings.Split(shares, ",")
			if len(sharesInvested) != 2 {
				return playerDetails, fmt.Errorf("invalid data format for shares and invested")
			}

			playerDetails[i].Shares, err = strconv.Atoi(sharesInvested[0])
			if err != nil {
				return playerDetails, err
			}

		} else {
			playerDetails[i].Shares = 0
		}
	}

	return playerDetails, nil
}

func (ts *TradeService) UpdatePortfolio(transactionType string, userId int, playerId string, leagueId string, transactionDetails TransactionDetails, shares int, invested float64) error {

	if transactionType == "buy" {

		transactionTotalAmount := float64(transactionDetails.Shares) * transactionDetails.PlayerCurrentPrice

		// Update or insert the shares if player is already in the portfolio
		if shares == 0 {
			// Update in table first...
			err := ts.DB.Exec("INSERT INTO portfolio (user_id, player_id, league_id, shares, invested) VALUES (?, ?, ?, ?, ?)", userId, playerId, leagueId, transactionDetails.Shares, transactionTotalAmount).Error
			if err != nil {
				return err
			}

			// Update in cache
			err = ts.KV.HSet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId, strconv.Itoa(transactionDetails.Shares)+","+fmt.Sprintf("%f", transactionTotalAmount))
			if err != nil {
				return err
			}
		} else {
			err := ts.DB.Exec("UPDATE portfolio SET shares = shares + ?, invested = invested + ? WHERE user_id = ? AND player_id = ? AND league_id = ?", transactionDetails.Shares, transactionDetails.PlayerCurrentPrice, userId, playerId, leagueId).Error
			if err != nil {
				return err
			}
			// Update in cache
			invested += transactionDetails.PlayerCurrentPrice * float64(transactionDetails.Shares)
			shares += transactionDetails.Shares
			err = ts.KV.HSet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId, strconv.Itoa(shares)+","+fmt.Sprintf("%f", invested))
			if err != nil {
				return err
			}
		}
	} else if transactionType == "sell" {
		// Update or delete the shares if player is already in the portfolio
		transactionTotalAmount := float64(transactionDetails.Shares) * transactionDetails.PlayerCurrentPrice

		shares -= transactionDetails.Shares
		invested -= transactionDetails.PlayerCurrentPrice * float64(transactionDetails.Shares)

		err := ts.DB.Exec("UPDATE portfolio SET shares = ?, invested = invested - ? WHERE user_id = ? AND player_id = ? AND league_id = ?", shares, transactionTotalAmount, userId, playerId, leagueId).Error
		if err != nil {
			return err
		}
		// update in cache ..
		err = ts.KV.HSet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId, strconv.Itoa(shares)+","+fmt.Sprintf("%f", invested))
		if err != nil {
			return err
		}
	}

	return nil
}

func (ts *TradeService) GetTimeseriesPlayerPoints(player_id, league_id string) ([]string, error) {
	// Get the timeseries data from the players_{league_id}_{player_id} list
	return ts.KV.LRange("players_"+league_id+"_"+player_id, 0, -1)
}
