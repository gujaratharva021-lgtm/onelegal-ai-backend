package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/repositories"
	"legaltech-backend/internal/utils"

	"github.com/google/uuid"
)

var allowedAIDocumentExtensions = map[string]bool{
	".pdf":  true,
	".docx": true,
	".txt":  true,
}

const aiUploadsRoot = "uploads/ai_documents"

type AIDocumentService struct {
	repo      *repositories.AIDocumentAnalysisRepository
	aiService *AIService
}

func NewAIDocumentService(aiService *AIService) *AIDocumentService {
	return &AIDocumentService{
		repo:      repositories.NewAIDocumentAnalysisRepository(),
		aiService: aiService,
	}
}

func (s *AIDocumentService) saveUpload(file *multipart.FileHeader) (path, ext string, err error) {
	ext = strings.ToLower(filepath.Ext(file.Filename))
	if !allowedAIDocumentExtensions[ext] {
		return "", "", fmt.Errorf("unsupported file type: %s (allowed: pdf, docx, txt)", ext)
	}

	if err := os.MkdirAll(aiUploadsRoot, 0o755); err != nil {
		return "", "", err
	}

	storedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(aiUploadsRoot, storedName)

	src, err := file.Open()
	if err != nil {
		return "", "", err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", "", err
	}

	return fullPath, ext, nil
}

func (s *AIDocumentService) UploadAndSummarize(userID uuid.UUID, file *multipart.FileHeader) (*models.AIDocumentAnalysis, error) {
	path, ext, err := s.saveUpload(file)
	if err != nil {
		return nil, err
	}

	text, err := utils.ExtractText(path, ext)
	if err != nil {
		return nil, err
	}

	summary, err := s.aiService.Summarize(text)
	if err != nil {
		return nil, err
	}

	analysis := &models.AIDocumentAnalysis{
		UserID:   userID,
		Type:     models.AIAnalysisTypeSummary,
		FileName: file.Filename,
		FilePath: path,
		Result:   summary,
	}
	if err := s.repo.Create(analysis); err != nil {
		return nil, err
	}
	return analysis, nil
}

func (s *AIDocumentService) UploadAndAnalyzeContract(userID uuid.UUID, file *multipart.FileHeader) (*models.AIDocumentAnalysis, error) {
	path, ext, err := s.saveUpload(file)
	if err != nil {
		return nil, err
	}

	text, err := utils.ExtractText(path, ext)
	if err != nil {
		return nil, err
	}

	result, err := s.aiService.AnalyzeContract(text)
	if err != nil {
		return nil, err
	}

	analysis := &models.AIDocumentAnalysis{
		UserID:   userID,
		Type:     models.AIAnalysisTypeContract,
		FileName: file.Filename,
		FilePath: path,
		Result:   result,
	}
	if err := s.repo.Create(analysis); err != nil {
		return nil, err
	}
	return analysis, nil
}

func (s *AIDocumentService) History(userID uuid.UUID, analysisType string) ([]models.AIDocumentAnalysis, error) {
	return s.repo.ListByUser(userID, analysisType)
}
