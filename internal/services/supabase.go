package services

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	storage "github.com/supabase-community/storage-go"
)

type Client struct {
	DB      *sql.DB
	Storage *storage.Client
}

// SUPABASE_CONN_STRING, SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY.
func NewSupabaseClient() (*Client, error) {
	connStr := os.Getenv("SUPABASE_CONN_STRING")
	if connStr == "" {
		return nil, fmt.Errorf("SUPABASE_CONN_STRING not set")
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err = db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	sbURL := os.Getenv("SUPABASE_URL")
	sbSvcKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if sbURL == "" || sbSvcKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL or SUPABASE_SERVICE_ROLE_KEY not set")
	}
	st := storage.NewClient(sbURL, sbSvcKey, nil)

	log.Println("Supabase DB + Storage initialized")
	return &Client{DB: db, Storage: st}, nil
}

func (c *Client) UploadResumeFile(ctx context.Context, bucketName, filePath string, file io.Reader) (string, error) {
	// 1. Create a temporary file
	tmpFile, err := os.CreateTemp("", "upload-*")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// 2. Copy uploaded content into temp file
	if _, err := io.Copy(tmpFile, file); err != nil {
		return "", fmt.Errorf("write temp file: %w", err)
	}

	// 3. Rewind file pointer
	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("rewind file: %w", err)
	}

	// 4. Upload using tmpFile (io.Reader)
	// log upload attempt details
	if fi, err := tmpFile.Stat(); err == nil {
		log.Printf("Uploading file to storage: bucket=%s path=%s tmp=%s size=%d", bucketName, filePath, tmpFile.Name(), fi.Size())
	} else {
		log.Printf("Uploading file to storage: bucket=%s path=%s tmp=%s (stat failed: %v)", bucketName, filePath, tmpFile.Name(), err)
	}

	_, err = c.Storage.UploadFile(bucketName, filePath, tmpFile, storage.FileOptions{})
	if err != nil {
		// log the raw error with %#v to include type information when string is empty
		log.Printf("Storage upload error (raw): %#v", err)

		// Attempt a fallback HTTP PUT to the Supabase Storage REST endpoint to get a clearer response
		if _, seekErr := tmpFile.Seek(0, io.SeekStart); seekErr != nil {
			log.Printf("fallback seek failed: %v", seekErr)
			return "", fmt.Errorf("storage upload: %v", err)
		}

		supabaseURL := os.Getenv("SUPABASE_URL")
		svcKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		if supabaseURL == "" || svcKey == "" {
			return "", fmt.Errorf("storage upload: %v", err)
		}

		uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", supabaseURL, bucketName, filePath)
		req, rerr := http.NewRequestWithContext(ctx, "PUT", uploadURL, tmpFile)
		if rerr != nil {
			log.Printf("fallback new request failed: %v", rerr)
			return "", fmt.Errorf("storage upload: %v", err)
		}
		req.Header.Set("Authorization", "Bearer "+svcKey)
		req.Header.Set("Content-Type", "application/octet-stream")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, derr := client.Do(req)
		if derr != nil {
			log.Printf("fallback http error: %v", derr)
			return "", fmt.Errorf("storage upload: %v", derr)
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("fallback upload response: status=%d body=%s", resp.StatusCode, string(body))

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return "", fmt.Errorf("storage upload: fallback HTTP %d: %s", resp.StatusCode, string(body))
		}
	}

	// 5. Construct public URL
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s",
		os.Getenv("SUPABASE_URL"), bucketName, filePath)

	return publicURL, nil
}
