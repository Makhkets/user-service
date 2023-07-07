package configs

import (
	"Makhkets/pkg/logging"
	"Makhkets/pkg/utils"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"sync"
)

type Config struct {
	Service struct {
		IsDebug   *bool  `yaml:"is_debug"`
		SecretKey string `yaml:"secret_key"`
	} `yaml:"service"`

	Listen struct {
		Address string `yaml:"address"`
		Port    string `yaml:"port"`
	} `yaml:"listen"`

	Storage struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Database string `yaml:"database"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"storage"`

	Redis struct {
		Address  string `yaml:"address"`
		Password string `yaml:"password"`
		Db       int    `yaml:"repository"`
	} `yaml:"redis"`

	Jwt struct {
		Duration int `yaml:"duration"`
		Refresh  int `yaml:"refresh"`
	} `yaml:"jwt"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("Read Application Config")
		instance = &Config{}

		if _, err := os.Stat("config.yaml"); !os.IsExist(err) {
			if err = cleanenv.ReadConfig("config.yaml", instance); err != nil {
				help, _ := cleanenv.GetDescription(instance, nil)
				logger.Error(err.Error())
				logger.Fatal(help)
			}
		} else {
			// Находим путь до корневого каталога, где и находится config.yaml
			projectDirPath, err := utils.GetRootDirectory("config.yaml")
			projectDirPath = projectDirPath + "\\"
			if err != nil {
				panic(err)
			}

			if err = cleanenv.ReadConfig(projectDirPath+"config.yaml", instance); err != nil {
				help, _ := cleanenv.GetDescription(instance, nil)
				logger.Error(err.Error())
				logger.Fatal(help)
			}
		}
	})

	return instance
}
