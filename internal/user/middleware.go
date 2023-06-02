package user

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (h *handler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Next()

		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Undefined Authorization token"})
			return
		}
	}
}
