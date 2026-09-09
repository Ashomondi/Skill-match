package main

import (
	"context"
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

	jwtManager := utils.NewJWTManager(cfg.JWTSecret, 24*time.Hour)

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

		// Chat and CV tailoring are powered by Google Gemini. When no API key is
		// configured the endpoints are not registered.
		if cfg.GeminiAPIKey != "" {
			geminiClient, err := clients.NewGeminiClient(cfg.GeminiAPIKey, cfg.GeminiModel)
			if err != nil {
				log.Printf("WARNING: failed to init Gemini client: %v — chat disabled", err)
			} else {
				conversationRepo := repositories.NewConversationRepository(pool)
				aiService := services.NewAIService(services.NewAIServiceInput{
					Generator:     geminiClient,
					Conversations: conversationRepo,
					Resumes:       repositories.NewResumeRepository(pool),
				})
				memoryService := services.NewMemoryService(conversationRepo)
				chatService := services.NewChatService(aiService, memoryService)
				routes.RegisterChat(mux, handlers.NewChatHandler(chatService), jwtManager)
				routes.RegisterTailor(mux, handlers.NewTailorHandler(aiService), jwtManager)
				log.Printf("INFO: chat/tailor enabled with Gemini model %q", cfg.GeminiModel)
			}
		} else {
			log.Println("WARNING: GEMINI_API_KEY not set — chat disabled")
		}
	} else {
		log.Println("WARNING: DATABASE_URL not set — auth endpoints are disabled")
	}

	// Resume files are stored on the local filesystem and served from /storage.
	var storage *clients.LocalFS
	storage, err = clients.NewLocalFS(cfg.StorageDir, "/storage")
	if err != nil {
		log.Fatalf("init local storage: %v", err)
	}
	mux.Handle("/storage/", storage.Handler())
	log.Printf("INFO: local file storage active at %s (served at /storage)", cfg.StorageDir)

	if pool != nil {
		resumeService := services.NewResumeService(repositories.NewResumeRepository(pool), storage)
		routes.RegisterResumes(mux, handlers.NewResumeHandler(resumeService), jwtManager)
	}

	healthHandler := handlers.NewHealthHandler(pool, storage)
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
