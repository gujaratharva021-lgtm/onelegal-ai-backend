package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"legaltech-backend/internal/ai"
	"legaltech-backend/internal/utils"

	"github.com/google/uuid"
)

var allowedOCRExtensions = map[string]bool{
	".pdf":  true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

var imageMimeTypes = map[string]string{
	".png":  "image/png",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
}

const ocrUploadsRoot = "uploads/ocr"
const ocrRequestTimeout = 30 * time.Second

type OCRService struct{}

func NewOCRService() *OCRService {
	return &OCRService{}
}

// Extract pulls text out of an uploaded PDF or image. PDFs use the existing
// text-layer extractor; images are sent to a Groq vision model since there is
// no reliable pure-Go OCR engine — this keeps the app on a single AI provider
// (Groq) rather than adding a separate OCR dependency.
func (s *OCRService) Extract(file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedOCRExtensions[ext] {
		return "", fmt.Errorf("unsupported file type: %s (allowed: pdf, png, jpg, jpeg)", ext)
	}

	if err := os.MkdirAll(ocrUploadsRoot, 0o755); err != nil {
		return "", err
	}
	tempPath := filepath.Join(ocrUploadsRoot, fmt.Sprintf("%s%s", uuid.New().String(), ext))
	defer os.Remove(tempPath)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if _, err := io.Copy(dst, io.TeeReader(src, &buf)); err != nil {
		dst.Close()
		return "", err
	}
	dst.Close()
	data := buf.Bytes()

	if ext == ".pdf" {
		return utils.ExtractText(tempPath, ext)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ocrRequestTimeout)
	defer cancel()

	encoded := base64.StdEncoding.EncodeToString(data)
	return ai.ExtractImageText(ctx, encoded, imageMimeTypes[ext])
}
