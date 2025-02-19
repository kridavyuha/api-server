package main

import (
	"fmt"
	"log"
	"time"

	"github.com/kridavyuha/api-server/internals/leagues"
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
	// initialize db
	db, err := gorm.Open(postgres.Open(app.Config.GetString("App.DB_URL")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (app *App) initKVStore() error {
	// initialize redis
	kvstore, err := kvstore.NewRedis(app.Config.GetString("App.REDIS_URL"), "", 0)
	if err != nil {
		fmt.Println("Error initializing KVStore: ", err)
		return err
	}
	app.KVStore = kvstore
	return nil
}

func (app *App) initConsumer() {

	fmt.Println("Initializing consumer")
	ch, err := app.MQConn.Channel()
	failOnError(err, "Failed to open a channel")
	err = ch.ExchangeDeclare(
		"balls",  // name
		"fanout", // type
		true,     // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	failOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		q.Name,  // queue name
		"",      // routing key
		"balls", // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind a queue")
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
			app.BallPicker(d.Body)
		}
	}()

}

func (app *App) initCoreProcessing() {
	// Fetch league status from db -> open/active -> Run this for every 30 sec
	// Run a ticker
	// for every 1 sec
	fetchUpdatedScoresTicker := time.NewTicker(2 * time.Second)

	go func() {
		for {
			select {
			case <-fetchUpdatedScoresTicker.C:
				// Fetch the leagues table from the database and get the leagues which are not in status 'active'
				leagues, err := leagues.New(app.KVStore, app.DB).GetLeaguesGlobally()
				if err != nil {
					fmt.Println("Error fetching leagues from the database: ", err)
					continue
				}
				// Now run through these leagues and get the cur price of each player from the map
				app.HandleUpdateCorePrices(leagues)
			}
		}
	}()

}
