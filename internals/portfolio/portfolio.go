package portfolio

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kridavyuha/api-server/internals/cache"
	"github.com/kridavyuha/api-server/pkg/kvstore"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type PortfolioService struct {
	KV kvstore.KVStore
	DB *gorm.DB
}

func New(kv kvstore.KVStore, db *gorm.DB) *PortfolioService {
	return &PortfolioService{
		KV: kv,
		DB: db,
	}
}

func (ps *PortfolioService) getPurse(userId int, leagueId string) (string, error) {
	balanceAndRemainingTxnsStr, err := ps.KV.Get("purse_" + strconv.Itoa(userId) + "_" + leagueId)

	if err != nil {
		if err == redis.Nil {
			balanceAndRemainingTxnsStr, err = cache.New(ps.DB, ps.KV).LoadUserBalanceAndRemainingTxns(leagueId, strconv.Itoa(userId))
			if err != nil {
				return "0", err
			}
		} else {
			return "0", err
		}
	}

	return balanceAndRemainingTxnsStr, nil
}

func (ps *PortfolioService) getPlayerPriceList(leagueId, playerId string) ([]string, error) {
	players, err := ps.KV.LRange("players_"+leagueId+"_"+playerId, 0, -1)

	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		players, err = cache.New(ps.DB, ps.KV).LoadPlayerData(leagueId, playerId)
		if err != nil {
			return nil, err
		}

	}
	return players, nil
}

func (ps *PortfolioService) getCurPrice(league_id string, player_id string) (float64, string, error) {
	playerData, err := ps.getPlayerPriceList(league_id, player_id)
	if err != nil {
		return 0, "", err
	}

	if len(playerData) == 0 {
		return 0, "", fmt.Errorf("no price data available for player")
	}

	lastEntry := playerData[len(playerData)-1]
	TsAndPrice := strings.Split(lastEntry, ",")
	if len(TsAndPrice) != 2 {
		return 0, "", fmt.Errorf("invalid data format for price and timestamp")
	}

	price, err := strconv.ParseFloat(TsAndPrice[1], 64)
	if err != nil {
		return 0, "", err
	}
	return price, TsAndPrice[0], nil
}

func (ps *PortfolioService) GetPortfolio(user_id int, league_id string) ([]Portfolio, error) {
	// Get the portfolio of the user from the DB
	var portfolio []Portfolio
	_, err := ps.KV.HGet("portfolio_"+strconv.Itoa(user_id)+"_"+league_id, "is_cached")

	if err != nil {
		if err == redis.Nil {
			cache.New(ps.DB, ps.KV).LoadUserPortfolioData(league_id, strconv.Itoa(user_id))
		} else {
			return nil, err
		}
	}

	porfolio, err := ps.KV.HGetAll("portfolio_" + strconv.Itoa(user_id) + "_" + league_id)
	if err != nil {
		return nil, err
	}
	for key, value := range porfolio {
		if key != "is_cached" {
			data := strings.Split(value, ",")
			shares, _ := strconv.Atoi(data[0])
			avg_price, _ := strconv.ParseFloat(data[1], 64)
			portfolio = append(portfolio, Portfolio{PlayerId: key, Shares: shares, AvgPrice: avg_price})
		}
	}

	return portfolio, nil
}

func (ps *PortfolioService) GetDetailedPortfolio(user_id int, league_id string) (DetailedPortfolio, error) {
	// Get the portfolio of the user from the DB
	var detailedPortfolio DetailedPortfolio
	portfolio, err := ps.GetPortfolio(user_id, league_id)
	if err != nil {
		return detailedPortfolio, err
	}

	// Using LeagueID and PlayerID, get the player name and current price
	for i, player := range portfolio {
		price, _, err := ps.getCurPrice(league_id, player.PlayerId)
		if err != nil {
			return detailedPortfolio, err
		}
		var playerInfo struct {
			PlayerName string
			Team       string
		}
		ps.DB.Raw("SELECT player_name,team FROM players WHERE player_id = ?", player.PlayerId).Scan(&playerInfo)
		portfolio[i].CurPrice = price
		portfolio[i].PlayerName = playerInfo.PlayerName
		portfolio[i].TeamName = playerInfo.Team
	}

	// Get the remaining purse balance
	balance, err := ps.getPurse(user_id, league_id)
	if err != nil {
		return detailedPortfolio, err
	}

	balanceFloat, err := strconv.ParseFloat(strings.Split(balance, ",")[0], 64)
	if err != nil {
		return detailedPortfolio, err
	}
	detailedPortfolio.Balance = balanceFloat
	detailedPortfolio.Players = portfolio

	return detailedPortfolio, nil
}
