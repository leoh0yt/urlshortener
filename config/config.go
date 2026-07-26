package config

import (
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DB   DbConfig
	Logs LogsConfig
}

type LogsConfig struct {
	Level string
}

type DbConfig struct {
	Host     string
	Port     string
	DbName   string
	User     string
	Password string
}

func LoadConfig() *Config {
	return &Config{
		DB: DbConfig{
			Host:     os.Getenv("POSTGRE_HOST"),
			Port:     os.Getenv("POSTGRE_PORT"),
			DbName:   os.Getenv("POSTGRE_DB_NAME"),
			User:     os.Getenv("POSTGRE_USER"),
			Password: os.Getenv("POSTGRE_PASSWORD"),
		},
		Logs: LogsConfig{
			Level: os.Getenv("LOGS_LEVEL"),
		},
	}
}
