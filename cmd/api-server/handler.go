package main

import "net/http"

func (app *App) initHandlers() {
	app.R.Get("/ws", app.handleWebSocket)

	app.R.Post("/auth/login", app.Login)
	app.R.Post("/auth/signup", app.SignUp)
	app.R.Post("/auth/logout", app.Middleware(http.HandlerFunc(app.Logout)))

	app.R.Post("/leagues/create", app.Middleware(http.HandlerFunc(app.CreateLeague)))
	app.R.Get("/leagues", app.Middleware(http.HandlerFunc(app.GetLeagues)))
	app.R.Get("/leagues/delete", app.Middleware(http.HandlerFunc(app.DeleteLeague)))
	app.R.Post("/leagues/register", app.Middleware(http.HandlerFunc(app.RegisterLeague)))

	app.R.Post("/trade/transaction", app.Middleware(http.HandlerFunc(app.TransactPlayers)))
	app.R.Get("/trade", app.Middleware(http.HandlerFunc(app.Trade)))
	app.R.Get("/trade/points", app.Middleware(http.HandlerFunc(app.GetPointsPlayerWise)))

	app.R.Get("/profile", app.Middleware(http.HandlerFunc(app.GetProfile)))

	app.R.Get("/portfolio", app.Middleware(http.HandlerFunc(app.GetPortfolio)))

	app.R.Get("/leaderboard", app.Middleware(http.HandlerFunc(app.GetLeaderboard)))

	// app.R.Post("/points", app.)
	app.R.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

}
