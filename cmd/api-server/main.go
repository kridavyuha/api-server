package main

import (
	KVStore "backend/pkg"
	"log"

	"net/http"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
	"gorm.io/gorm"
)

type App struct {
	DB       *gorm.DB
	R        *chi.Mux
	WS       map[*websocket.Conn]WSDetails
	ClientsM sync.Mutex
	KVStore  KVStore.KVStore
	Ch       *amqp.Channel
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	app := &App{
		WS: make(map[*websocket.Conn]WSDetails),
		Ch: ch,
	}

	db, err := app.initDB()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	// CORS middleware configuration
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"}, // Your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}).Handler)

	app.DB = db
	app.R = r

	// create a map relation btw  player name and player_id
	// app.initQueue()
	app.initHandlers()
	app.initKVStore()
	app.initTxnQueue()

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

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}

}
