package user

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/handlers"
	user "Makhkets/internal/user/db"
	user_service "Makhkets/internal/user/service"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/logging"
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	userURL   = "/user/:id"
	usersURL  = "/users"
	userMeURL = "/user/me"
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
		api.GET(usersURL, h.GetUsers)
		api.GET(userMeURL, h.AboutMyInfo)
		api.POST(usersURL, h.CreateUser)
		api.GET(userURL, h.GetUser)
		api.PATCH(userURL, h.PartialUpdateUser)
		api.DELETE(userURL, h.PartialUpdateUser)
	}
}

func (h *handler) GetUsers(c *gin.Context) {
	// TODO GET_USERS
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

	tokenData, err := h.service.CreateUser(ctx, &userDTO)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Debug("Successfully created user")
	c.JSON(http.StatusCreated, tokenData)
}

func (h *handler) GetUser(c *gin.Context) {
	// TODO GetUser
}

func (h *handler) PartialUpdateUser(c *gin.Context) {
	// TODO PartialUpdateUser
}
