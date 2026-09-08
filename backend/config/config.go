package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Port string

	DatabaseURL   string
	JWTSecret     string
	StorageDir    string
	CORSOrigin    string
	AllowedOrigin string

	AWSRegion        string
	S3Bucket         string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool

	BedrockRegion       string
	BedrockModelID      string
	BedrockChatModelID  string
	BedrockEmbedModelID string
}

func Load() (*Config, error) {
	LoadEnvFile()

	dbURL := getEnv("DATABASE_URL", "")
	if dbURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL environment variable is required")
	}

	cfg := &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL:   dbURL,
		JWTSecret:     jwtSecret(),
		StorageDir:    getEnv("STORAGE_DIR", "./data/resumes"),
		CORSOrigin:    getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),
		AllowedOrigin: getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:5173"),

		AWSRegion:        awsRegion(),
		S3Bucket:         s3Bucket(),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3AccessKey:      getEnv("AWS_ACCESS_KEY_ID", ""),
		S3SecretKey:      getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle: getEnv("S3_FORCE_PATH_STYLE", "true") == "true",

		BedrockRegion:       bedrockRegion(),
		BedrockModelID:      bedrockModelID(),
		BedrockChatModelID:  bedrockChatModelID(),
		BedrockEmbedModelID: bedrockEmbedModelID(),
	}

	return cfg, nil
}

// jwtSecret returns the configured JWT secret, or an ephemeral random
// secret for development when one is not set. This keeps startup simple
// for local/offline development while explicitly warning that tokens
// will not survive a restart.
func jwtSecret() string {
	if value := os.Getenv("JWT_SECRET"); value != "" {
		return value
	}

	log.Println("WARNING: JWT_SECRET not set — using an ephemeral development secret; tokens will be invalidated on restart")
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "skill-match-development-secret"
	}
	return hex.EncodeToString(buf)
}
