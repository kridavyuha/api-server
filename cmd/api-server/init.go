package main

import (
	"log"

	"github.com/kridavyuha/api-server/pkg/kvstore"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func (app *App) initDB() (*gorm.DB, error) {
	dsn := "host=localhost user=postgres password=postgres dbname=db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *App) initKVStore() {
	// initialize redis
	app.KVStore = kvstore.NewRedis("localhost:6379", "", 0)
}

func (app *App) initTxnQueue() {
	_, err := app.Ch.QueueDeclare(
		"txns", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	failOnError(err, "Failed to declare a queue")
}
