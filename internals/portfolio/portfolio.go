package portfolio

import (
	KVStore "backend/pkg"
	"fmt"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type PortfolioService struct {
	KV KVStore.KVStore
	DB *gorm.DB
}

func New(kv KVStore.KVStore, db *gorm.DB) *PortfolioService {
	return &PortfolioService{
		KV: kv,
		DB: db,
	}
}

func (ps *PortfolioService) getPurse(userId int, leagueId string) (int, error) {
	balanceStr, err := ps.KV.Get("purse_" + strconv.Itoa(userId) + "_" + leagueId)
	if err != nil {
		if err == redis.Nil {
			// TODO: load table data into cache
			// Load the purse from the table to cache
			// if not in table return err.
		} else {
			return 0, err
		}
	}

	balance, err := strconv.Atoi(balanceStr)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

func (ps *PortfolioService) getPlayerPriceList(leagueId, playerId string) ([]string, error) {
	players, err := ps.KV.LRange("players_"+leagueId+"_"+playerId, 0, -1)

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

func (ps *PortfolioService) getCurPrice(league_id string, player_id string) (int, string, error) {

	playerData, err := ps.getPlayerPriceList(league_id, player_id)
	fmt.Println(playerData)
	if err != nil {
		return 0, "", err
	}

	TsAndPrice := strings.Split(playerData[0], ",")
	if len(TsAndPrice) != 2 {
		return 0, "", fmt.Errorf("invalid data format for price and timestamp")
	}

	price, err := strconv.Atoi(TsAndPrice[1])
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
			var portfolioFromTable []Portfolio
			err = ps.DB.Raw("SELECT player_id, shares, invested FROM portfolio WHERE user_id = ? AND league_id = ?", user_id, league_id).Scan(&portfolioFromTable).Error
			if err != nil {
				return nil, err
			}
			ps.KV.HSet("portfolio_"+strconv.Itoa(user_id)+"_"+league_id, "is_cached", "active")
			for _, player := range portfolioFromTable {
				portfolioData := strconv.Itoa(player.Shares) + "," + strconv.Itoa(player.Invested)
				ps.KV.HSet("portfolio_"+strconv.Itoa(user_id)+"_"+league_id, player.PlayerId, portfolioData)
			}
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
			invested, _ := strconv.Atoi(data[1])
			portfolio = append(portfolio, Portfolio{PlayerId: key, Shares: shares, Invested: invested})
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
	detailedPortfolio.Balance = balance
	detailedPortfolio.Players = portfolio

	return detailedPortfolio, nil
}
