package main

import (
	"log"
	"net/http"

	"skill-match/backend/config"
	"skill-match/backend/middleware"
	"skill-match/backend/routes"
)

func main() {
	cfg := config.Load()

	mux := routes.NewMux()
	routes.RegisterAll(mux,
		routes.RegisterHealth,
		
	)

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