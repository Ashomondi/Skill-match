package config

import (
	"log"
	"os"

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

func bedrockRegion() string {
	return getEnv("BEDROCK_REGION", "us-east-1")
}

func bedrockModelID() string {
	return getEnv("BEDROCK_MODEL_ID", "")
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