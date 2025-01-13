package profile

import (
	KVStore "backend/pkg"

	"gorm.io/gorm"
)

type ProfileService struct {
	KV KVStore.KVStore
	DB *gorm.DB
}

func New(kv KVStore.KVStore, db *gorm.DB) *ProfileService {
	return &ProfileService{
		KV: kv,
		DB: db,
	}
}

func (ps *ProfileService) GetProfile(userId int) (Profile, error) {
	// Fetch user details from Users table
	var profile Profile
	err := ps.DB.Table("users").Select("user_id, user_name, mail_id, profile_pic").Where("user_id = ?", userId).Scan(&profile).Error
	if err != nil {
		return profile, err
	}
	return profile, nil
}
