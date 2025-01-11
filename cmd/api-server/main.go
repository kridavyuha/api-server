package main

import (
	KVStore "backend/pkg"

	"net/http"
	"sync"

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
}

func main() {

	app := &App{
		WS: make(map[*websocket.Conn]WSDetails),
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

	app.initHandlers()
	app.initKVStore()

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}

}
