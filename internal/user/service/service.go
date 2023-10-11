package user_service

import (
	"Makhkets/internal/configs"
	user "Makhkets/internal/user/repository"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/logging"
	"Makhkets/pkg/utils"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Service interface {
	CreateUser(ctx context.Context, c *gin.Context, u *user.UserDTO) (map[string]string, *errors.CustomError)
	LoginUser(c *gin.Context, username, password string) (map[string]string, *errors.CustomError)

	DeleteAccount(id string) (map[string]any, *errors.CustomError)
	UpdateAccount(id string, u *user.User) (map[string]any, *errors.CustomError)

	GetUser(id string) (map[string]any, *errors.CustomError)
	GetUserSessions(id string) (map[string][]map[string]any, *errors.CustomError)

	AboutAccessToken(token string) (map[string]any, *errors.CustomError)
	RefreshAccessToken(c *gin.Context, refreshToken string) (map[string]string, *errors.CustomError)

	UsernameUpdate(username, accessToken string) (map[string]string, *errors.CustomError)
	PasswordUpdate(old_password, new_password, accessToken string) (map[string]any, *errors.CustomError)
	StatusUpdate(accessToken, status string) (map[string]string, *errors.CustomError)
	PermissionUpdate(id string, permission bool) (map[string]bool, *errors.CustomError)

	GetAdminTokens(c *gin.Context, refreshToken string) (string, *errors.CustomError)
}

type service struct {
	repository user.Repository
	logger     *logging.Logger
	config     *configs.Config
}

func NewUserService(r user.Repository, l *logging.Logger, cfg *configs.Config) Service {
	return &service{
		repository: r,
		logger:     l,
		config:     cfg,
	}
}

func (s *service) LoginUser(c *gin.Context, username, password string) (map[string]string, *errors.CustomError) {
	// Находим юзера по login и password
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	dto, err := s.repository.FindLoginUser(ctx, username, utils.PasswordToHash(password, s.config.Service.SecretKey))
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "Invalid data",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Удаляем из redis токен refresh
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	if err := s.repository.DeleteRefreshSession(ctx, utils.GetFingerprint(c.Request.Header), strconv.Itoa(dto.Id)); err != nil {
		_, file, line, _ := runtime.Caller(0)
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
		Password: utils.PasswordToHash(dto.PasswordHash, s.config.Service.SecretKey),
		Status:   dto.Status,
		IsBanned: dto.IsBanned,
	}, c.Request)
	if error != nil {
		return nil, error
	}
	return tokenPair, nil
}

func (s *service) CreateUser(ctx context.Context, c *gin.Context, u *user.UserDTO) (map[string]string, *errors.CustomError) {
	// Создаем пользователя
	fingerprint := utils.GetFingerprint(c.Request.Header)
	u.Password = utils.PasswordToHash(u.Password, s.config.Service.SecretKey)
	dto, err := s.repository.CreateUser(ctx, u)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "SQL Query Error",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	dto.Password = utils.PasswordToHash(dto.Password, s.config.Service.SecretKey)
	tokenPair, exp, error := s.CreateTokenPair(dto, c.Request)
	if error != nil {
		return nil, error
	}

	// Заносим Refresh Token в Redis хранилище
	ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.repository.SaveRefreshSession(ctx, &user.RefreshSession{
		RefreshToken: tokenPair["refresh"],
		UserId:       dto.Id,
		Ua:           c.Request.UserAgent(),
		Ip:           c.ClientIP(),
		Fingerprint:  fingerprint,
		ExpiresIn:    time.Duration(exp),
		CreatedAt:    time.Now(),
	}); err != nil {
		_, file, line, _ := runtime.Caller(0)
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
	cfg := configs.GetConfig()
	fingerprint := utils.GetFingerprint(c.Request.Header)
	jwt, err := s.ParseToken(refreshToken, false)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
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
		_, file, line, _ := runtime.Caller(0)
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
		_, file, line, _ := runtime.Caller(0)
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
		_, file, line, _ := runtime.Caller(0)
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
		Password: utils.PasswordToHash(dto.PasswordHash, cfg.Service.SecretKey),
		Status:   dto.Status,
		IsBanned: dto.IsBanned,
	}, c.Request)
	if error != nil {
		return nil, error
	}

	return tokenPair, nil
}

func (s *service) AboutAccessToken(token string) (map[string]any, *errors.CustomError) {
	// Проверяем токен на валидность
	jwt, err := s.ParseToken(token, true)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
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
		"status":   jwt["status"],
		"isBanned": jwt["isBanned"],
	}, nil

}

func (s *service) DeleteAccount(id string) (map[string]any, *errors.CustomError) {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	if err := s.repository.Delete(ctx, id); err != nil {
		_, file, line, _ := runtime.Caller(0)
		custErr := ""
		if !strings.Contains(err.Error(), "user with ID") {
			custErr = "SQL Query Error"
		}
		return nil, &errors.CustomError{
			CustomErr: custErr,
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]any{
		"message": fmt.Sprintf("user with id %s deleted", id),
	}, nil
}

func (s *service) UpdateAccount(id string, u *user.User) (map[string]any, *errors.CustomError) {
	// Определяем какие поля были запрошены для изменения
	// Если получили поле, которое может изменить только админ, возвращаем ошибку, еслиэ то не админ
	fields := utils.CheckEmptyFields(*u)
	if len(fields) <= 0 {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "Empty data",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       nil,
		}
	}

	// Проверяем на запрещенные поля в отправленной форме
	for _, field := range fields {
		if user.BlackListCheck(field) {
			_, file, line, _ := runtime.Caller(0)
			return nil, &errors.CustomError{
				CustomErr: "",
				Field:     strconv.Itoa(line),
				File:      file,
				Err:       nil,
			}
		}
	}

	return nil, nil
}

func (s *service) UsernameUpdate(username, accessToken string) (map[string]string, *errors.CustomError) {
	// Парсим токен
	data, err := s.ParseToken(accessToken, true)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	if strings.ToLower(data["username"].(string)) == strings.ToLower(username) {
		return map[string]string{
			"new": username,
			"old": data["username"].(string),
		}, nil
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err = s.repository.UpdateUsername(ctx, data["sub"].(string), username); err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]string{
		"new": username,
		"old": data["username"].(string),
	}, nil
}

func (s *service) PasswordUpdate(oldPassword, newPassword, accessToken string) (map[string]any, *errors.CustomError) {
	// Парсим токен
	data, err := s.ParseToken(accessToken, true)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Переводим пароль, в хэш пароль
	oldPasswordHash := utils.PasswordToHash(oldPassword, s.config.Service.SecretKey)
	newPasswordHash := utils.PasswordToHash(newPassword, s.config.Service.SecretKey)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err = s.repository.UpdatePassword(ctx, data["sub"].(string), oldPasswordHash, newPasswordHash); err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]any{
		"new_password": newPassword,
		"old_password": oldPassword,
	}, nil
}

func (s *service) StatusUpdate(id, status string) (map[string]string, *errors.CustomError) {
	// Проверяем какой статус к нам пришел, есть ли он в нашем списке статусов
	if !utils.ContainsStringInArray(status, user.Roles) {
		return map[string]string{"error": "status undefined"}, nil
	}

	// Изменяем роль юзера
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	if err := s.repository.ChangeStatus(ctx, id, status); err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "SQL Query Error",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]string{
		"status": status,
	}, nil
}

func (s *service) PermissionUpdate(id string, permission bool) (map[string]bool, *errors.CustomError) {
	// Изменяем роль юзера
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	if err := s.repository.ChangePermission(ctx, id, permission); err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "SQL Query Error",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]bool{
		"permission": permission,
	}, nil
}

func (s *service) GetUser(id string) (map[string]any, *errors.CustomError) {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	user, err := s.repository.GetUser(ctx, id)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "SQL Query Error",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	return map[string]any{
		"id":        user.Id,
		"username":  user.Username,
		"status":    user.Status,
		"is_banned": user.IsBanned,
	}, nil
}

func (s *service) GetUserSessions(id string) (map[string][]map[string]any, *errors.CustomError) {
	ctx, _ := context.WithTimeout(context.Background(), 3*time.Second)
	sessions, err := s.repository.GetRefreshSessionsByUserId(ctx, id)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return nil, &errors.CustomError{
			CustomErr: "",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	data := map[string][]map[string]any{
		id: []map[string]any{},
	}

	for _, session := range sessions {
		data[id] = append(data[id], map[string]any{
			"id":           session.UserId,
			"ip":           session.Ip,
			"user-agent":   session.Ua,
			"fingerprint":  session.Fingerprint,
			"refreshToken": session.RefreshToken,
			"createdAt":    session.CreatedAt,
			"ExpiresIn":    session.ExpiresIn,
		})
	}

	return data, nil
}

func (s *service) GetAdminTokens(c *gin.Context, refreshToken string) (string, *errors.CustomError) {
	// Проверяем на валидность refresh token и вытаскиваем id юзера
	cfg := configs.GetConfig()
	fingerprint := utils.GetFingerprint(c.Request.Header)
	jwt, err := s.ParseToken(refreshToken, false)
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		switch err.(type) {
		case error:
			return "", &errors.CustomError{
				CustomErr: "",
				Field:     strconv.Itoa(line),
				File:      file,
				Err:       err,
			}
		case errors.NotLoggingErr:
			return "", &errors.CustomError{
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
		_, file, line, _ := runtime.Caller(0)
		return "", &errors.CustomError{
			CustomErr:         "",
			Field:             strconv.Itoa(line),
			File:              file,
			Err:               err,
			IsNotWriteMessage: true,
		}
	}

	// Проверяем FingerPrint, если они не равны, то возвращаем ошибку
	if refreshSession.Fingerprint != fingerprint || refreshSession.RefreshToken != refreshToken {
		_, file, line, _ := runtime.Caller(0)
		return "", &errors.CustomError{
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
		_, file, line, _ := runtime.Caller(0)
		return "", &errors.CustomError{
			CustomErr: "Token is invalid",
			Field:     strconv.Itoa(line),
			File:      file,
			Err:       err,
		}
	}

	// Обновляем access и refresh токен
	accessToken, err := s.GenerateAccessToken(&user.UserDTO{
		Id:       strconv.Itoa(dto.Id),
		Username: dto.Username,
		Password: utils.PasswordToHash(dto.PasswordHash, cfg.Service.SecretKey),
		Status:   dto.Status,
		IsBanned: dto.IsBanned,
	})
	if err != nil {
		_, file, line, _ := runtime.Caller(0)
		return "", &errors.CustomError{
			CustomErr:         err.Error(),
			Field:             strconv.Itoa(line),
			File:              file,
			Err:               err,
			IsNotWriteMessage: true,
		}
	}

	return accessToken, nil
}
