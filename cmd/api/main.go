package main

import (
	"Makhkets/internal/configs"
	"Makhkets/internal/database/postgres"
	rdb "Makhkets/internal/database/redis"
	"Makhkets/internal/user"
	user2 "Makhkets/internal/user/repository"
	user_service "Makhkets/internal/user/service"
	"Makhkets/pkg/logging"
	"fmt"
	"github.com/gin-gonic/gin"

	_ "Makhkets/docs"
)

// @title           User Service
// @version         1.0
// @description     This is user service server
// @termsOfService  http://swagger.io/terms/

// @contact.name   Makhkets
// @contact.url    https://makhkets.t.me/
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8000

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	logger := logging.GetLogger()
	logger.Info("Initialize Logger")

	logger.Info("Get Config")
	cfg := configs.GetConfig()

	logger.Info("Initialize Databases")
	pool := postgres.InitDatabase()
	rpool := rdb.InitRedis()

	r := gin.Default()
	r.Use(corsMiddleware())

	userStorage := user2.NewStorage(&logger, pool, rpool)
	userService := user_service.NewUserService(userStorage, &logger, cfg)
	userHandler := user.NewHandler(&logger, cfg, userService)
	userHandler.Register(r)

	logger.Debug("Listening this host: http://localhost:" + cfg.Listen.Port)
	if err := r.Run(fmt.Sprintf("%s:%s", cfg.Listen.Address, cfg.Listen.Port)); err != nil {
		panic(err)
	}

	return nil
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
