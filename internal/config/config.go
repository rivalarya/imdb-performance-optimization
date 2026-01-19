package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port          string
	Environment   string
	EnableCORS    bool
	AllowedOrigin string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	ServerPort    string
}

func Load() *Config {
	enableCORS, _ := strconv.ParseBool(os.Getenv("ENABLE_CORS"))

	return &Config{
		Port:          getEnv("PORT", "5000"),
		Environment:   getEnv("NODE_ENV", "development"),
		EnableCORS:    enableCORS,
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "https://example.com"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "postgres"),
		DBPassword:    getEnv("DB_PASSWORD", "password"),
		DBName:        getEnv("DB_NAME", "myapp"),
		ServerPort:    getEnv("PORT", "8080"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
