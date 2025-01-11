package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Users struct {
	UserID     int    `json:"user_id"`
	UserName   string `json:"user_name"`
	MailId     string `json:"mail_id"`
	ProfilePic string `json:"profile_pic"`
}

func (app *App) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch user details from Users table
	var profile Users
	err := app.DB.Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		http.Error(w, "Failed to get profile", http.StatusInternalServerError)
		return
	}

	fmt.Println("Profile:", profile)

	// Write profile to response
	json.NewEncoder(w).Encode(profile)
}
