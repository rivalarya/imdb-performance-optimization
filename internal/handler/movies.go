package handler

import (
	"net/http"

	"imdb-performance-optimization/internal/models"

	"github.com/gin-gonic/gin"
)

type MoviesHandler struct {
	repo models.MoviesRepository
}

func NewMoviesHandler(repo models.MoviesRepository) *MoviesHandler {
	return &MoviesHandler{repo: repo}
}

func (h *MoviesHandler) GetAll(c *gin.Context) {
	optimize := c.DefaultQuery("optimize", "true")
	title := c.Query("title")

	movies, err := h.repo.GetAll(c, title, optimize == "true")
	if err != nil {
		print(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Title not found"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (h *MoviesHandler) GetAllExplain(c *gin.Context) {
	optimize := c.DefaultQuery("optimize", "true")
	title := c.Query("title")

	executionPlan, err := h.repo.GetAllExplain(c, title, optimize == "true")
	if err != nil {
		print(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "Error"})
		return
	}

	c.String(http.StatusOK, executionPlan)
}

func (h *MoviesHandler) GetByID(c *gin.Context) {
	tconst := c.Param("tconst")
	optimize := c.DefaultQuery("optimize", "true")

	movies, err := h.repo.GetByID(c, tconst, optimize == "true")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Title not found"})
		return
	}

	c.JSON(http.StatusOK, movies)
}

func (h *MoviesHandler) GetByIDExplain(c *gin.Context) {
	tconst := c.Param("tconst")
	optimize := c.DefaultQuery("optimize", "true")

	executionPlan, err := h.repo.GetByIDExplain(c, tconst, optimize == "true")
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Title not found"})
		return
	}

	c.String(http.StatusOK, executionPlan)
}

func RegisterMoviesRoutes(router *gin.Engine, repo models.MoviesRepository) {
	handler := NewMoviesHandler(repo)

	titleGroup := router.Group("/movies")
	{
		titleGroup.GET("/", handler.GetAll)
		titleGroup.GET("/explain", handler.GetAllExplain)
		titleGroup.GET("/:tconst", handler.GetByID)
		titleGroup.GET("/:tconst/explain", handler.GetByIDExplain)
	}
}
