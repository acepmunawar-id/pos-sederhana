// Package config loads application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv   string
	HTTPPort string
	Timezone *time.Location

	DatabaseURL   string
	DBMaxConns    int32
	DBAutoMigrate bool

	JWTAccessSecret string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration

	CORSAllowedOrigins []string

	UploadDir     string
	UploadBaseURL string
	MaxUploadSize int64
}

// Load reads .env (if present) and the process environment.
func Load() (*Config, error) {
	_ = godotenv.Load()

	tzName := getEnv("APP_TIMEZONE", "Asia/Jakarta")
	loc, err := time.LoadLocation(tzName)
	if err != nil {
		return nil, fmt.Errorf("invalid APP_TIMEZONE %q: %w", tzName, err)
	}

	accessTTL, err := time.ParseDuration(getEnv("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TTL: %w", err)
	}
	refreshTTL, err := time.ParseDuration(getEnv("JWT_REFRESH_TTL", "168h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TTL: %w", err)
	}

	cfg := &Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		Timezone:           loc,
		DatabaseURL:        getEnv("DATABASE_URL", ""),
		DBMaxConns:         int32(getEnvInt("DB_MAX_CONNS", 10)),
		DBAutoMigrate:      getEnvBool("DB_AUTO_MIGRATE", true),
		JWTAccessSecret:    getEnv("JWT_ACCESS_SECRET", ""),
		JWTAccessTTL:       accessTTL,
		JWTRefreshTTL:      refreshTTL,
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")),
		UploadDir:          getEnv("UPLOAD_DIR", "./uploads"),
		UploadBaseURL:      strings.TrimRight(getEnv("UPLOAD_BASE_URL", "http://localhost:8080/uploads"), "/"),
		MaxUploadSize:      int64(getEnvInt("MAX_UPLOAD_SIZE_MB", 2)) << 20,
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if len(cfg.JWTAccessSecret) < 32 {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET must be at least 32 characters")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}
