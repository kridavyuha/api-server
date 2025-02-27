package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func (app *App) initHandlers() {
	app.R.Get("/ws", app.handleWebSocket)
	app.R.Get("/ws/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("In ws handler ...")
		var upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("Error is ", err)
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}

		fmt.Println("Connection established successfully")
		defer conn.Close()

		for {
			time.Sleep(5 * time.Second)
			err := conn.WriteMessage(websocket.TextMessage, []byte("I am healthy"))
			if err != nil {
				fmt.Println(err)
				break
			}
		}

	})

	app.R.Post("/auth/login", app.Login)
	app.R.Post("/auth/signup", app.SignUp)
	app.R.Post("/auth/logout", app.Middleware(http.HandlerFunc(app.Logout)))

	app.R.Post("/leagues/create", app.Middleware(http.HandlerFunc(app.CreateLeague)))
	app.R.Get("/leagues", app.Middleware(http.HandlerFunc(app.GetLeagues)))
	app.R.Get("/leagues/delete", app.Middleware(http.HandlerFunc(app.DeleteLeague)))
	app.R.Post("/leagues/register", app.Middleware(http.HandlerFunc(app.RegisterLeague)))
	app.R.Get("/leagues/open", http.HandlerFunc(app.OpenLeague))
	app.R.Get("/leagues/close", http.HandlerFunc(app.CloseLeague))
	app.R.Get("/leagues/start", http.HandlerFunc(app.StartLeague))
	app.R.Get("/leagues/activate", http.HandlerFunc(app.ActivateLeague))

	app.R.Post("/trade/transaction", app.Middleware(http.HandlerFunc(app.TransactPlayers)))
	app.R.Get("/trade", app.Middleware(http.HandlerFunc(app.Trade)))
	app.R.Get("/trade/points", app.Middleware(http.HandlerFunc(app.GetPointsPlayerWise)))
	app.R.Get("/trade/txns", app.Middleware(http.HandlerFunc(app.GetTxns)))

	app.R.Get("/profile", app.Middleware(http.HandlerFunc(app.GetProfile)))

	app.R.Get("/portfolio", app.Middleware(http.HandlerFunc(app.GetPortfolio)))
	app.R.Get("/portfolio/active", app.Middleware(http.HandlerFunc(app.GetActivePortfolios)))

	app.R.Get("/notifications", app.Middleware(http.HandlerFunc(app.HandleGetNotifications)))
	app.R.Post("/notifications/update/status", app.Middleware(http.HandlerFunc(app.HandleUpdateNotificationStatus)))

	app.R.Get("/notifications", app.Middleware(http.HandlerFunc(app.HandleGetNotifications)))
	app.R.Post("/notifications/update/status", app.Middleware(http.HandlerFunc(app.HandleUpdateNotificationStatus)))

	app.R.Get("/leaderboard", app.Middleware(http.HandlerFunc(app.GetLeaderboard)))

	app.R.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("I am Healthy"))
	})

}
