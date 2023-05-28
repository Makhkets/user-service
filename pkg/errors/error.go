package errors

import (
	"Makhkets/pkg/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func NewResponseError(l *logging.Logger, c *gin.Context, customError *CustomError) {
	// Надо ли логировать эту ошибку
	if customError.IsNotWrite {
		l.Error(customError.Err.Error(), zap.String("field", customError.Field), zap.String("file", customError.File))
	} else {
		l.Info(customError.Err.Error())
	}

	// Если не присутствует кастомная ошибка, то проигнорируй ее
	if customError.CustomErr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": customError.Err.Error(),
		})
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": customError.CustomErr,
		})
	}
}
