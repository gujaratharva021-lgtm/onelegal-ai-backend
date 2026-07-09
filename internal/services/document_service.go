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

	"github.com/google/uuid"
)

const generatedDocumentsRoot = "uploads/documents"
const signaturesRoot = "uploads/signatures"

var allowedSignatureExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
}

var AllowedDocumentExtensions = map[string]bool{
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".jpg":  true,
	".jpeg": true,
	".png":  true,
}

type DocumentService struct {
	repo     *repositories.DocumentRepository
	caseRepo *repositories.CaseRepository
	userRepo *repositories.UserRepository
}

func NewDocumentService() *DocumentService {
	return &DocumentService{
		repo:     repositories.NewDocumentRepository(),
		caseRepo: repositories.NewCaseRepository(),
		userRepo: repositories.NewUserRepository(),
	}
}

const uploadsRoot = "uploads"

func (s *DocumentService) UploadForCase(userID, caseID uuid.UUID, title string, file *multipart.FileHeader) (*models.Document, error) {
	if _, err := s.caseRepo.FindByID(caseID, userID); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !AllowedDocumentExtensions[ext] {
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}

	caseDir := filepath.Join(uploadsRoot, fmt.Sprintf("case_%s", caseID.String()))
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		return nil, err
	}

	storedName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	fullPath := filepath.Join(caseDir, storedName)

	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return nil, err
	}

	if title == "" {
		title = file.Filename
	}

	doc := models.Document{
		UserID:   userID,
		CaseID:   &caseID,
		Title:    title,
		FileName: file.Filename,
		FilePath: fullPath,
		FileURL:  fmt.Sprintf("/api/v1/documents/%s", ""), // placeholder, set after create
		FileType: ext,
		FileSize: size,
	}
	if err := s.repo.Create(&doc); err != nil {
		return nil, err
	}
	doc.FileURL = fmt.Sprintf("/api/v1/documents/%s/download", doc.ID.String())
	_ = s.repo.Update(&doc)
	return &doc, nil
}

func (s *DocumentService) ListForCase(caseID, userID uuid.UUID) ([]models.Document, error) {
	if _, err := s.caseRepo.FindByID(caseID, userID); err != nil {
		return nil, err
	}
	return s.repo.ListByCase(caseID, userID)
}

// Create registers a document. When Content is provided (e.g. an AI-generated
// petition/agreement/notice or OCR result) and no FileURL was given, the
// content is also persisted to disk under uploads/documents/ so it can be
// opened/downloaded like any other document.
func (s *DocumentService) Create(userID uuid.UUID, req models.DocumentRequest) (*models.Document, error) {
	doc := models.Document{
		UserID:       userID,
		CaseID:       req.CaseID,
		ClientID:     req.ClientID,
		Title:        req.Title,
		DocumentType: req.DocumentType,
		Content:      req.Content,
		FileURL:      req.FileURL,
		FileType:     req.FileType,
		FileSize:     req.FileSize,
	}

	if req.Content != "" && req.FileURL == "" {
		if err := os.MkdirAll(generatedDocumentsRoot, 0o755); err != nil {
			return nil, err
		}
		fileName := fmt.Sprintf("%s.txt", uuid.New().String())
		fullPath := filepath.Join(generatedDocumentsRoot, fileName)
		if err := os.WriteFile(fullPath, []byte(req.Content), 0o644); err != nil {
			return nil, err
		}
		doc.FilePath = fullPath
		doc.FileName = req.Title + ".txt"
		doc.FileType = "txt"
		doc.FileSize = int64(len(req.Content))
	}

	if err := s.repo.Create(&doc); err != nil {
		return nil, err
	}
	if doc.FilePath != "" {
		doc.FileURL = fmt.Sprintf("/api/v1/documents/%s/download", doc.ID.String())
		_ = s.repo.Update(&doc)
	}
	return &doc, nil
}

func (s *DocumentService) List(userID uuid.UUID, caseID string) ([]models.Document, error) {
	return s.repo.ListByUser(userID, caseID)
}

func (s *DocumentService) ListForClient(userID, clientID uuid.UUID) ([]models.Document, error) {
	return s.repo.ListByClient(userID, clientID)
}

func (s *DocumentService) Get(id, userID uuid.UUID) (*models.Document, error) {
	return s.repo.FindByID(id, userID)
}

func (s *DocumentService) Delete(id, userID uuid.UUID) error {
	doc, err := s.repo.FindByID(id, userID)
	if err != nil {
		return err
	}
	if doc.FilePath != "" {
		_ = os.Remove(doc.FilePath)
	}
	return s.repo.Delete(id, userID)
}

// SaveSignature stores (or replaces) the current user's signature image.
// This is intentionally simple — a single image file per user overlaid onto
// exported PDFs — not a PKI/certificate-based signing scheme.
func (s *DocumentService) SaveSignature(userID uuid.UUID, file *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedSignatureExtensions[ext] {
		return "", fmt.Errorf("unsupported signature file type: %s (allowed: png, jpg, jpeg)", ext)
	}

	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(signaturesRoot, 0o755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(signaturesRoot, fmt.Sprintf("%s%s", userID.String(), ext))

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Remove a previously saved signature with a different extension so a
	// user can't end up with two stale signature files.
	if user.SignaturePath != "" && user.SignaturePath != fullPath {
		_ = os.Remove(user.SignaturePath)
	}

	user.SignaturePath = fullPath
	if err := s.userRepo.Update(user); err != nil {
		return "", err
	}

	return fullPath, nil
}

// SignaturePath returns the current user's saved signature file path, or an
// empty string if none has been uploaded yet.
func (s *DocumentService) SignaturePath(userID uuid.UUID) (string, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", err
	}
	return user.SignaturePath, nil
}
