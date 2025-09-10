package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	// sqlc-generated package (adjust the path to match your project layout)
	"github.com/AymaneIsmail/chirpy/internal/database"
)

type apiConfig struct {
	db             *database.Queries
	fileServerHits atomic.Int32
	Platform       string
	JWTSecret string
	PolkaKey string
}

func main() {
	_ = godotenv.Load() // don't fail if .env file is missing

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL is not set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM is not set")
	}

	JWTSecret := os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	polkaKey := os.Getenv("POLKA_KEY")
	if JWTSecret == "" {
		log.Fatal("POLKA_KEY is not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Cannot open database connection (%s): %v", dbURL, err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Cannot reach database (%s): %v", dbURL, err)
	}

	dbQueries := database.New(db)

	const port = "8080"
	mux := http.NewServeMux()

	cfg := apiConfig{
		db:       dbQueries,
		Platform: platform,
		JWTSecret: JWTSecret,
		PolkaKey: polkaKey,
	}

	// File server with metrics middleware
	fileHandler := http.StripPrefix("/app", cfg.middlewareMetricsInc(http.FileServer(http.Dir("."))))
	mux.Handle("/app/", fileHandler)

	// API routes
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", cfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", cfg.resetUserHandler)
	mux.HandleFunc("POST /api/chirps", cfg.createChirpHandler)
	mux.HandleFunc("GET /api/chirps", cfg.GetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.GetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirpHandler)
	mux.HandleFunc("POST /api/users", cfg.createUserHandler)
	mux.HandleFunc("POST /api/login", cfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", cfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", cfg.revokeRefreshTokenHandler)
	mux.HandleFunc("PUT  /api/users", cfg.updateUserHandler)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.webhooks)


	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("✅ Server is listening on port %s", port)
	log.Fatal(server.ListenAndServe())
}
