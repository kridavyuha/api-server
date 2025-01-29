package cache

import (
	"fmt"
	"strconv"

	"github.com/kridavyuha/api-server/pkg/kvstore"
	"gorm.io/gorm"
)

type CacheService struct {
	DB *gorm.DB
	KV kvstore.KVStore
}

func New(db *gorm.DB, kv kvstore.KVStore) *CacheService {
	return &CacheService{
		DB: db,
		KV: kv,
	}
}

func (c *CacheService) LoadPlayerData(league_id, player_id string) ([]string, error) {
	// get player data from the database
	tableName := "players_" + league_id

	var playerData PlayerInLeague

	err := c.DB.Table(tableName).Select("cur_price, last_updated").Where("player_id = ?", player_id).Scan(&playerData).Error
	if err != nil {
		return nil, err
	}

	// insert into redis cache
	key := "players_" + league_id + "_" + player_id
	value := fmt.Sprintf("%s,%.2f", playerData.LastUpdated, playerData.CurPrice)

	err = c.KV.RPush(key, value)

	if err != nil {
		return nil, err
	}

	return []string{player_id, value}, nil

}

func (c *CacheService) LoadUserBalance(league_id, user_id string) (float64, error) {
	// get user balance from the database
	var balance float64

	err := c.DB.Table("purse").Select("remaining_purse").Where("user_id = ? and league_id = ?", user_id, league_id).Scan(&balance).Error
	if err != nil {
		return 0, err
	}

	// insert into redis cache
	key := "purse_" + user_id + "_" + league_id
	value := fmt.Sprintf("%.2f", balance)

	err = c.KV.Set(key, value)

	if err != nil {
		return 0, err
	}

	return balance, nil
}

func (c *CacheService) LoadUserPortfolioData(league_id, user_id string) error {
	var portfolio []Portfolio

	err := c.DB.Table("portfolio").Where("user_id = ? AND league_id = ?", user_id, league_id).Scan(&portfolio).Error
	if err != nil {
		return err
	}

	// insert into redis cache
	key := "portfolio_" + user_id + "_" + league_id

	err = c.KV.HSet(key, "is_cached", "active")

	// loop through the portfolio and insert into redis cache
	for _, player := range portfolio {
		value := strconv.Itoa(player.Shares) + "," + fmt.Sprintf("%.2f", player.AvgPrice)
		err = c.KV.HSet(key, player.PlayerId, value)
		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (c *CacheService) LoadPlayers(league_id string) error {
	// get all the players from the database
	tableName := "players_" + league_id

	var players []PlayerInLeague

	err := c.DB.Table(tableName).Scan(&players).Error
	if err != nil {
		return err
	}

	// insert into redis cache
	for _, player := range players {
		key := "players_" + league_id + "_" + player.PlayerID

		timestamp := player.LastUpdated.Format("2006-01-02 15:04:05.000000-07")
		value := fmt.Sprintf("%s,%.2f", timestamp, player.CurPrice)

		err = c.KV.RPush(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}
