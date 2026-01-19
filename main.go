package main

import (
	"context"
	"log"
	"os"

	"imdb-performance-optimization/internal/config"
	"imdb-performance-optimization/internal/database"
	"imdb-performance-optimization/internal/handler"
	"imdb-performance-optimization/internal/middleware"
	"imdb-performance-optimization/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg := config.Load()

	pools, err := database.ConnectToPostgres(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pools.Close()

	conn, err := pools.UnoptimizedPool.Acquire(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Release()

	_, err = conn.Exec(context.Background(), `
	SET enable_indexscan = OFF;
	SET enable_bitmapscan = OFF;
	SET enable_indexonlyscan = OFF;
`)
	if err != nil {
		log.Fatal(err)
	}
	conn.Release()

	moviesRepo := repositories.NewMoviesRepository(pools.OptimizedPool, pools.UnoptimizedPool)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	middleware.SetupCORS(r, cfg)
	handler.RegisterPingRoutes(r)
	handler.RegisterMoviesRoutes(r, moviesRepo)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
