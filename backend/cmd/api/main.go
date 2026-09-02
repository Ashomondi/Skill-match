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

		// Instantiate repositories and services
		jobRepo := repositories.NewJobRepository(pool)
		seedSource := services.NewSeedJobSource()
		jobSource := services.NewExternalJobSource(seedSource)
		matchingSvc := services.NewMatchingService()

		savedJobs := handlers.NewSavedJobsHandler(services.NewSavedJobService(repositories.NewSavedJobRepository(pool)))
		routes.RegisterSavedJobs(mux, savedJobs, jwtManager)
		routes.RegisterApplications(mux,
			handlers.NewApplicationHandler(services.NewApplicationService(repositories.NewApplicationRepository(pool))),
			jwtManager,
		)

		jobService := services.NewJobService(jobRepo, jobSource)
		routes.RegisterJobs(mux, handlers.NewJobsHandler(jobService), jwtManager)
		routes.RegisterRecommendations(
			mux,
			handlers.NewRecommendationHandler(services.NewRecommendationService(jobRepo, repositories.NewProfileRepository(pool), matchingSvc)),
			jwtManager,
		)

		// Async boot ingestion with retries and backoff
		go func() {
			ingestCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()

			if ingested, skipped, err := jobService.IngestJobsWithRetry(ingestCtx, 3); err != nil {
				log.Printf("WARNING: background boot job ingestion failed: %v", err)
			} else {
				log.Printf("job ingestion completed: %d ingested, %d skipped", ingested, skipped)
			}
		}()

		if cfg.BedrockChatModelID != "" {
			bedrockClient, err := clients.NewBedrockClient(ctx, cfg.BedrockRegion, cfg.BedrockChatModelID)
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
				routes.RegisterChat(mux, handlers.NewChatHandler(chatService), jwtManager)
				routes.RegisterTailor(mux, handlers.NewTailorHandler(aiService), jwtManager)
			}
		} else {
			log.Println("WARNING: BEDROCK_CHAT_MODEL_ID not set — chat disabled")
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

	if pool != nil {
		var storage services.ObjectStorage
		if s3Client != nil {
			storage = s3Client
		}
		resumeService := services.NewResumeService(repositories.NewResumeRepository(pool), storage)
		routes.RegisterResumes(mux, handlers.NewResumeHandler(resumeService), jwtManager)
	}

	healthHandler := handlers.NewHealthHandler(pool, s3Client)
	routes.RegisterAll(mux,
		func(m *http.ServeMux) { routes.RegisterHealth(m, healthHandler) },
	)

	handler := middleware.Chain(mux,
		middleware.Logging,
		middleware.Recovery,
		middleware.CORS(cfg.AllowedOrigin),
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
