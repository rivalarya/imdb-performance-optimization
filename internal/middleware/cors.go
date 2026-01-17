package middleware

import (
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"imdb-performance-optimization/internal/config"
)

func SetupCORS(router *gin.Engine, cfg *config.Config) {
	if !cfg.EnableCORS {
		return
	}

	allowedOrigins := strings.Split(cfg.AllowedOrigin, ",")
	
	for i, origin := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(origin)
	}

	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins, // Now supports multiple origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:          12 * time.Hour,
	}

	router.Use(cors.New(corsConfig))
}