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
	userURL             = "/user/:id"
	usersURL            = "/users"
	userMeURL           = "/user/me"
	userRefreshTokenURL = "/user/refresh"
	userLoginURL        = "/user/login"
	userSessionUrl      = "user/:id/sessions"
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
		api.Handle(http.MethodGet, usersURL, h.AuthMiddleware(), h.GetUsers)
		api.Handle(http.MethodGet, userURL, h.AuthMiddleware(), h.GetUser)
		api.Handle(http.MethodGet, userSessionUrl, h.AuthMiddleware(), h.GetSessions)

		api.GET(userMeURL, h.AboutMyInfo)

		api.POST(usersURL, h.CreateUser)
		api.POST(userLoginURL, h.Login)
		api.POST(userRefreshTokenURL, h.RefreshToken)

		api.Handle(http.MethodPatch, userURL, h.AuthMiddleware(), h.PartialUpdateUser)
		api.Handle(http.MethodDelete, userURL, h.AuthMiddleware(), h.PartialUpdateUser)
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

	userDTO.IsAdmin = false
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

func (h *handler) GetUsers(c *gin.Context) {
	// TODO GET_USERS
}

func (h *handler) GetUser(c *gin.Context) {
	// TODO GetUser
}

func (h *handler) PartialUpdateUser(c *gin.Context) {
	// TODO PartialUpdateUser
}

func (h *handler) GetSessions(c *gin.Context) {
	// TODO GetSessions
}
