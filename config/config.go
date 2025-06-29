package config

import (
	"github.com/joho/godotenv"
	"os"
)

type AppConfig struct {
	Mode             string
	LogPath          string
	LogMaxAge        string
	ConsoleLogFormat string
}

func init() {
	_ = godotenv.Load("config/.env")
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetAppConfig() AppConfig {
	return AppConfig{
		Mode:             GetEnv("MODE", "development"),
		LogPath:          GetEnv("LOG_PATH", "./mwp.log"),
		LogMaxAge:        GetEnv("LOG_MAX_AGE", "30"),
		ConsoleLogFormat: GetEnv("CONSOLE_LOG_FORMAT", "plain"),
	}
}
