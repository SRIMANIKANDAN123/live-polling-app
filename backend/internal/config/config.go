package config

import (
	"os"
)

// Config holds all environment-driven configuration for the service.
type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	RedisURL      string
	JWTSecret     string
	CORSOrigin    string
	Environment   string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads configuration from environment variables (populated via
// .env in local dev, or real environment variables in production).
func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		MongoURI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDatabase: getEnv("MONGODB_DATABASE", "livepoll"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     getEnv("JWT_SECRET", "change-me"),
		CORSOrigin:    getEnv("CORS_ORIGIN", "http://localhost:5173"),
		Environment:   getEnv("ENV", "development"),
	}
}
