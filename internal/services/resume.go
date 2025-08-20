package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"sw-components-jobgenie-restapi/internal/models"

	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unipdf/v3/extractor"
	"github.com/unidoc/unipdf/v3/model"
)

type ResumeService struct {
	Supabase *Client
}

func NewResumeService(sc *Client) *ResumeService {
	return &ResumeService{Supabase: sc}
}

// UploadAndProcessResume uploads the file, extracts text, and persists metadata.
func (s *ResumeService) UploadAndProcessResume(ctx context.Context, userFirebaseUID string, file *multipart.FileHeader) (*models.Resume, error) {
	var idStr string
	q := `SELECT id FROM users WHERE firebase_uid = $1 LIMIT 1;`
	err := s.Supabase.DB.QueryRowContext(ctx, q, userFirebaseUID).Scan(&idStr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve user id for firebase_uid %s: %w", userFirebaseUID, err)
	}
	userUUID, err := uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user id stored in DB: %w", err)
	}

	url, err := s.uploadFileToStorage(ctx, file, idStr)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}

	extracted, err := s.extractTextFromFile(file)
	if err != nil {
		return nil, fmt.Errorf("text extraction failed: %w", err)
	}

	resume := &models.Resume{
		UserID:      userUUID,
		FileURL:     url,
		TextContent: extracted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.saveResumeToDB(ctx, resume); err != nil {
		return nil, fmt.Errorf("db save failed: %w", err)
	}

	return resume, nil
}

func (s *ResumeService) uploadFileToStorage(ctx context.Context, file *multipart.FileHeader, userID string) (string, error) {
	bucket := os.Getenv("SUPABASE_STORAGE_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("SUPABASE_STORAGE_BUCKET not set")
	}

	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	path := fmt.Sprintf("resumes/%s/%d_%s", userID, time.Now().Unix(), sanitizeFilename(file.Filename))
	return s.Supabase.UploadResumeFile(ctx, bucket, path, bytes.NewReader(buf))
}

func (s *ResumeService) extractTextFromFile(file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))

	opened, err := file.Open()
	if err != nil {
		return "", err
	}
	defer opened.Close()

	switch ext {
	case ".pdf":
		// --- PDF Extraction with unipdf ---
		reader, err := model.NewPdfReader(opened)
		if err != nil {
			return "", err
		}
		if enc, _ := reader.IsEncrypted(); enc {
			return "", fmt.Errorf("PDF is encrypted; cannot parse")
		}
		numPages, err := reader.GetNumPages()
		if err != nil {
			return "", err
		}

		var out strings.Builder
		for i := 1; i <= numPages; i++ {
			page, err := reader.GetPage(i)
			if err != nil {
				return "", err
			}
			ex, err := extractor.New(page)
			if err != nil {
				return "", err
			}
			txt, err := ex.ExtractText()
			if err != nil {
				return "", err
			}
			out.WriteString(txt)
			out.WriteByte('\n')
		}
		return out.String(), nil

	case ".docx":
		// --- DOCX Extraction with unioffice ---
		data, err := io.ReadAll(opened)
		if err != nil {
			return "", err
		}
		doc, err := document.Read(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return "", err
		}
		defer doc.Close()

		var out strings.Builder
		paras := doc.Paragraphs()
		for _, p := range paras {
			for _, r := range p.Runs() {
				out.WriteString(r.Text())
			}
			out.WriteByte('\n')
		}
		return out.String(), nil

	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

func (s *ResumeService) saveResumeToDB(ctx context.Context, resume *models.Resume) error {
	const q = `
		INSERT INTO resumes (user_id, file_url, text_content, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id;
	`
	return s.Supabase.DB.
		QueryRowContext(ctx, q,
			resume.UserID,
			resume.FileURL,
			resume.TextContent,
			resume.CreatedAt,
			resume.UpdatedAt,
		).
		Scan(&resume.ID)
}

func sanitizeFilename(name string) string {
	return strings.ReplaceAll(name, " ", "_")
}
