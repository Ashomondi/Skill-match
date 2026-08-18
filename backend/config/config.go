package config

import (
	"fmt"
)

type Config struct {
	Port string

	DatabaseURL string
	JWTSecret   string
	CORSOrigin  string
	AllowedOrigin string

	AWSRegion        string
	S3Bucket         string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool

	BedrockRegion  string
	BedrockModelID string

	MCPEndpoint  string
	MCPAPIKey    string
	MCPClusterID string

	BedrockEmbedModelID string
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

		DatabaseURL: dbURL,
		JWTSecret:   jwtSecret,
		CORSOrigin:  getEnv("CORS_ALLOWED_ORIGIN", "http://localhost:3000"),

		AWSRegion:        awsRegion(),
		S3Bucket:         s3Bucket(),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3AccessKey:      getEnv("S3_ACCESS_KEY", getEnv("AWS_ACCESS_KEY_ID", "")),
		S3SecretKey:      getEnv("S3_SECRET_KEY", getEnv("AWS_SECRET_ACCESS_KEY", "")),
		S3ForcePathStyle: getEnv("S3_FORCE_PATH_STYLE", "true") == "true",

		BedrockRegion:  bedrockRegion(),
		BedrockModelID: bedrockModelID(),
		MCPClusterID:   mcpClusterID(),

		BedrockEmbedModelID: bedrockEmbedModelID(),
	}

	return cfg, nil
}