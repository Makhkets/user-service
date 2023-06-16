package user

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/handlers"
	user "Makhkets/internal/user/repository"
	user_service "Makhkets/internal/user/service"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/logging"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	usersURL            = "/users"
	userMeURL           = "/user/me"
	userRefreshTokenURL = "/user/refresh"
	userLoginURL        = "/user/login"

	userURL        = "/user/:id"
	userSessionUrl = "/user/:id/sessions"

	userUpdateUsernameURL = "/user/:id/change_username"
	userUpdatePasswordURL = "/user/:id/change_password"
	userChangeStatus      = "/user/:id/change_status"
	userChangePermission  = "/user/:id/change_permission"
)

type handler struct {
	logger  *logging.Logger
	service user_service.Service
	cfg     *configs.Config
}

func NewHandler(l *logging.Logger, c *configs.Config, s user_service.Service) handlers.Handler {
	return &handler{
		service: s,
		logger:  l,
		cfg:     c,
	}
}

func (h *handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.Handle(http.MethodGet, usersURL, h.AuthMiddleware(), h.GetUsers) // TODO
		api.Handle(http.MethodGet, userURL, h.AuthMiddleware(), h.GetUser)   // TODO

		api.Handle(http.MethodDelete, userURL, h.SelfUserMiddleware(), h.DeleteUser)
		api.Handle(http.MethodGet, userSessionUrl, h.SelfUserMiddleware(), h.GetSessions) // TODO

		api.Handle(http.MethodPost, userUpdateUsernameURL, h.SelfUserMiddleware(), h.UsernameUpdate)
		api.Handle(http.MethodPost, userUpdatePasswordURL, h.SelfUserMiddleware(), h.PasswordUpdate)

		// TODO
		api.Handle(http.MethodPost, userChangeStatus, h.AdminMiddleware(), h.StatusChange)
		api.Handle(http.MethodPost, userChangePermission, h.AdminMiddleware(), h.PartialUpdateUser)

		// Тест админского Middleware
		api.Handle(http.MethodGet, "/user/test", h.AdminMiddleware())

		api.GET(userMeURL, h.AboutMyInfo)
		api.POST(usersURL, h.CreateUser)
		api.POST(userLoginURL, h.Login)
		api.POST(userRefreshTokenURL, h.RefreshToken)

	}
}

func (h *handler) AboutMyInfo(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token != "" {
		responseData, err := h.service.AboutAccessToken(token)
		if err != nil {
			errors.NewResponseError(h.logger, c, err)
			return
		}
		h.logger.Info("Received data from token")
		c.JSON(http.StatusAccepted, responseData)
	} else {
		h.logger.Info("Did not find the token")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "None \"Authorization\" key in your headers",
		})
	}
}

func (h *handler) RefreshToken(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	var data struct {
		Refresh string
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.RefreshAccessToken(c, data.Refresh)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Successfully refreshed access token")
	c.JSON(http.StatusAccepted, response)
}

func (h *handler) CreateUser(c *gin.Context) {
	var userDTO user.UserDTO
	c.Header("Content-Type", "application/json; charset=utf-8")
	if err := c.BindJSON(&userDTO); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	userDTO.Status = "user"
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tokenData, err := h.service.CreateUser(ctx, c, &userDTO)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Debug("Successfully created user")
	c.JSON(http.StatusCreated, tokenData)
}

func (h *handler) Login(c *gin.Context) {
	var data struct {
		Username string `binding:"required,min=4"`
		Password string `binding:"required,min=8"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.LoginUser(c, data.Username, data.Password)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Successfully Login in Account")
	c.JSON(http.StatusCreated, response)
}

func (h *handler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	response, err := h.service.DeleteAccount(id)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Deleted User with id: " + id)
	c.JSON(http.StatusOK, response)
}

func (h *handler) PartialUpdateUser(c *gin.Context) {
	// Извлекаем идентификатор пользователя из пути запроса
	// Привязываем JSON-запрос к структуре пользователя
	id := c.Param("id")
	var u user.User
	if err := c.BindJSON(&u); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	// Обновляем данные пользователя
	response, err := h.service.UpdateAccount(id, &u)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusAccepted, response)
}

func (h *handler) GetUsers(c *gin.Context) {
	// TODO GET_USERS
}

func (h *handler) GetUser(c *gin.Context) {
	// TODO GetUser
}

func (h *handler) GetSessions(c *gin.Context) {
	// TODO GetSessions
}

func (h *handler) UsernameUpdate(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	var data struct {
		Username string `json:"username" binding:"required,min=4"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.UsernameUpdate(data.Username, accessToken)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Username " + response["old"] + " updated his nickname to " + response["new"])
	c.JSON(http.StatusAccepted, response)
}

func (h *handler) PasswordUpdate(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	var data struct {
		OldPassword string `json:"old_password" binding:"required,min=8"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.PasswordUpdate(data.OldPassword, data.NewPassword, accessToken)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info(data.OldPassword + " password has been changed to " + data.NewPassword)
	c.JSON(http.StatusAccepted, response)
}

func (h *handler) StatusChange(c *gin.Context) {
	id := c.Param("id")
	var data struct {
		Status string `json:"status" binding:"required,min=4"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.StatusUpdate(id, data.Status)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Changed status to: " + data.Status) // todo
	c.JSON(http.StatusAccepted, response)
}
