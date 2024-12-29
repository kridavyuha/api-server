// backend/models/setup.go
package models

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Trade struct {
	gorm.Model
	PlayerID uint    `json:"playerId"`
	Price    float64 `json:"price"`
	Action   string  `json:"action"` // buy/sell
	UserID   uint    `json:"userId"`
}

type Player struct {
	gorm.Model
	Name         string  `json:"name"`
	ShortCode    string  `json:"shortCode"`
	CurrentPrice float64 `json:"currentPrice"`
	Team         string  `json:"team"`
}

func SetupDB() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// // Auto Migrate
	// err = db.AutoMigrate(&Trade{}, &Player{})
	// if err != nil {
	// 	return nil, err
	// }

	return db, nil
}
