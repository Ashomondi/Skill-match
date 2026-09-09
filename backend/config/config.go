package config

import (
	"fmt"
)

type Config struct {
	Port string

	DatabaseURL   string
	JWTSecret     string
	StorageDir    string
	AllowedOrigin string

	GeminiAPIKey string
	GeminiModel  string
}

func Load() (*Config, error) {
	LoadEnvFile()

	dbURL := getEnv("DATABASE_URL", "")
	if dbURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL environment variable is required")
	}

	jwtSecret := getEnv("JWT_SECRET", "")
	if jwtSecret == "" {
		return nil, fmt.Errorf("config: JWT_SECRET environment variable is required")
	}

	cfg := &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL:   dbURL,
		JWTSecret:     jwtSecret,
		StorageDir:    getEnv("STORAGE_DIR", "./data/resumes"),
		AllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),

		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-2.5-flash"),
	}

	return cfg, nil
}
