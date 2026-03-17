package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"viperai/internal/config"
	"viperai/internal/engine"
	"viperai/internal/pkg/utils"
)

type FileService struct{}

func NewFileService() *FileService {
	return &FileService{}
}

func (s *FileService) UploadRAGFile(userID int64, file *multipart.FileHeader) (string, error) {
	if err := validateTextFile(file); err != nil {
		return "", err
	}

	userDir := filepath.Join("uploads", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(userDir, 0755); err != nil {
		log.Printf("Failed to create user directory %s: %v", userDir, err)
		return "", err
	}

	files, err := os.ReadDir(userDir)
	if err == nil {
		for _, f := range files {
			if !f.IsDir() {
				filename := f.Name()
				if err := engine.DeleteRAGIndex(context.Background(), filename); err != nil {
					log.Printf("Failed to delete index for %s: %v", filename, err)
				}
			}
		}
	}

	if err := cleanDirectory(userDir); err != nil {
		log.Printf("Failed to clean user directory %s: %v", userDir, err)
		return "", err
	}

	fileUUID := utils.GenerateUUID()
	ext := filepath.Ext(file.Filename)
	filename := fileUUID + ext
	filePath := filepath.Join(userDir, filename)

	src, err := file.Open()
	if err != nil {
		log.Printf("Failed to open uploaded file: %v", err)
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create destination file %s: %v", filePath, err)
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Printf("Failed to copy file content: %v", err)
		return "", err
	}

	log.Printf("File uploaded successfully: %s", filePath)

	cfg := config.Get().AIModel
	indexer, err := engine.NewRAGIndexer(filename, cfg.EmbeddingModel)
	if err != nil {
		log.Printf("Failed to create RAG indexer: %v", err)
		os.Remove(filePath)
		return "", err
	}

	if err := indexer.IndexFile(context.Background(), filePath); err != nil {
		log.Printf("Failed to index file: %v", err)
		os.Remove(filePath)
		engine.DeleteRAGIndex(context.Background(), filename)
		return "", err
	}

	log.Printf("File indexed successfully: %s", filename)
	return filePath, nil
}

func validateTextFile(file *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".md" && ext != ".txt" {
		return ErrInvalidFileType
	}
	return nil
}

func cleanDirectory(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return err
			}
		}
	}
	return nil
}

var ErrInvalidFileType = NewServiceError(2001, "Invalid file type, only .md and .txt files are allowed")
