package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go/v4"

	"sw-components-jobgenie-restapi/internal/utils"
)

// AuthMiddleware validates the JWT from the cookie
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
			return
		}

		tokenString, err := c.Cookie("jwt")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "No JWT token found in cookie"})
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &utils.MyClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT token"})
			return
		}

		if claims, ok := token.Claims.(*utils.MyClaims); ok && token.Valid {
			// Set the user ID in the context for handlers to use
			c.Set("user_id", claims.UserID)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid JWT claims"})
		}
	}
}