package leagues

import (
	KVStore "backend/pkg"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/exp/rand"
	"gorm.io/gorm"
)

var (
	tableName               = "default_players_table"
	CreatePlayersTableQuery = `create table ` + tableName + ` (
		player_id VARCHAR(6) PRIMARY KEY,
		base_price INT,
		cur_price INT,
		last_change VARCHAR(3) CHECK (last_change IN ('pos', 'neg', 'neu'))
	);`
	InsertPlayerQuery = `INSERT INTO ` + tableName + ` (player_id, base_price, cur_price, last_change) VALUES (?, ?, ?, ?);`
)

type LeagueService struct {
	KV KVStore.KVStore
	DB *gorm.DB
}

func New(kv KVStore.KVStore, db *gorm.DB) *LeagueService {
	return &LeagueService{
		KV: kv,
		DB: db,
	}
}

func generateLeagueID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(uint64(time.Now().UnixNano()))
	b := make([]byte, 8)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func getSquad(team string) (Squad, error) {
	// make a get request to squads end point to get players data.
	var squad Squad

	resp, err := http.Get("http://localhost:8081/squad?team_name=" + team)
	if err != nil {
		return squad, err
	}

	err = json.NewDecoder(resp.Body).Decode(&squad)
	if err != nil {
		return squad, err
	}

	return squad, nil
}

func getFixtures(matchID string) (Fixture, error) {
	resp, err := http.Get("http://localhost:8081/fixtures?match_id=" + matchID)
	if err != nil {
		return Fixture{}, err
	}

	var fixture Fixture
	err = json.NewDecoder(resp.Body).Decode(&fixture)
	if err != nil {
		return fixture, err
	}

	return fixture, nil
}

// CreateLeague function
func (l *LeagueService) CreateLeague(league CreateLeagueRequestBody) error {
	// Generate a unique match_id
	leagueID := generateLeagueID()

	// Create a table name using the match_id
	tableName = "players_" + leagueID

	// Insert players into the newly created table.
	// get team details from fixtures endpoint
	fixture, err := getFixtures(league.MatchID)
	if err != nil {
		return fmt.Errorf("error getting fixtures")
	}

	teams := []string{fixture.TeamA, fixture.TeamB}

	// Get squads for each team
	var playerIDs []string

	for _, team := range teams {
		squad, err := getSquad(team)
		if err != nil {
			return fmt.Errorf("error getting squad")
		}
		for _, player := range squad.Players {
			playerIDs = append(playerIDs, player.Id)
		}
	}

	// Get base price for each player
	var playerBasePrices []struct {
		PlayerID  string `json:"player_id"`
		BasePrice int    `json:"base_price"`
	}

	err = l.DB.Table("base_price").Where("player_id IN ?", playerIDs).Find(&playerBasePrices).Error
	if err != nil {
		return fmt.Errorf("error fetching base prices: %v", err)
	}

	// Create a table for the league
	err = l.DB.Exec(CreatePlayersTableQuery).Error
	if err != nil {
		return fmt.Errorf("error creating table: %v", err)
	}

	for _, player := range playerBasePrices {
		err = l.DB.Exec(InsertPlayerQuery, player.PlayerID, player.BasePrice, player.BasePrice, "neu").Error
		if err != nil {
			return fmt.Errorf("error inserting player: %v", err)
		}
	}

	// Insert the league into the leagues table
	err = l.DB.Table("leagues").Create(&League{
		LeagueID: leagueID,
		MatchID:  league.MatchID,
		EntryFee: league.EntryFee,
		Capacity: league.Capacity,
	}).Error
	if err != nil {
		return fmt.Errorf("error inserting league: %v", err)
	}

	// Create Redis key value pair for the league id and the table name
	// {league_id}_{player_id} is the key and value is the pair of <timestamp, points>
	for _, player := range playerBasePrices {
		key := "players_" + leagueID + "_" + player.PlayerID
		timestamp := time.Now().Unix()
		value := fmt.Sprintf("%d,%d", timestamp, player.BasePrice)
		err = l.KV.RPush(key, value)
		if err != nil {
			return fmt.Errorf("error inserting into KV store: %v", err)
		}
	}
	return nil
}

// GetLeague function
func (l *LeagueService) GetLeagues(user_id int) ([]League, error) {

	var leagues []League
	err := l.DB.Table("leagues").Find(&leagues).Error
	if err != nil {
		return nil, err
	}

	// Get the teams involved from match id
	for i, league := range leagues {
		fixture, err := getFixtures(league.MatchID)
		if err != nil {
			return nil, err
		}
		leagues[i].TeamA = fixture.TeamA
		leagues[i].TeamB = fixture.TeamB

		// Add a field to check if the user is registered for the league
		leagues[i].IsRegistered = false
		if strings.Contains(league.UsersRegistered, fmt.Sprintf("%d", user_id)) {
			leagues[i].IsRegistered = true
		}
	}

	return leagues, nil
}

// RegisterToLeague function
func (l *LeagueService) RegisterToLeague(user_id int, league_id string) error {

	// Get the capacity and registered count from the leagues table
	// Order the struct fields in an optimal way to avoid padding
	var league League

	// Get the league details
	err := l.DB.Table("leagues").Where("league_id = ?", league_id).Scan(&league).Error

	if err != nil {
		return err
	}

	// Check if the league is full
	if league.Registered == league.Capacity {
		return fmt.Errorf("league is full")
	}

	// Add the user to the users_registered list
	newRegisteredUsers := strings.TrimPrefix(league.UsersRegistered+fmt.Sprintf(",%d", user_id), ",")
	league.Registered = league.Registered + 1

	// TODO: @anveshreddy18 : Need to relook on whether to update the league_status here or where the league is started.
	// Update the users_registered,registered column in the leagues table

	err = l.DB.Model(&League{}).Where("league_id = ?", league_id).Updates(map[string]interface{}{"registered": league.Registered, "users_registered": newRegisteredUsers}).Error

	if err != nil {
		return fmt.Errorf("error updating league: %v", err)
	}

	// Also add the user to the purse table
	err = l.DB.Table("purse").Create(map[string]interface{}{"user_id": user_id, "league_id": league_id, "balance": 10000}).Error

	if err != nil {
		return fmt.Errorf("error updating purse: %v", err)
	}

	return nil
}

// Delete 
