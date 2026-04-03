package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv                 string
	AppBaseDir             string
	HTTPPort               string
	DatabaseURL            string
	JWTSecret              string
	JTTTL                  time.Duration
	AllowedOrigins         []string
	StorageDriver          string
	StorageBucket          string
	StorageRegion          string
	StorageEndpoint        string
	StorageAccessKey       string
	StorageSecretKey       string
	StorageUsePathStyle    bool
	LocalStoragePath       string
	GoogleAPIKey           string
	GoogleAnalysisModel    string
	GooglePromptModel      string
	GoogleImageModel       string
	WorkerPollInterval     time.Duration
	WorkerMaxAttempts      int
	GenerationRateLimitRPM int
	DefaultPlanCode        string
	StripeSecretKey        string
	AsaasAPIKey            string
	PayPalClientID         string
	PayPalClientSecret     string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:                 env("APP_ENV", "development"),
		AppBaseDir:             env("APP_BASE_DIR", "."),
		HTTPPort:               env("HTTP_PORT", "8080"),
		DatabaseURL:            env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/foto_magica?sslmode=disable"),
		JWTSecret:              env("JWT_SECRET", "replace-me"),
		JTTTL:                  duration("JWT_TTL", 24*time.Hour),
		AllowedOrigins:         csv("ALLOWED_ORIGINS", "http://localhost:5173,http://localhost:5174"),
		StorageDriver:          env("STORAGE_DRIVER", "local"),
		StorageBucket:          env("STORAGE_BUCKET", "foto-magica"),
		StorageRegion:          env("STORAGE_REGION", "us-east-1"),
		StorageEndpoint:        env("STORAGE_ENDPOINT", ""),
		StorageAccessKey:       env("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:       env("STORAGE_SECRET_KEY", ""),
		StorageUsePathStyle:    boolEnv("STORAGE_USE_PATH_STYLE", true),
		LocalStoragePath:       env("LOCAL_STORAGE_PATH", "./tmp/storage"),
		GoogleAPIKey:           env("GOOGLE_API_KEY", ""),
		GoogleAnalysisModel:    env("GOOGLE_ANALYSIS_MODEL", "gemini-2.5-flash"),
		GooglePromptModel:      env("GOOGLE_PROMPT_MODEL", "gemini-2.5-flash"),
		GoogleImageModel:       env("GOOGLE_IMAGE_MODEL", "gemini-2.5-flash-image"),
		WorkerPollInterval:     duration("WORKER_POLL_INTERVAL", 5*time.Second),
		WorkerMaxAttempts:      intEnv("WORKER_MAX_ATTEMPTS", 3),
		GenerationRateLimitRPM: intEnv("GENERATION_RATE_LIMIT_RPM", 10),
		DefaultPlanCode:        env("DEFAULT_PLAN_CODE", "growth"),
		StripeSecretKey:        env("STRIPE_SECRET_KEY", ""),
		AsaasAPIKey:            env("ASAAS_API_KEY", ""),
		PayPalClientID:         env("PAYPAL_CLIENT_ID", ""),
		PayPalClientSecret:     env("PAYPAL_CLIENT_SECRET", ""),
	}

	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func env(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func intEnv(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func boolEnv(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func duration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func csv(key, fallback string) []string {
	raw := env(key, fallback)
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		result = append(result, part)
	}
	return result
}

