package main

import (
	"github.com/kridavyuha/api-server/pkg/kvstore"

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
	KVStore  kvstore.KVStore
	MQConn   *amqp.Connection
}

func main() {

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	app := &App{
		WS:     make(map[*websocket.Conn]WSDetails),
		MQConn: conn,
	}

	db, err := app.initDB()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	// CORS middleware configuration
	r.Use(cors.New(cors.Options{
		AllowedOrigins:   []string{"https://main.d2j3qqk7sh27x4.amplifyapp.com/"}, // Your frontend URL
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
	app.initConsumer()

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}

}
