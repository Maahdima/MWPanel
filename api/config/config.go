package config

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"

	"github.com/maahdima/mwp/ui"
)

type AppConfig struct {
	Mode             string
	Host             string
	Port             string
	ConsoleLogFormat string
	DataDirPath      string
	UIAssetsFs       fs.FS
	PeerFilesDir     string
}

type DBConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
	Dialect  string
}

type ServerConfig struct {
	Comment   string
	Name      string
	IPAddress string
	APIPort   int
	Username  string
	Password  string
}

type AdminConfig struct {
	Username string
	Password string
}

func init() {
	err := loadEnv()
	if err != nil {
		log.Println("Failed to load environment variables: " + err.Error())
	}
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
		if err != nil {
			log.Fatalf("Failed to get user config directory: %v", err)
		}

		dataDir = filepath.Join(userConfigDir, "mwp")
	}

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	return AppConfig{
		Mode:             getEnv("MODE", "production"),
		Host:             getEnv("SERVER_HOST", "0.0.0.0"),
		Port:             getEnv("SERVER_PORT", "3000"),
		ConsoleLogFormat: getEnv("CONSOLE_LOG_FORMAT", "plain"),
		UIAssetsFs:       echo.MustSubFS(ui.GetUIAssets(), "dist"),
		PeerFilesDir:     getEnv("PEER_FILES_DIR", filepath.Join(dataDir, "peer-files")),
		DataDirPath:      dataDir,
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

func GetServerConfig() ServerConfig {
	apiPort := getEnv("SERVER_API_PORT", "80")

	var apiPortInt int
	if port, err := strconv.Atoi(apiPort); err == nil {
		apiPortInt = port
	} else {
		apiPortInt = 80
	}

	return ServerConfig{
		Comment:   getEnv("SERVER_COMMENT", "Default Server"),
		Name:      getEnv("SERVER_NAME", "Default Server"),
		IPAddress: getEnv("SERVER_IP_ADDRESS", "127.0.0.1"),
		APIPort:   apiPortInt,
		Username:  getEnv("SERVER_USERNAME", "admin"),
		Password:  getEnv("SERVER_PASSWORD", "admin1234"),
	}
}

func GetAdminConfig() AdminConfig {
	return AdminConfig{
		Username: getEnv("ADMIN_USERNAME", "mwpadmin"),
		Password: getEnv("ADMIN_PASSWORD", "mwpadmin"),
	}
}
