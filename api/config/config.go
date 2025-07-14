package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
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
		fmt.Println("Failed to load environment variables: " + err.Error())
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
	return AppConfig{
		Mode:             getEnv("MODE", "development"),
		Host:             getEnv("SERVER_HOST", "127.0.0.1"),
		Port:             getEnv("SERVER_PORT", "3000"),
		ConsoleLogFormat: getEnv("CONSOLE_LOG_FORMAT", "plain"),
		PublicDir:        getEnv("PUBLIC_DIR", "./public/"),
		PeerFilesDir:     getEnv("PEER_FILES_DIR", "./peer-files/"),
	}
}

func GetDBConfig() DBConfig {
	return DBConfig{
		Host:         getEnv("DB_HOST", "127.0.0.1"),
		Port:         getEnv("DB_PORT", "5432"),
		Username:     getEnv("DB_USERNAME", "root"),
		Password:     getEnv("DB_PASSWORD", "1234"),
		Database:     getEnv("DB_NAME", "mwp_db"),
		Dialect:      getEnv("DB_DIALECT", "postgres"),
		MigrationDir: getEnv("MIGRATION_DIR", "./migrations/"),
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
		Username: getEnv("ADMIN_USERNAME", "admin"),
		Password: getEnv("ADMIN_PASSWORD", "admin1234"),
	}
}
