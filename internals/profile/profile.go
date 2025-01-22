package profile

import (
	"github.com/kridavyuha/api-server/internals/leagues"
	"github.com/kridavyuha/api-server/pkg/kvstore"

	"gorm.io/gorm"
)

type ProfileService struct {
	KV kvstore.KVStore
	DB *gorm.DB
	LS *leagues.LeagueService
}

func New(kv kvstore.KVStore, db *gorm.DB) *ProfileService {
	return &ProfileService{
		KV: kv,
		DB: db,
		LS: leagues.New(kv, db),
	}
}

func (ps *ProfileService) GetProfile(userId int) (CompleteProfile, error) {
	// Fetch user details from Users table
	var completeProfile CompleteProfile
	err := ps.DB.Table("users").Select("user_id, user_name, mail_id, profile_pic,credits, rating").Where("user_id = ?", userId).Scan(&completeProfile.Profile).Error
	if err != nil {
		return completeProfile, err
	}

	// Get my leagues..
	leagues, err := ps.LS.GetMyLeagues(userId)
	if err != nil {
		return completeProfile, err
	}

	completeProfile.Leagues = leagues

	return completeProfile, nil
}
