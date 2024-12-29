package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type UserDetalis struct {
	UserName string `json:"user_name"`
	Password string `json:"password"`
}

type UserSignupDetails struct {
	UserName string `json:"user_name"`
	MailId   string `json:"mail_id"`
	Password string `json:"password"`
}

// generateToken generates a JWT token for the given username
func (app *App) GenerateToken(userId int) (string, error) {
	// Implement token generation logic here
	// For example, using the "github.com/dgrijalva/jwt-go" package
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userId,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (app *App) ValidateToken(tokenString string) (int, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil
	})
	if err != nil {
		return 0, err
	}

	// Extract user_id from token claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := int(claims["user_id"].(float64))
		return userId, nil
	}

	return 0, errors.New("invalid token")

}

func (app *App) RevokeToken(tokenString string) error {
	// Implement token revocation logic here
	// Even if someone gets this token, it will be invalid after this
	err := app.KVStore.Set("blacklisted_"+tokenString, true)
	if err != nil {
		return err
	}

	// For example, maintaining a blacklist of revoked tokens
	return nil
}

func (app *App) Login(w http.ResponseWriter, r *http.Request) {
	// Extract the username and password from the request
	var loginDetails UserDetalis
	err := json.NewDecoder(r.Body).Decode(&loginDetails)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the user exists in the database
	var user struct {
		Password string
		UserID   int
	}
	err = app.DB.Table("users").Select("password, user_id").Where("user_name = ?", loginDetails.UserName).Scan(&user).Error
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Verify the password
	if user.Password != loginDetails.Password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Generate a JWT token
	token, err := app.GenerateToken(user.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type Response struct {
		Token string `json:"token"`
	}

	// Return the token in the response
	response := Response{Token: token}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func (app *App) SignUp(w http.ResponseWriter, r *http.Request) {
	// Extract the username and password from the request
	var userDetails UserSignupDetails
	err := json.NewDecoder(r.Body).Decode(&userDetails)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the user already exists
	var count int64
	err = app.DB.Table("users").Where("user_name = ?", userDetails.UserName).Count(&count).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "User already exists", http.StatusBadRequest)
		return
	}

	// Insert the new user into the database
	err = app.DB.Exec("INSERT INTO users (user_name, password) VALUES (?, ?)", userDetails.UserName, userDetails.Password).Error
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("User created successfully"))
}

func (app *App) Logout(w http.ResponseWriter, r *http.Request) {
	// Perform logout logic here
	// Extract the token from the request body

	token, ok := r.Context().Value("token").(string)
	if !ok {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	// Revoke the token
	err := app.RevokeToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// For example, clearing the session or revoking the token
	w.Write([]byte("Logged out successfully"))
}
