package trade

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

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

func priceChange(shares int, cur_price int) float64 {
	// Lets say the threshold by which we change price for a transaction is x
	// x = 0.1% of cur_price

	//

	return 0.0
}

// TODO: change return type to float64, err as we need to send no.of shares/sale price for buy or sell respectively.
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
	var ownedShares float64

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

	ownedShares, err = strconv.ParseFloat(sharesInvested[0], 64)
	if err != nil {
		return err
	}

	invested, err := strconv.Atoi(sharesInvested[1])
	if err != nil {
		return err
	}

	// ----------------------------------------------

	var avgPriceOfTransaction float64

	if transactionType == "buy" {
		// Check the user's balance from the purse table
		estimatedPurchaseAmount := transactionDetails.Shares * transactionDetails.PlayerCurrentPrice
		// Check if the user has enough balance
		if userRemainingPoints < estimatedPurchaseAmount {
			return fmt.Errorf("insufficient balance")
		}

		sharesBought := float64(0)
		varyingPlayerPrice := transactionDetails.PlayerCurrentPrice
		for estimatedPurchaseAmount >= varyingPlayerPrice {
			estimatedPurchaseAmount -= varyingPlayerPrice
			varyingPlayerPrice++
			sharesBought++
		}
		// Now buy whatever fraction of share the remaining amount could afford
		fractionOfShares := estimatedPurchaseAmount / varyingPlayerPrice
		sharesBought += fractionOfShares
		varyingPlayerPrice += fractionOfShares

		avgPriceOfTransaction = estimatedPurchaseAmount / sharesBought

		// TODO: Use varyingPlayerPrice as the new player price

		// Reduce b
		// userRemainingPoints = userRemainingPoints - totalCost
	} else if transactionType == "sell" {

		if ownedShares < transactionDetails.Shares {
			return fmt.Errorf("insufficient shares")
		}

		intPartOfTransactShares := math.Floor(transactionDetails.Shares)
		fractionalPartOfTransactShares := transactionDetails.Shares - intPartOfTransactShares

		// calculate the sale worth
		var saleWorth float64

		saleWorth = ((intPartOfTransactShares)*(transactionDetails.PlayerCurrentPrice) - ((intPartOfTransactShares)*(intPartOfTransactShares-1))/2) + (fractionalPartOfTransactShares * (transactionDetails.PlayerCurrentPrice - intPartOfTransactShares + 1))

		avgPriceOfTransaction = saleWorth / transactionDetails.Shares

		userRemainingPoints = userRemainingPoints + saleWorth
	}

	// Create an entry in transactions table
	err = ts.DB.Exec("INSERT INTO transactions (user_id, player_id, league_id, shares, price, transaction_type, transaction_time) VALUES (?, ?, ?, ?, ?, ?, ?)", userId, playerId, leagueId, transactionDetails.Shares, avgPriceOfTransaction, transactionType, time.Now()).Error
	if err != nil {
		return err
	}

	// Also update the purse table
	err = ts.DB.Exec("UPDATE purse SET remaining_purse = ? WHERE user_id = ? AND league_id = ?", userRemainingPoints, userId, leagueId).Error
	if err != nil {
		return err
	}
	// Update the user's balance in cache ..
	err = ts.KV.Set("purse_"+strconv.Itoa(userId)+"_"+leagueId, strconv.FormatFloat(userRemainingPoints, 'g', 5, 64))
	if err != nil {
		return err
	}

	// Update the portfolio table
	// If the player is already present in the portfolio, update the shares and average price
	// If the player is not present, insert a new row
	err = ts.UpdatePortfolio(transactionType, userId, playerId, leagueId, transactionDetails, ownedShares, invested)
	if err != nil {
		return err
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

func (ts *TradeService) UpdatePortfolio(transactionType string, userId int, playerId string, leagueId string, transactionDetails TransactionDetails, shares int, invested int) error {

	if transactionType == "buy" {

		// Update or insert the shares if player is already in the portfolio
		if shares == 0 {
			// Update in table first...
			err := ts.DB.Exec("INSERT INTO portfolio (user_id, player_id, league_id, shares, invested) VALUES (?, ?, ?, ?, ?)", userId, playerId, leagueId, transactionDetails.Shares, transactionDetails.Shares*transactionDetails.PlayerCurrentPrice).Error
			if err != nil {
				return err
			}

			// Update in cache
			err = ts.KV.HSet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId, strconv.Itoa(transactionDetails.Shares)+","+strconv.Itoa(transactionDetails.PlayerCurrentPrice*transactionDetails.Shares))
			if err != nil {
				return err
			}
		} else {

			err := ts.DB.Exec("UPDATE portfolio SET shares = shares + ?, invested = invested + ? WHERE user_id = ? AND player_id = ? AND league_id = ?", transactionDetails.Shares, transactionDetails.PlayerCurrentPrice, userId, playerId, leagueId).Error
			if err != nil {
				return err
			}
			// Update in cache
			invested += transactionDetails.PlayerCurrentPrice * transactionDetails.Shares
			shares += transactionDetails.Shares
			err = ts.KV.HSet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId, strconv.Itoa(shares)+","+strconv.Itoa(invested))
			if err != nil {
				return err
			}
		}
	} else if transactionType == "sell" {
		// Update or delete the shares if player is already in the portfolio

		shares -= transactionDetails.Shares
		invested -= transactionDetails.PlayerCurrentPrice * transactionDetails.Shares

		err := ts.DB.Exec("UPDATE portfolio SET shares = ?, invested = invested - ? WHERE user_id = ? AND player_id = ? AND league_id = ?", shares, transactionDetails.PlayerCurrentPrice*transactionDetails.Shares, userId, playerId, leagueId).Error
		if err != nil {
			return err
		}
		// update in cache ..
		err = ts.KV.HSet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId, strconv.Itoa(shares)+","+strconv.Itoa(invested))
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
