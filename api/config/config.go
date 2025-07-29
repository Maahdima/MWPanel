package config

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	"github.com/maahdima/mwp/ui"
)

type AppConfig struct {
	Mode               string
	Host               string
	Port               string
	ConsoleLogFormat   string
	DataDirPath        string
	UIAssetsFs         fs.FS
	PeerFilesDir       string
	TrafficJobInterval string
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	Dialect  string
}

type AdminConfig struct {
	Username string
	Password string
}

type AuthConfig struct {
	AccessTokenTTL  string
	RefreshTokenTTL string
}

func init() {
	_ = loadEnv()
}

func loadEnv() error {
	return godotenv.Load(getEnv("ENV_FILE", "config/.env"))
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetAppConfig() AppConfig {
	dataDir := getEnv("DATA_DIR", "")
	if dataDir == "" {
		userConfigDir, err := os.UserConfigDir()
		if err == nil {
			info, statErr := os.Stat(userConfigDir)
			if statErr == nil && info.IsDir() {
				dataDir = filepath.Join(userConfigDir, "mwp")
			}
		}

		if dataDir == "" {
			exePath, err := os.Executable()
			if err != nil {
				log.Fatalf("Failed to get executable path: %v", err)
			}
			dataDir = filepath.Join(filepath.Dir(exePath), "mwp")
		}
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	return AppConfig{
		Mode:               getEnv("MODE", "production"),
		Host:               getEnv("SERVER_HOST", "0.0.0.0"),
		Port:               getEnv("SERVER_PORT", "3000"),
		ConsoleLogFormat:   getEnv("CONSOLE_LOG_FORMAT", "plain"),
		UIAssetsFs:         echo.MustSubFS(ui.GetUIAssets(), "dist"),
		PeerFilesDir:       getEnv("PEER_FILES_DIR", filepath.Join(dataDir, "peer-files")),
		DataDirPath:        dataDir,
		TrafficJobInterval: getEnv("TRAFFIC_JOB_INTERVAL", "300"),
	}
}

func GetDBConfig() DBConfig {
	dialect := getEnv("DB_DIALECT", "sqlite")

	defaultDatabaseName := "mwp_db"
	if dialect == "sqlite" {
		appCfg := GetAppConfig()
		defaultDatabaseName = filepath.Join(appCfg.DataDirPath, "mwp.db")
	}

	return DBConfig{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "5432"),
		Username: getEnv("DB_USERNAME", "root"),
		Password: getEnv("DB_PASSWORD", "1234"),
		Database: getEnv("DB_NAME", defaultDatabaseName),
		Dialect:  dialect,
	}
}

func GetAdminConfig() AdminConfig {
	return AdminConfig{
		Username: getEnv("ADMIN_USERNAME", "mwpadmin"),
		Password: getEnv("ADMIN_PASSWORD", "mwpadmin"),
	}
}

func GetAuthConfig() AuthConfig {
	return AuthConfig{
		AccessTokenTTL:  getEnv("AUTH_ACCESS_TOKEN_TTL", "900"),
		RefreshTokenTTL: getEnv("AUTH_REFRESH_TOKEN_TTL", "86400"),
	}
}
