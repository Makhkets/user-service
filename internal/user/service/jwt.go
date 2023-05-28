package service

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/user/db"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func CreateJWT(userId int, user user.UserDTO) (string, error) {
	cfg := configs.GetConfig()
	token := jwt.New(jwt.SigningMethodES256)
	claims := token.Claims.(jwt.MapClaims)
	claims["userId"] = userId
	claims["isAdmin"] = user.IsAdmin
	claims["username"] = user.Username

	claims["exp"] = time.Now().Add(time.Duration(cfg.Jwt.Duration) * time.Hour).Unix()

	tokenString, err := token.SignedString([]byte(cfg.Service.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func verifyJWT(token string) (string, error) {
	//tokenString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
	//
	//})
	//

	return "", nil
}
