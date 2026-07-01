package api

import (
	"codebuddy/config"
	"codebuddy/handlers"
	resHandler "codebuddy/results/handlers" 

	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.GET("/health", handlers.HealthCheck)

	v1 := router.Group("/api/v1")
	{
		v1.POST("/ingest" , handlers.IngestRepo(cfg))
		v1.POST("/query" , resHandler.QueryRepo(cfg))
	}
	return router
}