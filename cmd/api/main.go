package main

import (
	"Makhkets/database/postgres"
	"Makhkets/internal/configs"
	"Makhkets/internal/user"
	user2 "Makhkets/internal/user/db"
	user_service "Makhkets/internal/user/service"
	"Makhkets/pkg/logging"
	"fmt"
	"github.com/gin-gonic/gin"
)

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

	logger.Info("Initialize Database")
	pool := postgres.InitDatabase()

	r := gin.Default()

	userStorage := user2.NewStorage(&logger, pool)
	userService := user_service.NewUserService(userStorage, &logger)
	userHandler := user.NewHandler(&logger, cfg, userService)
	userHandler.Register(r)

	logger.Debug("Listening this host: http://localhost:" + cfg.Listen.Port)
	if err := r.Run(fmt.Sprintf("%s:%s", cfg.Listen.Address, cfg.Listen.Port)); err != nil {
		panic(err)
	}

	return nil
}
