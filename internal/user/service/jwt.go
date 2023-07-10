package user_service

import (
	"Makhkets/internal/configs"
	user "Makhkets/internal/user/repository"
	user2 "Makhkets/internal/user/repository"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/utils"
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"runtime"
	"strconv"
	"time"
)

func (s *service) GenerateAccessToken(user *user.UserDTO) (string, error) {
	// Проверяем id юзера
	if id, err := strconv.Atoi(user.Id); err != nil {
		return "", err
	} else {
		if id <= 0 {
			return "", fmt.Errorf("zero ID")
		}
	}

	// Проверяем роль
	if user.Status == "" {
		user.Status = user2.StatusUser
	}

	if !utils.ContainsStringInArray(user.Status, user2.Roles) {
		return "", fmt.Errorf("unknown user role")
	}

	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.Id
	claims["username"] = user.Username
	claims["status"] = user.Status
	claims["isBanned"] = user.IsBanned
	claims["exp"] = time.Now().Add(time.Minute * time.Duration(s.config.Jwt.Duration)).Unix()

	// Подписываем токен
	accessToken, err := token.SignedString([]byte(s.config.Service.SecretKey))
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func (s *service) GenerateRefreshToken(user *user.UserDTO) (string, int64, error) {
	// Проверяем id юзера
	if id, err := strconv.Atoi(user.Id); err != nil {
		return "", 0, err
	} else {
		if id <= 0 {
			return "", 0, fmt.Errorf("zero ID")
		}
	}

	// Create the token
	token := jwt.New(jwt.SigningMethodHS256)
	exp := time.Now().Add(time.Minute * time.Duration(s.config.Jwt.Refresh)).Unix()

	// Set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = user.Id
	claims["exp"] = exp

	// Sign the token
	refreshToken, err := token.SignedString([]byte(s.config.Service.SecretKey))
	if err != nil {
		return "", 0, err
	}

	return refreshToken, exp, nil
}

func (s *service) CreateTokenPair(dto *user.UserDTO, req *http.Request) (map[string]string, int64, *errors.CustomError) {
	// Создаем access токен
	accessToken, err := s.GenerateAccessToken(dto)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, 0, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Создаем refresh токен
	refreshToken, exp, err := s.GenerateRefreshToken(dto)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, 0, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Обновляем refresh в Redis'e
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err2 := s.repository.SaveRefreshSession(ctx, &user.RefreshSession{
		RefreshToken: refreshToken,
		UserId:       dto.Id,
		Ua:           req.UserAgent(),
		Ip:           req.RemoteAddr,
		Fingerprint:  utils.GetFingerprint(req.Header),
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	})

	if err2 != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, 0, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err2,
		}
	}

	return map[string]string{
		"access":  accessToken,
		"refresh": refreshToken,
	}, exp, nil
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
			return nil, fmt.Errorf("missing subject claim").(errors.NotLoggingErr)
		}
	}

	return claims, nil
}
