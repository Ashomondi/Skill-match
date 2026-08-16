package config

import "time"

type Config struct {
	Port string

	DatabaseURL string
	JWTSecret   string
	JWTExpiry   time.Duration

	AWSRegion        string
	S3Bucket         string
	S3Endpoint       string
	S3AccessKey      string
	S3SecretKey      string
	S3ForcePathStyle bool
}

func Load() *Config {
	LoadEnvFile()

	return &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   jwtSecret(),
		JWTExpiry:   jwtExpiry(),

		AWSRegion:        awsRegion(),
		S3Bucket:         s3Bucket(),
		S3Endpoint:       getEnv("S3_ENDPOINT", ""),
		S3AccessKey:      getEnv("AWS_ACCESS_KEY_ID", ""),
		S3SecretKey:      getEnv("AWS_SECRET_ACCESS_KEY", ""),
		S3ForcePathStyle: getEnv("S3_FORCE_PATH_STYLE", "true") == "true",
	}
}
