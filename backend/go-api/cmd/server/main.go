package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/OmaleGrace/Grace-Predict/internal/auth"
	"github.com/OmaleGrace/Grace-Predict/internal/database"
	"github.com/OmaleGrace/Grace-Predict/internal/handlers"
	"github.com/OmaleGrace/Grace-Predict/internal/migrations"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env when running locally.
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found; using environment variables")
	}

	// Connect to PostgreSQL.
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Connected to PostgreSQL")
	if err := migrations.Run(context.Background(), db); err != nil {
		log.Fatal(err)
	}

	log.Println("Database migrations completed")

	mlServiceURL := os.Getenv("ML_SERVICE_URL")
	if mlServiceURL == "" {
		mlServiceURL = "http://localhost:8000"
	}

	port := os.Getenv("GO_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status":"ok","service":"grace-predict-api"}`)
	})

	predictionHandler := handlers.NewPredictionHandler(mlServiceURL)

	authHandler := handlers.NewAuthHandler(db)

	mux.HandleFunc(
		"/api/auth/register",
		authHandler.Register,
	)

	mux.HandleFunc(
		"/api/auth/login",
		authHandler.Login,
	)

	mux.Handle(
		"/api/auth/me",
		auth.RequireAuth(http.HandlerFunc(authHandler.Me)),
	)
	mux.HandleFunc(
		"/api/predict/test",
		predictionHandler.TestPrediction,
	)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf(
		"Grace Predict API running on http://localhost:%s",
		port,
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
