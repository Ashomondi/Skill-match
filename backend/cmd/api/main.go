package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"time"

	"skill-match/backend/clients"
	"skill-match/backend/config"
	"skill-match/backend/handlers"
	"skill-match/backend/middleware"
	"skill-match/backend/repositories"
	"skill-match/backend/routes"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

func main() {
	cfg := config.Load()

	jwtManager := utils.NewJWTManager(jwtSecret(cfg), 24*time.Hour)

	mux := routes.NewMux()
	routes.RegisterAll(mux,
		routes.RegisterHealth,
	)

	if cfg.DatabaseURL != "" {
		ctx := context.Background()

		pool, err := clients.NewPool(ctx, cfg.DatabaseURL, clients.PoolOptions{})
		if err != nil {
			log.Fatalf("connect to database: %v", err)
		}
		defer pool.Close()

		authService := services.NewAuthService(
			repositories.NewUserRepository(pool),
			jwtManager,
		)

		routes.RegisterAuth(
			mux,
			handlers.NewAuthHandler(authService),
		)
		routes.RegisterAuth(mux, handlers.NewAuthHandler(authService))
	} else {
		log.Println("WARNING: DATABASE_URL not set — auth endpoints are disabled")
	}

	handler := middleware.Chain(mux,
		middleware.Recovery,
		middleware.Logging,
		middleware.CORS,
	)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}

// jwtSecret returns the configured JWT secret, or an ephemeral random
// secret for development when JWT_SECRET is not set. Set JWT_SECRET in
// production so sessions survive restarts.
func jwtSecret(cfg *config.Config) string {
	if cfg.JWTSecret != "" {
		return cfg.JWTSecret
	}
	log.Println("WARNING: JWT_SECRET not set — using an ephemeral development secret")
	return devSecret()
}

func devSecret() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "skill-match-development-secret"
	}
	return hex.EncodeToString(buf)
}
