package config

type Config struct {
	Port string

	DatabaseURL string
	JWTSecret   string

	AWSRegion string
	S3Bucket  string

	BedrockRegion  string
	BedrockModelID string

	MCPEndpoint string
	MCPAPIKey   string
	MCPClusterID string

}

func Load() *Config {
	LoadEnvFile()

	return &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),

		AWSRegion: awsRegion(),
		S3Bucket:  s3Bucket(),

		BedrockRegion:  bedrockRegion(),
		BedrockModelID: bedrockModelID(),
		MCPClusterID: mcpClusterID(),
	}
}