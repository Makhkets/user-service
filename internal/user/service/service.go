package user_service

import (
	user "Makhkets/internal/user/repository"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/logging"
	"Makhkets/pkg/utils"
	"context"
	"github.com/gin-gonic/gin"
	"runtime"
	"strconv"
	"time"
)

type Service interface {
	CreateUser(ctx context.Context, c *gin.Context, u *user.UserDTO) (map[string]string, *errors.CustomError)
	LoginUser(c *gin.Context, username, password string) (map[string]string, *errors.CustomError)

	AboutAccessToken(token string) (map[string]any, *errors.CustomError)
	RefreshAccessToken(c *gin.Context, refreshToken string) (map[string]string, *errors.CustomError)
}

type service struct {
	repository user.Repository
	logger     *logging.Logger
}

func NewUserService(r user.Repository, l *logging.Logger) Service {
	return &service{
		repository: r,
		logger:     l,
	}
}

func (s *service) LoginUser(c *gin.Context, username, password string) (map[string]string, *errors.CustomError) {
	// Находим юзера по login и password
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	dto, err := s.repository.FindLoginUser(ctx, username, password)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "Invalid data",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Удаляем из redis токен refresh
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	if err := s.repository.DeleteRefreshSession(ctx, utils.GetFingerprint(c.Request.Header)); err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Генерируем access и refresh токен, попутно занеся в redis
	tokenPair, _, error := s.CreateTokenPair(&user.UserDTO{
		Id:       strconv.Itoa(dto.Id),
		Username: dto.Username,
		Password: dto.PasswordHash,
		IsAdmin:  dto.IsAdmin,
		IsBanned: dto.IsBanned,
	}, c)
	if error != nil {
		return nil, error
	}
	return tokenPair, nil
}

func (s *service) CreateUser(ctx context.Context, c *gin.Context, u *user.UserDTO) (map[string]string, *errors.CustomError) {
	// Создаем пользователя
	fingerprint := utils.GetFingerprint(c.Request.Header)
	dto, err := s.repository.
		CreateUser(ctx, u)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "SQL Query Error",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	tokenPair, exp, error := s.CreateTokenPair(dto, c)
	if error != nil {
		return nil, error
	}

	// Заносим Refresh Token в Redis хранилище
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	if err := s.repository.SaveRefreshSession(ctx, &user.RefreshSession{
		RefreshToken: tokenPair["refresh"],
		UserId:       dto.Id,
		Ua:           c.Request.UserAgent(),
		Ip:           c.ClientIP(),
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}); err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return tokenPair, nil
}

func (s *service) RefreshAccessToken(c *gin.Context, refreshToken string) (map[string]string, *errors.CustomError) {
	// Проверяем на валидность refresh token и вытаскиваем id юзера
	fingerprint := utils.GetFingerprint(c.Request.Header)
	jwt, err := s.ParseToken(refreshToken, false)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		switch err.(type) {
		case error:
			return nil, &errors.CustomError{
				CustomErr: "",
				Field:     strconv.Itoa(line),
				File:      file,
				Err:       err,
			}
		case errors.NotLoggingErr:
			return nil, &errors.CustomError{
				CustomErr:       "",
				Field:           strconv.Itoa(line),
				File:            file,
				Err:             err,
				IsNotWriteError: true,
			}
		}
	}

	// Проверяем fingerprint, user-agent и т.п юзера
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	refreshSession, err := s.repository.GetRefreshSession(ctx, utils.GetFingerprint(c.Request.Header))
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr:         "",
			Field:             strconv.Itoa(line),
			File:              file,
			Err:               err,
			IsNotWriteMessage: true,
		}
	}

	// Проверяем FingerPrint, если они не равны, то возвращаем ошибку
	if refreshSession.Fingerprint != fingerprint || refreshSession.RefreshToken != refreshToken {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "Fingerprint or refresh token invalid",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       nil,
		}
	}

	// Находим юзера для создания пары ключей
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	dto, err := s.repository.FindOne(ctx, jwt["sub"].(string))
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "Token is invalid",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Обновляем access и refresh токен
	tokenPair, _, error := s.CreateTokenPair(&user.UserDTO{
		Id:       strconv.Itoa(dto.Id),
		Username: dto.Username,
		Password: dto.PasswordHash,
		IsAdmin:  dto.IsAdmin,
		IsBanned: dto.IsBanned,
	}, c)
	if error != nil {
		return nil, error
	}

	return tokenPair, nil
}

func (s *service) AboutAccessToken(token string) (map[string]any, *errors.CustomError) {
	// Проверяем токен на валидность
	jwt, err := s.ParseToken(token, true)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		switch err.(type) {
		case error:
			return nil, &errors.CustomError{
				CustomErr: "",
				Field:     strconv.Itoa(line),
				File:      file,
				Err:       err,
			}
		case errors.NotLoggingErr:
			return nil, &errors.CustomError{
				CustomErr:       "",
				Field:           strconv.Itoa(line),
				File:            file,
				Err:             err,
				IsNotWriteError: true,
			}
		}
	}

	return map[string]any{
		"id":       jwt["sub"],
		"username": jwt["username"],
		"isAdmin":  jwt["isAdmin"],
		"isBanned": jwt["isBanned"],
	}, nil

}
