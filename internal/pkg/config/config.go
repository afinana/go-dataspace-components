package config

import (
	"os"
	"strconv"
)

// Config holds the common configuration values for the Dataspace Connector.
type Config struct {
	ServiceName     string
	Environment     string
	LogLevel        string
	ControlPlaneURL string
	DataPlaneURL    string
	Port            int
	DatabaseURL     string
}

// LoadConfig fetches values from environment variables or returns defaults.
func LoadConfig() *Config {
	return &Config{
		ServiceName:     getEnv("SERVICE_NAME", "dataspace-connector"),
		Environment:     getEnv("ENVIRONMENT", "production"),
		LogLevel:        getEnv("LOG_LEVEL", "INFO"),
		ControlPlaneURL: getEnv("CONTROL_PLANE_URL", "http://localhost:8081"),
		DataPlaneURL:    getEnv("DATA_PLANE_URL", "http://localhost:8082"),
		Port:            getEnvAsInt("PORT", 8080),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/dataspace_identity?sslmode=disable"),
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
