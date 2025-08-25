package internal

import (
	"net/http"

	"sw-components-jobgenie-restapi/internal/handlers"
	"sw-components-jobgenie-restapi/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Dependencies are injected from main.
func InitServer(userHandler *handlers.UserHandler, resumeHandler *handlers.ResumeHandler) *gin.Engine {
	router := gin.Default()
	router.Use(middleware.CORSMiddleware())

	// Auth endpoints
	router.POST("/api/auth/signup", userHandler.Signup)
	router.POST("/api/auth/login", userHandler.Login)

	// Protected
	protected := router.Group("/api/protected")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to the dashboard!"})
		})
		protected.POST("/resume/upload", resumeHandler.UploadResume)
	}

	return router
}