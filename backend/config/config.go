package config

import "time"

type Config struct {
	Port string

	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration

	AWSRegion string
	S3Bucket  string

	BedrockRegion  string
	BedrockModelID string

	MCPEndpoint  string
	MCPAPIKey    string
	MCPClusterID string

	BedrockEmbedModelID string
}

func Load() *Config {
	LoadEnvFile()

	return &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   jwtSecret(),
		JWTExpiry:   jwtExpiry(),

		AWSRegion: awsRegion(),
		S3Bucket:  s3Bucket(),

		BedrockRegion:  bedrockRegion(),
		BedrockModelID: bedrockModelID(),
		MCPClusterID:   mcpClusterID(),

		BedrockEmbedModelID: bedrockEmbedModelID(),
	}
}
