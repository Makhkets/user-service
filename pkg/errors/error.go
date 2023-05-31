package errors

import (
	"Makhkets/pkg/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func NewResponseError(l *logging.Logger, c *gin.Context, customError *CustomError) {
	// Надо ли проигнорировать логирование
	if !customError.IsNotWriteMessage {
		// Надо ли логировать эту ошибку
		if customError.IsNotWriteError {
			l.Error(customError.Err.Error(), zap.String("field", customError.Field), zap.String("file", customError.File))
		} else {
			if customError.Err != nil {
				l.Info(customError.Err.Error(), zap.String("field", customError.Field), zap.String("file", customError.File))
			}
		}
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
