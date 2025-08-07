package internal

import (
	"net/http"

	"sw-components-jobgenie-restapi/internal/handlers"
	"sw-components-jobgenie-restapi/internal/middleware"
	"sw-components-jobgenie-restapi/internal/services"

	"github.com/gin-gonic/gin"
)

func InitServer() *gin.Engine {
	// Initialize external services
	services.InitFirebase()
	services.InitSupabase()

	// Create a Gin router with default middleware (logger and recovery)
	router := gin.Default()

	// Setup CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Public routes
	router.POST("/api/auth/signup", handlers.Signup)
	router.POST("/api/auth/login", handlers.Login)

	// Protected routes
	protected := router.Group("/api/protected")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to the dashboard!"})
		})
	}

	return router
}
