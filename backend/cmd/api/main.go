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

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	ctx := context.Background()

<<<<<<< HEAD
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

		if cfg.S3Bucket != "" {
			s3Client, err := clients.NewS3Client(ctx, clients.S3Config{
				Region:         cfg.AWSRegion,
				Bucket:         cfg.S3Bucket,
				Endpoint:       cfg.S3Endpoint,
				AccessKey:      cfg.S3AccessKey,
				SecretKey:      cfg.S3SecretKey,
				ForcePathStyle: cfg.S3ForcePathStyle,
			})
			if err != nil {
				log.Fatalf("init s3 client: %v", err)
			}

			resumeService := services.NewResumeService(repositories.NewResumeRepository(pool), s3Client)
			routes.RegisterResumes(mux, handlers.NewResumeHandler(resumeService), middleware.Auth(jwtManager))
		} else {
			log.Println("warning: S3_BUCKET_NAME not set — resume endpoints disabled")
		}
=======
	ctx := context.Background()

	jwtManager := utils.NewJWTManager(jwtSecret(cfg), 24*time.Hour)

	mux := routes.NewMux()

	var pool *pgxpool.Pool
	if cfg.DatabaseURL != "" {
		var err error
		pool, err = clients.NewPool(ctx, cfg.DatabaseURL, clients.PoolOptions{})
		if err != nil {
			log.Fatalf("connect to database: %v", err)
		}
		defer pool.Close()

		authService := services.NewAuthService(
			repositories.NewUserRepository(pool),
			jwtManager,
		)

		routes.RegisterAuth(mux, handlers.NewAuthHandler(authService))

		jobRepo := repositories.NewJobRepository(pool)
		savedJobs := handlers.NewSavedJobsHandler(services.NewSavedJobService(repositories.NewSavedJobRepository(pool)))
		routes.RegisterSavedJobs(mux, savedJobs, jwtManager)
		jobService := services.NewJobService(jobRepo, services.NewSeedJobSource())

		ingested, skipped, err := jobService.IngestJobs(ctx)
		if err != nil {
			log.Printf("WARNING: job ingestion failed: %v", err)
		} else {
			log.Printf("job ingestion: %d ingested, %d skipped", ingested, skipped)
		}
	} else {
		log.Println("WARNING: DATABASE_URL not set — auth endpoints are disabled")
>>>>>>> dev
	}

	var s3Client *clients.S3Client
	if cfg.S3Bucket != "" {
		var err error
		s3Client, err = clients.NewS3Client(ctx, cfg.AWSRegion, cfg.S3Bucket)
		if err != nil {
			log.Printf("WARNING: failed to connect to S3: %v — storage health checks disabled", err)
		}
	} else {
		log.Println("WARNING: S3_BUCKET_NAME not set — storage health checks disabled")
	}

	healthHandler := handlers.NewHealthHandler(pool, s3Client)
	routes.RegisterAll(mux,
		func(m *http.ServeMux) { routes.RegisterHealth(m, healthHandler) },
	)

	handler := middleware.Chain(mux,
		middleware.Logging,
		middleware.Recovery,
		middleware.CORS,
	)

	log.Printf("listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal(err)
	}
}

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
