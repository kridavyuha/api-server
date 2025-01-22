package main

import (
	"net/http"

	"github.com/kridavyuha/api-server/internals/auth"
)

func (app *App) Login(w http.ResponseWriter, r *http.Request) {

	var loginDetails auth.LoginRequestBody
	err := getBody(r, &loginDetails)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "Invalid request body"})
	}

	// Check if the user exists in the database
	token, err := auth.New(app.KVStore, app.DB).Login(loginDetails)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"data": token, "message": "Logged in successfully"}})

}

func (app *App) SignUp(w http.ResponseWriter, r *http.Request) {

	var signupDetails auth.SignUpRequestBody
	err := getBody(r, &signupDetails)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: "Invalid request body"})
	}

	err = auth.New(app.KVStore, app.DB).SignUp(signupDetails)
	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusBadRequest, IsError: true, Error: err.Error()})
		return
	}

	sendResponse(w, httpResp{Status: http.StatusCreated, Data: map[string]interface{}{"message": "User created successfully"}})
}

func (app *App) Logout(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value("user_id").(int)
	token := r.Context().Value("token").(string)

	err := auth.New(app.KVStore, app.DB).RevokeToken(userId, token)

	if err != nil {
		sendResponse(w, httpResp{Status: http.StatusInternalServerError, IsError: true, Error: err.Error()})
	}

	sendResponse(w, httpResp{Status: http.StatusOK, Data: map[string]interface{}{"message": "Logged out successfully"}})
}
