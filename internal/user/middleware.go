package user

import (
	"Makhkets/pkg/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем токен, есть ли он вообще
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Undefined Authorization token"})
			return
		}

		// Парсим токен и проверяем на валидность
		data, err := h.service.AboutAccessToken(token)
		if err != nil {
			errors.NewResponseError(h.logger, c, err)
			return
		}
		c.Set("tokenData", data)

		// Проверяем пользователя, не заблокирован ли он
		if data["isBanned"].(bool) {
			c.AbortWithStatusJSON(http.StatusBadRequest, ResponseErrors("your account is banned"))
			return
		}

		// Если флаг outsideCall установлен в true, значит AuthMiddleware был вызван из вне
		// В этом случае, мы не хотим вызывать c.Next(), поэтому просто возвращаемся из функции.
		outsideCall, exists := c.Get("outsideCall")
		if exists {
			if outsideCall.(bool) {
				return
			}
		}
		c.Next()
	}
}

func (h *handler) AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем на ввод правильного access токена
		c.Set("outsideCall", true)
		h.AuthMiddleware()(c)

		// Проверяем человека на admin
		data, _ := c.Get("tokenData")
		if data.(map[string]any)["status"].(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "This resource is not available to users",
			})
			return
		}

		c.Next()
	}
}

func (h *handler) SelfUserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем на ввод правильного access токена
		c.Set("outsideCall", true)
		h.AuthMiddleware()(c)

		data, exists := c.Get("tokenData")
		if !exists {
			c.Abort()
			return
		}

		tokenUserId := data.(map[string]any)["id"].(string)

		if tokenUserId == c.Param("id") {
			c.Next()
			return
		}

		// Если человек имеет роль Админа, то разрешаем доступ
		if data.(map[string]any)["status"].(string) == "admin" {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "This resource is not available to users",
		})
	}
}
