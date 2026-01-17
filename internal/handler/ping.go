package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type PingResponse struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
}

func PingHandler(c *gin.Context) {
	response := PingResponse{
		Message:   "pong",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    "healthy",
	}

	c.JSON(http.StatusOK, response)
}

func RegisterPingRoutes(router *gin.Engine) {
	router.GET("/ping", PingHandler)
}