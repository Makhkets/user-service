package user_service

import (
	"Makhkets/internal/configs"
	user "Makhkets/internal/user/db"
	"Makhkets/pkg/errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func (s *service) GenerateAccessToken(user *user.UserDTO) (string, error) {
	cfg := configs.GetConfig()

	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.Id
	claims["username"] = user.Username
	claims["isAdmin"] = user.IsAdmin
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(cfg.Jwt.Duration)).Unix()

	// Sign the token
	accessToken, err := token.SignedString([]byte(cfg.Service.SecretKey))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *service) GenerateRefreshToken(user *user.UserDTO) (string, error) {
	cfg := configs.GetConfig()

	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.Id
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix()

	// Sign the token
	refreshToken, err := token.SignedString([]byte(cfg.Service.SecretKey))
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *service) ParseToken(tokenString string, isAccessToken bool) (jwt.MapClaims, error) {
	cfg := configs.GetConfig()

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.Service.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Validate the token
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]).(errors.NotLoggingErr)
	}

	// Extract the claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token").(errors.NotLoggingErr)
	}

	// Check if the token is an access token and has the required claims
	if isAccessToken {
		if _, ok := claims["sub"]; !ok {
			return nil, fmt.Errorf("missing subject claim").(errors.NotLoggingErr)
		}

		if _, ok := claims["username"]; !ok {
			return nil, fmt.Errorf("missing username claim").(errors.NotLoggingErr)
		}
	} else {
		// Check if the token is a refresh token and has the required claims
		if _, ok := claims["sub"]; !ok {
			return nil, fmt.Errorf("missing subject claim", true).(errors.NotLoggingErr)
		}
	}

	return claims, nil
}
