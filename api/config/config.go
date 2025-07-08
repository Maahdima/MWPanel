package config

import (
	"github.com/joho/godotenv"
	"os"
)

type AppConfig struct {
	Mode             string
	Host             string
	Port             string
	ConsoleLogFormat string
	PublicDir        string
	PeerFilesDir     string
}

type DBConfig struct {
	Host         string
	Port         string
	Username     string
	Password     string
	Database     string
	Dialect      string
	MigrationDir string
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
		Host:             GetEnv("SERVER_HOST", "127.0.0.1"),
		Port:             GetEnv("SERVER_PORT", "3000"),
		ConsoleLogFormat: GetEnv("CONSOLE_LOG_FORMAT", "plain"),
		PublicDir:        GetEnv("PUBLIC_DIR", "./public/"),
		PeerFilesDir:     GetEnv("PEER_FILES_DIR", "./peer-files/"),
	}
}

func GetDBConfig() DBConfig {
	return DBConfig{
		Host:         GetEnv("DB_HOST", "127.0.0.1"),
		Port:         GetEnv("DB_PORT", "5432"),
		Username:     GetEnv("DB_USERNAME", "root"),
		Password:     GetEnv("DB_PASSWORD", "1234"),
		Database:     GetEnv("DB_NAME", "mwp_db"),
		Dialect:      GetEnv("DB_DIALECT", "postgres"),
		MigrationDir: GetEnv("MIGRATION_DIR", "./migrations/"),
	}
}
