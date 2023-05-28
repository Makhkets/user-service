package user_service

import (
	user "Makhkets/internal/user/db"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/logging"
	"context"
	"runtime"
	"strconv"
)

type Service interface {
	CreateUser(ctx context.Context, user *user.UserDTO) (map[string]string, *errors.CustomError)
	AboutAccessToken(token string) (map[string]any, *errors.CustomError)
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

func (s *service) CreateUser(ctx context.Context, user *user.UserDTO) (map[string]string, *errors.CustomError) {
	dto, err := s.repository.Create(ctx, user)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "SQL Query Error",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	accessToken, err := s.GenerateAccessToken(dto)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	refreshToken, err := s.GenerateRefreshToken(dto)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]string{
		"access":  accessToken,
		"refresh": refreshToken,
	}, nil
}

func (s *service) AboutAccessToken(token string) (map[string]any, *errors.CustomError) {
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
				CustomErr:  "",
				Field:      strconv.Itoa(line),
				File:       file,
				Err:        err,
				IsNotWrite: true,
			}
		}
	}

	return map[string]any{
		"id":       jwt["sub"],
		"username": jwt["username"],
		"isAdmin":  jwt["isAdmin"],
	}, nil
}
