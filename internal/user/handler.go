package user

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/handlers"
	user "Makhkets/internal/user/db"
	UserStorage "Makhkets/internal/user/storage"
	"Makhkets/pkg/logging"
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"time"
)

const (
	userURL  = "/user/:id"
	usersURL = "/users"
)

type handler struct {
	logger     *logging.Logger
	cfg        *configs.Config
	validate   *validator.Validate
	repository user.Repository
}

func NewHandler(logger *logging.Logger, config *configs.Config, repository user.Repository) handlers.Handler {
	return &handler{
		logger:     logger,
		cfg:        config,
		repository: repository,
	}
}

func (h *handler) Register(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.GET(usersURL, h.GetUsers)
		api.POST(usersURL, h.CreateUser)
		api.GET(userURL, h.GetUser)
		api.PATCH(userURL, h.PartialUpdateUser)
		api.DELETE(userURL, h.PartialUpdateUser)
	}
}

func (h *handler) GetUsers(c *gin.Context) {
	// TODO GET_USERS
}

func (h *handler) CreateUser(c *gin.Context) {
	var userDTO UserStorage.UserDTO
	if err := c.BindJSON(&userDTO); err != nil {
		h.logger.Error(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	userDTO.IsAdmin = false

	h.logger.Debug(userDTO.Username)
	h.logger.Debug(userDTO.Password)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	id, err := h.repository.Create(ctx, &userDTO)
	if err != nil {
		h.logger.Error(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tokenString, err := createJWT(id, userDTO)
	if err != nil {
		h.logger.Error(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	h.logger.Debug("Successfully created user")
	c.JSON(http.StatusCreated, gin.H{
		"access": tokenString,
	})
}

func (h *handler) GetUser(c *gin.Context) {
	// TODO GET_USER
}

func (h *handler) PartialUpdateUser(c *gin.Context) {
	// TODO PartialUpdateUser
}
