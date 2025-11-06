package config

import (
	"os"
)

// Config holds application configuration
type Config struct {
	ServerPort string
	LogLevel   string
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		LogLevel:   getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
