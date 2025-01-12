package auth

import (
	KVStore "backend/pkg"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"gorm.io/gorm"
)

type AuthService struct {
	KV KVStore.KVStore
	DB *gorm.DB
}

func New(kv KVStore.KVStore, db *gorm.DB) *AuthService {
	return &AuthService{
		KV: kv,
		DB: db,
	}
}

// Login function
func (a *AuthService) Login(loginDetails LoginRequestBody) (string, error) {

	username := loginDetails.UserName
	password := loginDetails.Password

	var user Users
	fmt.Println("username", user.UserName, "password", user.Password, "user_id", user.UserID)

	err := a.DB.Table("users").Select("user_name, password, user_id").Where("user_name = ?", username).First(&user).Error

	if err != nil {
		return "", err
	}
	
	// Verify the password
	if user.Password != password {
		return "", errors.New("invalid credentials")
	}
	// Generate a JWT token
	token, err := a.GenerateToken(user.UserID)
	if err != nil {
		return "", err
	}
	// Insert the token into the KV store {List of tokens for a user: Multiple devices}
	err = a.KV.RPush("session_token_"+fmt.Sprintf("%d", user.UserID), token)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (a *AuthService) GenerateToken(userID int) (string, error) {
	// Implement token generation logic here
	// For example, using the "github.com/dgrijalva/jwt-go" package
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-secret-key"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (a *AuthService) ValidateToken(tokenString string) (int, error) {
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

func (a *AuthService) RevokeToken(userID int, tokenString string) error {
	// Implement token revocation logic here
	// Even if someone gets this token, it will be invalid after this
	// Get the list of tokens for this user
	tokens, err := a.KV.LRange("session_token_"+fmt.Sprintf("%d", userID), 0, -1)
	if err != nil {
		return err
	}

	// Remove the token from the list
	for _, t := range tokens {
		if t == tokenString {
			err = a.KV.LRem("session_token_"+fmt.Sprintf("%d", userID), 1, t)
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}

func (a *AuthService) CheckIfTokenIsWhiteListed(userID int, tokenString string) bool {
	// Get the list of tokens for this user
	tokens, err := a.KV.LRange("session_token_"+fmt.Sprintf("%d", userID), 0, -1)
	if err != nil {
		return false
	}

	// Check if the token is in the list
	for _, t := range tokens {
		if t == tokenString {
			return true
		}
	}

	return false
}

func (a *AuthService) Logout(userID int, tokenString string) error {
	// Revoke the token
	err := a.RevokeToken(userID, tokenString)
	if err != nil {
		return err
	}
	return nil
}

func (a *AuthService) SignUp(signUpDetails SignUpRequestBody) error {

	username := signUpDetails.UserName
	mail_id := signUpDetails.MailID
	password := signUpDetails.Password

	// Validate the mail_id if user already exists
	var count int64
	err := a.DB.Table("users").Where("mail_id = ?", mail_id).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("user already exists")
	}

	// Insert the user into the database
	err = a.DB.Table("users").Create(&Users{
		UserName:   username,
		MailID:     mail_id,
		Password:   password,
		ProfilePic: "default.jpg",
	}).Error

	if err != nil {
		return err
	}

	return nil

}
