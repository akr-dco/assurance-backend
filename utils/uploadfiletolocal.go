package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func UploadFileToLocal(
	c *gin.Context,
	fileHeader *multipart.FileHeader,
	baseDir string, // contoh: "uploads"
	subFolder string, // contoh: "companies/PT1"
) (string, error) {

	// ===== tanggal folder =====
	dateFolder := time.Now().Format("2006-01-02")

	// uploads/companies/PT1/2025-12-17
	dir := filepath.Join(baseDir, subFolder, dateFolder)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed create dir: %v", err)
	}

	// ===== ambil extension =====
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == "" {
		ext = ".bin" // fallback aman
	}

	// ===== filename final =====
	filename := fmt.Sprintf(
		"%d-%s%s",
		time.Now().Unix(),
		GenerateRandomString(36),
		ext,
	)

	fullPath := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(fileHeader, fullPath); err != nil {
		return "", fmt.Errorf("failed save file: %v", err)
	}

	// ===== path RELATIVE untuk DB =====
	// companies/PT1/2025-12-17/filename.jpg
	relativePath := filepath.ToSlash(
		filepath.Join(subFolder, dateFolder, filename),
	)

	return relativePath, nil
}
