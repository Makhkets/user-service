package logging

import (
	"Makhkets/pkg/utils"
	"fmt"
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

		// Находим путь до корневого каталога, где и находится config.yaml
		projectDirPath, err := utils.GetRootDirectory("config.yaml")

		if _, err := os.Stat(projectDirPath + "\\logs"); os.IsNotExist(err) {
			os.Mkdir(projectDirPath+"\\logs", 0755)
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

func (l *Logger) PrintInterface(v ...interface{}) {
	l.Info(fmt.Sprintf("%v", v))
}
