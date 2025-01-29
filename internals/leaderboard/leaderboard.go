package leaderboard

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kridavyuha/api-server/internals/portfolio"
	"github.com/kridavyuha/api-server/pkg/kvstore"

	"gorm.io/gorm"
)

type Leaderboard struct {
	KVStore kvstore.KVStore
	DB      *gorm.DB
	ps      *portfolio.PortfolioService
}

func New(kv kvstore.KVStore, db *gorm.DB) *Leaderboard {
	return &Leaderboard{
		KVStore: kv,
		DB:      db,
		ps:      portfolio.New(kv, db),
	}
}

// GetLeaderboard function

func (l *Leaderboard) GetLeaderboard(leagueId string) ([]score, error) {
	// First step -> Get all the registered users of this league from the leagues table using the league_id
	var participatingUsers string
	scores := make([]score, 0)
	err := l.DB.Raw("SELECT users_registered FROM leagues WHERE league_id = ?", leagueId).Scan(&participatingUsers).Error
	if err != nil {
		return nil, err
	}

	// Second step -> Go through each user, calculate their points by 1) fetching the remaining purse from the `purse` table 2) getting the portfolio of the user & calculate the sum of points.
	userIds := strings.Split(participatingUsers, ",")
	for _, userId := range userIds {
		score := &score{}
		userid, _ := strconv.Atoi(userId)
		resp, err := l.ps.GetDetailedPortfolio(userid, leagueId)
		if err != nil {
			return nil, err
		}
		score.UserId = userid
		score.Points = resp.Balance
		// Now go through each player from this user's portfolio and get their worth
		for _, player := range resp.Players {
			score.Points += player.CurPrice * float64(player.Shares)
		}
		l.DB.Raw("SELECT user_name FROM users WHERE user_id = ?", userid).Scan(&score.UserName)
		scores = append(scores, *score)
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Points > scores[j].Points
	})

	return scores, nil

}
