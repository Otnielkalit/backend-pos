package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
// All values are read once at startup; there is no hot-reload.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Port string
	Env  string // "development" | "production"
	Name string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
}

// Load reads .env file (if present) and populates Config from environment variables.
// Returns an error if any required variable is missing.
func Load() (*Config, error) {
	// Load .env file — ignore error if file not found (e.g., in production via injected env)
	_ = godotenv.Load()

	cfg := &Config{}

	// App
	cfg.App.Port = getEnv("APP_PORT", "8080")
	cfg.App.Env = getEnv("APP_ENV", "development")
	cfg.App.Name = getEnv("APP_NAME", "pos-backend")

	// Database
	dbURL, err := requireEnv("DB_URL")
	if err != nil {
		return nil, err
	}
	cfg.Database.URL = dbURL
	cfg.Database.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 25)
	cfg.Database.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 10)
	cfg.Database.ConnMaxLifetime = time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)) * time.Minute

	// Redis
	cfg.Redis.URL = getEnv("REDIS_URL", "redis://localhost:6379")

	// JWT
	jwtSecret, err := requireEnv("JWT_SECRET")
	if err != nil {
		return nil, err
	}
	cfg.JWT.Secret = jwtSecret
	cfg.JWT.AccessTokenTTL = time.Duration(getEnvInt("JWT_ACCESS_TOKEN_TTL_HOURS", 24)) * time.Hour

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func requireEnv(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("required environment variable %q is not set", key)
	}
	return v, nil
}

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultValue
	}
	return n
}
