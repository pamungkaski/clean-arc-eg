package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"ki.com/clean-arc-example/src"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	// Load .env if present (optional)
	_ = godotenv.Load()

	// ---- Config ----
	mongoURI := getenv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getenv("MONGO_DB", "budgetdb")
	collName := getenv("MONGO_COLLECTION", "budgets")
	addr := getenv("ADDR", ":8080")

	// ---- Mongo client ----
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("mongo connect: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("mongo ping: %v", err)
	}
	db := client.Database(dbName)

	// ---- Wire repo → usecase → http ----
	repo := src.NewBudgetMongo(db, collName)
	uc := src.NewUsecase(repo)
	httpHandler := src.NewBudgetHTTP(uc)

	mux := http.NewServeMux()
	mux.HandleFunc("/budgets", httpHandler.GetAllBudgets)

	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// ---- Start server ----
	go func() {
		log.Printf("HTTP server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	// ---- Graceful shutdown ----
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	if err := client.Disconnect(shutdownCtx); err != nil {
		log.Printf("mongo disconnect error: %v", err)
	}
	log.Println("bye 👋")
}
