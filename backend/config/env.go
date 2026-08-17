package config

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func LoadEnvFile() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, relying on system environment variables")
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return value
}

func awsRegion() string {
	return getEnv("AWS_REGION", "us-east-1")
}

func s3Bucket() string {
	return getEnv("S3_BUCKET_NAME", "")
}

// jwtSecret returns the configured JWT secret. When none is set (local
// development), an ephemeral random secret is generated so authentication
// works out of the box; tokens become invalid on restart. Production must
// always set JWT_SECRET explicitly.
func jwtSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	log.Println("warning: JWT_SECRET not set — generating an ephemeral secret; tokens will be invalid after restart. Set JWT_SECRET for a stable secret.")
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "insecure-ephemeral-dev-secret"
	}
	return hex.EncodeToString(buf)
}

// jwtExpiry returns the JWT lifetime parsed from JWT_EXPIRATION (default 24h).
func jwtExpiry() time.Duration {
	raw := getEnv("JWT_EXPIRATION", "24h")
	duration, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("warning: invalid JWT_EXPIRATION %q, using 24h", raw)
		return 24 * time.Hour
	}
	return duration
}

func mcpEndpoint() string {
	return getEnv("MCP_ENDPOINT", "https://cockroachlabs.cloud/mcp")
}

func mcpAPIKey() string {
	return getEnv("MCP_API_KEY", "")
}

func mcpClusterID() string {
	return getEnv("MCP_CLUSTER_ID", "")
}

func bedrockEmbedModelID() string {
	return getEnv("BEDROCK_EMBED_MODEL_ID", "")
}
