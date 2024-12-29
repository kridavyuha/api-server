package main

import (
	KVStore "backend/pkg"
	"net/http"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
	app.KVStore = KVStore.NewRedis("localhost:6379", "", 0)
}

func (app *App) initHandlers() {
	app.R.Get("/ws", app.handleWebSocket)
	app.R.Post("/login", app.Login)
	app.R.Post("/logout", app.Middleware(http.HandlerFunc(app.Logout)))
	app.R.Post("/points", app.PushPoints)
	app.R.Post("/createLeague", app.CreateLeague)
	app.R.Get("/getLeagues", app.Middleware(http.HandlerFunc(app.GetLeagues)))

	app.R.Get("/players", app.Middleware(http.HandlerFunc(app.GetLeague)))
	app.R.Delete("/deleteLeague", app.DeleteLeague)
	app.R.Get("/points", app.Middleware(http.HandlerFunc(app.GetPointsPlayerWise)))
	app.R.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	app.R.Post("/register", app.Middleware(http.HandlerFunc(app.RegisterLeague)))
	app.R.Get("/profile", app.Middleware(http.HandlerFunc(app.GetProfile)))
	app.R.Post("/trade/transaction", app.Middleware(http.HandlerFunc(app.TransactPlayers)))
	app.R.Get("/portfolio", app.Middleware(http.HandlerFunc(app.GetPortfolio)))

}
