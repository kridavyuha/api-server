package trade

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/kridavyuha/api-server/internals/cache"
	"github.com/kridavyuha/api-server/pkg/kvstore"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type TradeService struct {
	KV   kvstore.KVStore
	DB   *gorm.DB
	Conn *amqp.Connection
}

func New(kv kvstore.KVStore, db *gorm.DB, conn *amqp.Connection) *TradeService {
	return &TradeService{
		KV:   kv,
		DB:   db,
		Conn: conn,
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (ts *TradeService) getPlayerPriceList(leagueId, playerId string) ([]string, error) {
	players, err := ts.KV.LRange("players_"+leagueId+"_"+playerId, 0, -1)

	if err != nil {
		return nil, err
	}

	if len(players) == 0 {

		players, err = cache.New(ts.DB, ts.KV).LoadPlayerData(leagueId, playerId)
		if err != nil {
			return nil, err
		}

	}
	return players, nil
}

func (ts *TradeService) getPurse(userId int, leagueId string) (string, error) {
	balanceAndRemainingTxnsStr, err := ts.KV.Get("purse_" + strconv.Itoa(userId) + "_" + leagueId)

	if err != nil {
		if err == redis.Nil {
			balanceAndRemainingTxnsStr, err = cache.New(ts.DB, ts.KV).LoadUserBalanceAndRemainingTxns(leagueId, strconv.Itoa(userId))
			if err != nil {
				return "0", err
			}
		} else {
			return "0", err
		}
	}

	return balanceAndRemainingTxnsStr, nil
}

func (ts *TradeService) getCurPrice(league_id string, player_id string) (float64, string, error) {

	playerData, err := ts.getPlayerPriceList(league_id, player_id)
	if err != nil {
		return 0, "", err
	}

	TsAndPrice := strings.Split(playerData[len(playerData)-1], ",")
	if len(TsAndPrice) != 2 {
		return 0, "", fmt.Errorf("invalid data format for price and timestamp")
	}

	price, err := strconv.ParseFloat(TsAndPrice[1], 64)
	if err != nil {
		return 0, "", err
	}
	return price, TsAndPrice[0], nil
}

func (ts *TradeService) getBasePrice(league_id string, player_id string) (float64, error) {
	playerData, err := ts.getPlayerPriceList(league_id, player_id)
	if err != nil {
		return 0, err
	}

	TsAndPrice := strings.Split(playerData[0], ",")
	if len(TsAndPrice) != 2 {
		return 0, fmt.Errorf("invalid data format for price and timestamp")
	}

	price, err := strconv.ParseFloat(TsAndPrice[1], 64)
	if err != nil {
		return 0, err
	}
	return price, nil
}

func (ts *TradeService) getCurAndBasePrice(league_id string, player_id string) (float64, float64, error) {

	playerData, err := ts.getPlayerPriceList(league_id, player_id)
	if err != nil {
		return 0, 0, err
	}

	CurTsAndPrice := strings.Split(playerData[len(playerData)-1], ",")
	if len(CurTsAndPrice) != 2 {
		return 0, 0, fmt.Errorf("invalid data format for price and timestamp")
	}

	cur_price, err := strconv.ParseFloat(CurTsAndPrice[1], 64)
	if err != nil {
		return 0, 0, err
	}

	BaseTsAndPrice := strings.Split(playerData[0], ",")
	if len(CurTsAndPrice) != 2 {
		return 0, 0, fmt.Errorf("invalid data format for price and timestamp")
	}

	base_price, err := strconv.ParseFloat(BaseTsAndPrice[1], 64)
	if err != nil {
		return 0, 0, err
	}

	return cur_price, base_price, nil
}

func (ts *TradeService) Transaction(transactionType, playerId, leagueId string, userId int, transactionDetails TransactionDetails) error {

	// Check leagues_status if active proceed else return error
	var leagueStatus string
	err := ts.DB.Table("leagues").Select("league_status").Where("league_id = ?", leagueId).Scan(&leagueStatus).Error
	if err != nil {
		return fmt.Errorf("error getting league status: %v", err)
	}

	if leagueStatus != "active" && leagueStatus != "open" {
		return fmt.Errorf("league not active, Transaction cannot be proccessed")
	}

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

	// Get the user's balance from the purse table

	userRemainingPointsAndRemainingTxnsStr, err := ts.getPurse(userId, leagueId)
	if err != nil {
		return err
	}

	userRemainingPointsStr := strings.Split(userRemainingPointsAndRemainingTxnsStr, ",")[0]
	userRemainingPoints, err := strconv.ParseFloat(userRemainingPointsStr, 64)
	if err != nil {
		return err
	}
	userRemainingPoints = math.Round(userRemainingPoints*100) / 100

	var ownedShares int

	// Check in portfolio_user_id_league_id hash for this player_id field
	_, err = ts.KV.HGet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, "is_cached")
	if err != nil {
		if err == redis.Nil {
			err = cache.New(ts.DB, ts.KV).LoadUserPortfolioData(leagueId, strconv.Itoa(userId))
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	sharesAndAvgPriceStr, err := ts.KV.HGet("portfolio_"+strconv.Itoa(userId)+"_"+leagueId, playerId)

	switch {
	case err == redis.Nil:
		sharesAndAvgPriceStr = "0,0"
	case err != nil:
		return err
	}

	sharesAndAvgPrice := strings.Split(sharesAndAvgPriceStr, ",")
	if len(sharesAndAvgPrice) != 2 {
		return fmt.Errorf("invalid data format for shares and invested")
	}

	ownedShares, err = strconv.Atoi(sharesAndAvgPrice[0])
	if err != nil {
		return err
	}

	if transactionType == "buy" {
		// Check the user's balance from the purse table
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
	}

	transactionData := map[string]interface{}{
		"player_id":        playerId,
		"league_id":        leagueId,
		"user_id":          userId,
		"transaction_type": transactionType,
		"shares":           transactionDetails.Shares,
	}

	fmt.Println("Publishing transaction to the queue .. ", transactionData)

	err = ts.PublishTransaction(transactionData)
	if err != nil {
		return err
	}

	return nil
}

func (ts *TradeService) PublishTransaction(transactionData map[string]interface{}) error {
	// This function will be called by the core service to push the transaction to the queue
	// Create a channel
	fmt.Println("Publishing transaction to the queue")
	ch, err := ts.Conn.Channel()
	if err != nil {
		return fmt.Errorf("error creating channel: %s", err)
	}
	err = ch.ExchangeDeclare(
		"txns",   // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	defer ch.Close()

	// Publish the transaction to the queue
	transactionJSON, err := json.Marshal(transactionData)
	if err != nil {
		return fmt.Errorf("error marshalling transaction data")
	}

	err = ch.Publish(
		"txns", // exchange
		fmt.Sprintf("league.%s", transactionData["league_id"].(string)), // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        transactionJSON,
			Timestamp:   time.Now(),
		})

	if err != nil {
		return fmt.Errorf("error publishing transaction to the queue: %s", err)
	}

	return nil
}

func (ts *TradeService) GetPlayerDetails(leagueId string, userId int) ([]GetPlayerDetails, error) {

	playerDetails := make([]GetPlayerDetails, 0)

	players, err := ts.KV.Keys("players_" + leagueId + "_" + "*")
	if err != nil {
		return playerDetails, err
	}

	//TODO: what if only some players were missing
	if len(players) == 0 {
		// load from table and cache it
		err := cache.New(ts.DB, ts.KV).LoadPlayers(leagueId)
		if err != nil {
			return playerDetails, err
		}
		players, err = ts.KV.Keys("players_" + leagueId + "_" + "*")
		if err != nil {
			return playerDetails, err
		}
	}


	portfolio, err := ts.KV.HGetAll("portfolio_" + strconv.Itoa(userId) + "_" + leagueId)

	if err != nil {
		switch {
		case err == redis.Nil:
			// load from table and cache it
			err := cache.New(ts.DB, ts.KV).LoadUserPortfolioData(leagueId, strconv.Itoa(userId))
			if err != nil {
				return nil, err
			}
			portfolio, err = ts.KV.HGetAll("portfolio_" + strconv.Itoa(userId) + "_" + leagueId)
			if err != nil {
				return nil, err
			}

		default:
			return playerDetails, err
		}
	}

	_, err = ts.KV.Get("purse_" + strconv.Itoa(userId) + "_" + leagueId)
	if err != nil {
		switch {
		case err == redis.Nil:
			_, err = cache.New(ts.DB, ts.KV).LoadUserBalanceAndRemainingTxns(leagueId, strconv.Itoa(userId))
			if err != nil {
				return nil, err
			}
		default:
			return playerDetails, err
		}
	}

	for _, player := range players {

		var p GetPlayerDetails
		playerId := strings.Split(player, "_")[2]

		cur_price, base_price, err := ts.getCurAndBasePrice(leagueId, playerId)

		if err != nil {
			return playerDetails, err
		}

		// cant i get all the players data in a single go ...
		// what if i store metadata of all players in a match/ league in a same map. only one Hget ...
		playerMeta, err := ts.KV.HGetAll(fmt.Sprintf("players_%s", playerId))
		if err != nil || len(playerMeta) == 0 {
			switch {
			case err == redis.Nil || len(playerMeta) == 0:
				{
					p, err := cache.New(ts.DB, ts.KV).LoadPlayerMetaData(playerId)
					if err != nil {
						return nil, err
					}
					playerMeta["player_name"] = p.PlayerName
					playerMeta["team"] = p.Team
				}
			default:
				return nil, err

			}
		}

		var share int

		if shares, ok := portfolio[playerId]; ok {
			sharesInvested := strings.Split(shares, ",")
			if len(sharesInvested) != 2 {
				return playerDetails, fmt.Errorf("invalid data format for shares and invested")
			}

			share, err = strconv.Atoi(sharesInvested[0])
			if err != nil {
				return playerDetails, err
			}

		} else {
			share = 0
		}

		p.CurPrice = cur_price
		// p.LastChange = timestamp
		p.PlayerID = playerId
		p.BasePrice = base_price
		p.PlayerName = playerMeta["player_name"]
		p.Team = playerMeta["team"]
		p.Shares = share
		playerDetails = append(playerDetails, p)
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

func (ts *TradeService) GetRemainingTxns(userId int, leagueId string) (string, error) {
	balanceAndRemainingTxnsStr, err := ts.KV.Get("purse_" + strconv.Itoa(userId) + "_" + leagueId)
	if err != nil {
		return "", err
	}
	return strings.Split(balanceAndRemainingTxnsStr, ",")[1], nil
}
