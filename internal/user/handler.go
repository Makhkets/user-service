package user

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/handlers"
	user_repo "Makhkets/internal/user/repository"
	user_service "Makhkets/internal/user/service"
	"Makhkets/pkg/errors"
	"Makhkets/pkg/logging"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
	"time"

	_ "Makhkets/docs"
)

const (
	adminURL = "/users/admin"

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
		api.Handle(http.MethodPost, adminURL, h.GetAdminPermission)

		api.Handle(http.MethodGet, usersURL, h.AuthMiddleware(), h.GetUsers) // Под себя реализовать надо с offset'ами
		api.Handle(http.MethodGet, userURL, h.AuthMiddleware(), h.GetUser)
		api.GET(userMeURL, h.AuthMiddleware(), h.AboutMyInfo)

		api.Handle(http.MethodDelete, userURL, h.SelfUserMiddleware(), h.DeleteUser)
		api.Handle(http.MethodGet, userSessionUrl, h.SelfUserMiddleware(), h.GetSessions)

		api.Handle(http.MethodPost, userUpdateUsernameURL, h.SelfUserMiddleware(), h.UsernameUpdate)
		api.Handle(http.MethodPost, userUpdatePasswordURL, h.SelfUserMiddleware(), h.PasswordUpdate)
		api.Handle(http.MethodPost, userChangeStatus, h.AdminMiddleware(), h.StatusChange)
		api.Handle(http.MethodPost, userChangePermission, h.AdminMiddleware(), h.PermissionChange)

		// Тест админского Middleware
		api.Handle(http.MethodGet, "/user/test", h.AdminMiddleware())

		api.POST(usersURL, h.CreateUser)
		api.POST(userLoginURL, h.Login)
		api.POST(userRefreshTokenURL, h.RefreshToken)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// GetAdminPermission godoc
// @Summary Getting admin permissions
// @Description  Getting admin permissions
// @Tags         user
// @Accept       json
// @Produce      json
// @Param 		 input body 	 user_repo.GenerateTokenForm true "account info"
// @Success      200   {object}  user_repo.ResponseAccessToken
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/users/admin [post]
func (h *handler) GetAdminPermission(c *gin.Context) {
	if *h.cfg.Service.IsDebug {
		var data struct {
			Username string `binding:"required,min=4"`
			Password string `binding:"required,min=8"`
		}

		if err := c.BindJSON(&data); err != nil {
			h.logger.Error("dsadasdsa")
			c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
			return
		}

		response, err := h.service.LoginUser(c, data.Username, data.Password)
		if err != nil {
			errors.NewResponseError(h.logger, c, err)
			return
		}

		refreshToken := response["refresh"]
		accessToken, err := h.service.GetAdminTokens(c, refreshToken)
		if err != nil {
			errors.NewResponseError(h.logger, c, err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"access": accessToken})
	} else {
		c.AbortWithStatusJSON(400, gin.H{"error": "is not debug mode"})
		return
	}
}

// AboutMyInfo godoc
// @Summary About My Info
// @Tags         jwt
// @Description  Get My Info
// @Security     ApiKeyAuth
// @Accept       json
// @Produce      json
// @Success      200   {object}  user_repo.AboutAccessToken
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/me [get]
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

// RefreshToken godoc
// @Summary Refreshing token pair
// @Security     ApiKeyAuth
// @Description  Refreshing pair tokens
// @Tags         jwt
// @Accept       json
// @Produce      json
// @Param 		 input body 	 user_repo.RefreshTokenForm true "REFRESH TOKEN"
// @Success      200   {object}  user_repo.ResponseAccessToken
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/refresh [post]
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

// CreateUser godoc
// @Summary Creating User Handler
// @Tags         user
// @Accept       json
// @Produce      json
// @Param 		 input body 	 user_repo.UserDTOForm true "User Data"
// @Success      200   {object}  user_repo.CreateUserResponseForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/users [post]
func (h *handler) CreateUser(c *gin.Context) {
	var userDTO user_repo.UserDTO
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

// Login godoc
// @Summary Login Handler
// @Tags         user
// @Accept       json
// @Produce      json
// @Param 		 input body 	 user_repo.GenerateTokenForm true "REFRESH TOKEN"
// @Success      200   {object}  user_repo.ResponseAccessToken
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/login [post]
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

// DeleteUser godoc
// @Summary Deleting user handler
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Success      200   {object}  user_repo.MessageResponseForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id} [delete]
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

//func (h *handler) PartialUpdateUser(c *gin.Context) {
//	// Извлекаем идентификатор пользователя из пути запроса
//	// Привязываем JSON-запрос к структуре пользователя
//	id := c.Param("id")
//	var u user.User
//	if err := c.BindJSON(&u); err != nil {
//		c.AbortWithStatusJSON(http.StatusBadRequest, ResponseErrors(err.Error()))
//		return
//	}
//
//	// Обновляем данные пользователя
//	response, err := h.service.UpdateAccount(id, &u)
//	if err != nil {
//		errors.NewResponseError(h.logger, c, err)
//		return
//	}
//
//	// Возвращаем успешный ответ
//	c.JSON(http.StatusAccepted, response)
//}

// GetUser godoc
// @Summary Gettings user
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Success      200   {object}  user_repo.GetUserForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id} [get]
func (h *handler) GetUser(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.GetUser(id)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSessions godoc
// @Summary Getting user sessions
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Success      200   {object}  user_repo.UserSessionsForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id}/sessions [get]
func (h *handler) GetSessions(c *gin.Context) {
	id := c.Param("id")

	response, err := h.service.GetUserSessions(id)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// UsernameUpdate godoc
// @Summary Username Update
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Param 		 input body 	 user_repo.UsernameForm true "Data"
// @Success      200   {object}  user_repo.UsernameResponseForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id}/change_username [post]
func (h *handler) UsernameUpdate(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")

	type Data struct {
		Username string `json:"username" binding:"required,min=4"`
	}

	var d Data

	if err := c.BindJSON(&d); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.UsernameUpdate(d.Username, accessToken)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Username " + response["old"] + " updated his nickname to " + response["new"])
	c.JSON(http.StatusAccepted, response)
}

// PasswordUpdate godoc
// @Summary Password Update
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Param 		 input body 	 user_repo.PasswordForm true "Data"
// @Success      200   {object}  user_repo.PasswordResponseForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id}/change_password [post]
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

// StatusChange godoc
// @Summary User Status Update
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Param 		 input body 	 user_repo.StatusForm true "Data"
// @Success      200   {object}  user_repo.StatusForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id}/change_status [post]
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

// PermissionChange godoc
// @Summary User Permission Change
// @Security     ApiKeyAuth
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "ID"
// @Param 		 input body 	 user_repo.PermissionForm true "Data"
// @Success      200   {object}  user_repo.PermissionForm
// @Failure      400   {object}  user_repo.ResponseError
// @Router       /api/user/{id}/change_permission [post]
func (h *handler) PermissionChange(c *gin.Context) {
	id := c.Param("id")
	var data struct {
		Permission *bool `json:"permission" binding:"required"`
	}

	if err := c.BindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, ResponseErrors(err.Error()))
		return
	}

	response, err := h.service.PermissionUpdate(id, *data.Permission)
	if err != nil {
		errors.NewResponseError(h.logger, c, err)
		return
	}

	h.logger.Info("Changed permission to: " + fmt.Sprintf("%v", *data.Permission))
	c.JSON(http.StatusAccepted, response)
}

func (h *handler) GetUsers(c *gin.Context) {
	// Реализовать под себя, с offset'амм
	panic("implement me")
}
