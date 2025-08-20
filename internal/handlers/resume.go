package handlers

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sw-components-jobgenie-restapi/internal/services"

	"github.com/gin-gonic/gin"
)

type ResumeHandler struct {
	ResumeService *services.ResumeService
}

func NewResumeHandler(rs *services.ResumeService) *ResumeHandler {
	return &ResumeHandler{ResumeService: rs}
}

// POST /api/protected/resume/upload
func (h *ResumeHandler) UploadResume(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// limits
	maxMB, _ := strconv.ParseInt(os.Getenv("MAX_UPLOAD_SIZE_MB"), 10, 64)
	if maxMB <= 0 {
		maxMB = 5
	}
	maxBytes := maxMB << 20

	// file
	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resume file not found; field name must be 'resume'"})
		return
	}
	if file.Size > maxBytes {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	// type check by extension (reliable)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{".pdf": true, ".docx": true}
	if !allowed[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type; only PDF or DOCX allowed"})
		return
	}

	
	resume, err := h.ResumeService.UploadAndProcessResume(c.Request.Context(), userID, file)
	if err != nil {
		log.Printf("UploadResume error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "resume uploaded and processed successfully",
		"resume_url":     resume.FileURL,
		"extracted_text": resume.TextContent,
	})

}
