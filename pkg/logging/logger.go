package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

type Logger struct {
	*zap.Logger
	*zap.Config
}

var logger Logger
var once sync.Once

func GetLogger() Logger {
	once.Do(func() {
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		if _, err := os.Stat("logs"); os.IsNotExist(err) {
			os.Mkdir("logs", 0755)
		}
		config.OutputPaths = append(config.OutputPaths, "logs/logs.log")

		l, err := config.Build()
		if err != nil {
			panic(err)
		}
		logger = Logger{l, &config}
	})

	return logger
}
