package handlers

import (
	"context"
	"net/http"

	"sw-components-jobgenie-restapi/internal/models"
	"sw-components-jobgenie-restapi/internal/services"
	"sw-components-jobgenie-restapi/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthRequest struct {
	IDToken string `json:"idToken"`
}

func Login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx := context.Background()
	token, err := services.VerifyFirebaseToken(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase token"})
		return
	}

	userExists, err := services.UserExists(ctx, token.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}

	if !userExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User does not exist. Please sign up first."})
		return
	}

	jwtToken, err := utils.GenerateJWT(token.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}

	c.SetCookie("token", jwtToken, 3600*24, "/", "localhost", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func Signup(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	ctx := context.Background()
	token, err := services.VerifyFirebaseToken(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Firebase token"})
		return
	}

	userExists, err := services.UserExists(ctx, token.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error checking user existence"})
		return
	}

	if userExists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User already exists. Please log in."})
		return
	}

	authClient, err := services.GetFirebaseAuthClient(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Firebase auth client"})
		return
	}

	userRecord, err := authClient.GetUser(ctx, token.UID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user from Firebase"})
		return
	}

	newUser := models.User{
		FirebaseUID: token.UID,
		Email:       userRecord.Email,
		FullName:    userRecord.DisplayName,
	}

	if err := services.CreateUser(ctx, newUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}
