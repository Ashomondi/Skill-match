package config

type Config struct {
	Port string

	DatabaseURL string
	JWTSecret   string
}

func Load() *Config {
	LoadEnvFile()

	return &Config{
		Port: getEnv("PORT", "8080"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		JWTSecret:   getEnv("JWT_SECRET", ""),
	}
}