package main

import (
	"context"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"skill-match/backend/clients"
	"skill-match/backend/config"
	"skill-match/backend/handlers"
	"skill-match/backend/middleware"
	"skill-match/backend/migrations"
	"skill-match/backend/repositories"
	"skill-match/backend/routes"
	"skill-match/backend/services"
	"skill-match/backend/utils"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiry)

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		p, err := clients.NewPool(ctx, cfg.DatabaseURL, clients.PoolOptions{})
		if err != nil {
			log.Fatalf("connect to CockroachDB: %v", err)
		}
		defer p.Close()
		pool = p
		log.Println("connected to CockroachDB")

		if err := migrations.Apply(ctx, pool); err != nil {
			log.Fatalf("apply migrations: %v", err)
		}
		log.Println("database migrations up to date")
	} else {
		log.Println("warning: DATABASE_URL not set — running without persistence (health will report degraded)")
	}

	mux := routes.NewMux()
	routes.RegisterHealth(mux, pool)

	if pool != nil {
		userRepo := repositories.NewUserRepository(pool)
		authService := services.NewAuthService(userRepo, jwtManager)
		routes.RegisterAuth(mux, handlers.NewAuthHandler(authService))
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
