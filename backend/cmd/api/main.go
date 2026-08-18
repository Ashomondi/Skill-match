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
	"skill-match/backend/migrations"
	"skill-match/backend/repositories"
	"skill-match/backend/routes"
	"skill-match/backend/services"
	"skill-match/backend/utils"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

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

		if err := migrations.Apply(ctx, pool); err != nil {
			log.Fatalf("apply migrations: %v", err)
		}
		log.Println("database migrations up to date")

		authService := services.NewAuthService(
			repositories.NewUserRepository(pool),
			jwtManager,
		)
		routes.RegisterAuth(mux, handlers.NewAuthHandler(authService))

		jobRepo := repositories.NewJobRepository(pool)
		savedJobs := handlers.NewSavedJobsHandler(services.NewSavedJobService(repositories.NewSavedJobRepository(pool)))
		routes.RegisterSavedJobs(mux, savedJobs, jwtManager)
		routes.RegisterApplications(mux,
			handlers.NewApplicationHandler(services.NewApplicationService(repositories.NewApplicationRepository(pool))),
			jwtManager,
		)
		jobService := services.NewJobService(jobRepo, services.NewSeedJobSource())

		if ingestible, ok := interface{}(jobService).(interface {
			IngestJobs(ctx context.Context) (int, int, error)
		}); ok {
			ingested, skipped, err := ingestible.IngestJobs(ctx)
			if err != nil {
				log.Printf("WARNING: job ingestion failed: %v", err)
			} else {
				log.Printf("job ingestion: %d ingested, %d skipped", ingested, skipped)
			}
		}

		if cfg.BedrockModelID != "" {
			bedrockClient, err := clients.NewBedrockClient(ctx, cfg.BedrockRegion, cfg.BedrockModelID)
			if err != nil {
				log.Printf("WARNING: failed to init Bedrock client: %v — chat disabled", err)
			} else {
				conversationRepo := repositories.NewConversationRepository(pool)
				aiService := services.NewAIService(services.NewAIServiceInput{
					Bedrock:       bedrockClient,
					Conversations: conversationRepo,
					Resumes:       repositories.NewResumeRepository(pool),
				})
				memoryService := services.NewMemoryService(conversationRepo)
				chatService := services.NewChatService(aiService, memoryService)
				routes.RegisterChat(mux, handlers.NewChatHandler(chatService))
			}
		} else {
			log.Println("WARNING: BEDROCK_MODEL_ID not set — chat disabled")
		}
	} else {
		log.Println("WARNING: DATABASE_URL not set — auth endpoints are disabled")
	}

	var s3Client *clients.S3Client
	if cfg.S3Bucket != "" {
		var err error
		s3Client, err = clients.NewS3Client(ctx, clients.S3Config{
			Region:         cfg.AWSRegion,
			Bucket:         cfg.S3Bucket,
			Endpoint:       cfg.S3Endpoint,
			AccessKey:      cfg.S3AccessKey,
			SecretKey:      cfg.S3SecretKey,
			ForcePathStyle: cfg.S3ForcePathStyle,
		})
		if err != nil {
			log.Printf("WARNING: failed to connect to S3: %v — storage health checks disabled", err)
		}
	} else {
		log.Println("WARNING: S3_BUCKET_NAME not set — storage health checks disabled")
	}

	if pool != nil && s3Client != nil {
		resumeService := services.NewResumeService(repositories.NewResumeRepository(pool), s3Client)
		routes.RegisterResumes(mux, handlers.NewResumeHandler(resumeService), jwtManager)
	}

	// ... setup handlers, services, and route registrations above ...

    healthHandler := handlers.NewHealthHandler(pool, s3Client)
    routes.RegisterAll(mux,
        func(m *http.ServeMux) { routes.RegisterHealth(m, healthHandler) },
    )

    // 👇 Middleware chain setup belongs here
  handler := middleware.Chain(mux,
	middleware.Logging,
	middleware.Recovery,
	middleware.CORS(cfg.AllowedOrigin), // Pass cfg.AllowedOrigin (or "*" / your origin string)
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