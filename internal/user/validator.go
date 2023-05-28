package service

import (
	"github.com/gin-gonic/gin"
	"strings"
)

type ValidationError struct {
	Key   string `json:"key"`
	Error string `json:"error"`
	Tag   string `json:"tag"`
}

func ResponseErrors(errMsg string) map[string]interface{} {
	responseError := make(gin.H)

	// Разбиваем строку на отдельные сообщения об ошибках
	errorMessages := strings.Split(errMsg, "\n")

	for _, msg := range errorMessages {
		// Проверяем, что строка не пустая
		if msg == "" {
			continue
		}

		key := strings.Split(strings.Split(msg, "Key: '")[1], "'")[0]
		value := strings.Split(msg, "Error:")[1]

		responseError[key] = value
	}

	return responseError
}
