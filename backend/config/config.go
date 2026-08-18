package config

type Config struct {
	Port string

	DatabaseURL string
	JWTSecret   string

	AWSRegion        string
	S3Bucket         string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool

	BedrockRegion      string
	BedrockModelID     string
	BedrockChatModelID string

	MCPEndpoint  string
	MCPAPIKey    string
	MCPClusterID string

	BedrockEmbedModelID string
	AgentRouterAPIKey   string
	AgentRouterBaseURL  string
	AgentRouterModel    string
}

func Load() *Config {
	LoadEnvFile()

	return &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),

		AWSRegion:        awsRegion(),
		S3Bucket:         s3Bucket(),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3AccessKey:      getEnv("AWS_ACCESS_KEY_ID", ""),
		S3SecretKey:      getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle: getEnv("S3_FORCE_PATH_STYLE", "true") == "true",

		BedrockRegion:      bedrockRegion(),
		BedrockModelID:     bedrockModelID(),
		BedrockChatModelID: bedrockChatModelID(),
		MCPClusterID:       mcpClusterID(),

		BedrockEmbedModelID: bedrockEmbedModelID(),
		AgentRouterAPIKey:   getEnv("AGENTROUTER_API_KEY", ""),
		AgentRouterBaseURL:  getEnv("AGENTROUTER_BASE_URL", "https://agentrouter.org/v1"),
		AgentRouterModel:    getEnv("AGENTROUTER_MODEL", "gpt-5.6-sol"),
	}
}
