package main

import (
	"fmt"

	"github.com/kridavyuha/api-server/pkg/conf"
	"github.com/kridavyuha/api-server/pkg/kvstore"
	"github.com/spf13/viper"

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
	Config   *viper.Viper
}

func main() {

	app := &App{
		WS: make(map[*websocket.Conn]WSDetails),
	}
	app.Config = conf.Config(".")

	conn, err := amqp.Dial(app.Config.GetString("App.QUEUE_URL"))
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	app.MQConn = conn

	db, err := app.initDB()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	// CORS middleware configuration

	r.Use(cors.New(cors.Options{
		AllowedOrigins:   app.Config.GetStringSlice("App.CORS_ALLOWED_ORIGINS"), // Your frontend URL
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}).Handler)

	app.DB = db
	app.R = r
	app.Config = conf.Config(".")

	// create a map relation btw  player name and player_id
	// app.initQueue()
	app.initHandlers()
	app.initKVStore()
	app.initConsumer()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", app.Config.GetInt("app.SERVER_PORT")), r); err != nil {
		panic(err)
	}

}
