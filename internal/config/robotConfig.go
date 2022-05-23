package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type RobotConfig struct {
	AppName            string `yaml:"app_name"`
	TinkoffAccessToken string `env:"TINKOFF_ACCESS_TOKEN"`
	TinkoffApiEndpoint string `yaml:"tinkoff_api_endpoint"`
}

// LoadRobotConfig Загружает конфигурацию робота из файла и переменных окружения
func LoadRobotConfig(filename string) *RobotConfig {
	var robotCfg RobotConfig
	err := cleanenv.ReadConfig(filename, &robotCfg)
	if err != nil {
		log.Fatalf("Ошибка чтения конфигурации робота %s: %v", filename, err)
	}
	return &robotCfg
}
