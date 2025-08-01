package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	Log       LogConfig
	FileStore FileStoreConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Path           string
	MigrationsPath string
}

// FileStoreConfig holds the file store configuration
type FileStoreConfig struct {
	UploadDir string
}

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	SessionCookieName   string
	SessionCookieDomain string
	SessionCookieSecure bool
	SessionMaxAge       int
	JWTSecretKey        string
	JWTTokenDuration    int // in seconds
}

// LogConfig holds the logging configuration
type LogConfig struct {
	Level       string
	TimeFormat  string
	ShowCaller  bool
	FilePath    string
	EnableColor bool
}

// Load loads the configuration from environment variables
func Load() Config {
	return Config{
		FileStore: FileStoreConfig{
			UploadDir: getEnv("FILE_STORE_UPLOAD_DIR", "./data/uploads"),
		},
		Server: ServerConfig{
			Host:         getEnv("SERVER_HOST", "localhost"),
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 5*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			Path:           getEnv("DB_PATH", "./data/social_network.db"),
			MigrationsPath: getEnv("MIGRATIONS_PATH", "./pkg/db/migrations/sqlite"),
		},
		Auth: AuthConfig{
			SessionCookieName:   getEnv("SESSION_COOKIE_NAME", "session"),
			SessionCookieDomain: getEnv("SESSION_COOKIE_DOMAIN", ""),
			SessionCookieSecure: getEnvAsBool("SESSION_COOKIE_SECURE", false),
			SessionMaxAge:       getEnvAsInt("SESSION_MAX_AGE", 86400), // 24 hours
			JWTSecretKey:        getEnv("JWT_SECRET_KEY", "your-secret-key-change-in-production"),
			JWTTokenDuration:    getEnvAsInt("JWT_TOKEN_DURATION", 86400), // 24 hours
		},
		Log: LogConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			TimeFormat: getEnv("LOG_TIME_FORMAT", "2006-01-02 15:04:05"),
			ShowCaller: getEnvAsBool("LOG_SHOW_CALLER", true),
			// FilePath:   getEnv("LOG_FILE_PATH", "./data/logs/social_network.log"),
			EnableColor: getEnvAsBool("LOG_ENABLE_COLOR", true),
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvAsBool gets an environment variable as a boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvAsDuration gets an environment variable as a duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
