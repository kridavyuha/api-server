package main

import (
	"context"
	"fmt"
	"net/http"
)

// Middleware function
func (app *App) Middleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Perform middleware logic here
		// For example, logging or authentication
		authHeader := r.Header.Get("Authorization")
		var token string
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Validate the token (this is a placeholder, replace with actual validation logic)
		userID, err := app.ValidateToken(token)
		fmt.Println("UserID:", userID)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Check this token from the whitelist
		// If the token is not in the whitelist, return an error
		// This is to ensure that the token is not revoked
		tokens, err := app.KVStore.LRange("session_token_"+fmt.Sprintf("%d", userID), 0, -1)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		found := false
		for _, t := range tokens {
			if t == token {
				found = true
				break
			}
		}

		if !found {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Create a new context with the user ID

		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "token", token)

		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
