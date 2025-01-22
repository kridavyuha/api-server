package main

import (
	"context"
	"net/http"

	"github.com/kridavyuha/api-server/internals/auth"
)

// Middleware function
func (app *App) Middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		var token string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		if token == "" {
			sendResponse(w, httpResp{Status: http.StatusUnauthorized, IsError: true, Error: "Unauthorized"})
			return
		}

		// Validate the token and get the user ID
		userID, err := auth.New(app.KVStore, app.DB).ValidateToken(token)

		if err != nil {
			sendResponse(w, httpResp{Status: http.StatusUnauthorized, IsError: true, Error: "Unauthorized"})
			return
		}

		// Check if the token is in the list of valid tokens
		if !auth.New(app.KVStore, app.DB).CheckIfTokenIsWhiteListed(userID, token) {
			sendResponse(w, httpResp{Status: http.StatusUnauthorized, IsError: true, Error: "Unauthorized"})
			return
		}

		// Create a new context with the user ID and token

		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "token", token)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
