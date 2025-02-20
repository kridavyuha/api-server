package main

import (
	"net/http"

	"github.com/kridavyuha/api-server/internals/notification"
)

func (app *App) HandleGetNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	notifications, err := notification.New(app.KVStore, app.DB).GetNotifications(userID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
	}
	sendResponse(w, httpResp{Status: http.StatusOK, Data: notifications})
}

func (app *App) HandleUpdateNotificationStatus(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int)

	err := notification.New(app.KVStore, app.DB).UpdateNotificationStatus(userID)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
	}
	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "Notification status of this user updated successfully"}})
}
