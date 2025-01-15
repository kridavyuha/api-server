package main

import (
	"backend/internals/profile"
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
	userID := r.Context().Value("user_id").(int)

	profile, err := profile.New(app.KVStore, app.DB).GetProfile(userID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
		return
	}
	sendResponse(w, httpResp{Status: http.StatusOK, Data: profile})
}
