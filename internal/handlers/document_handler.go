package handlers

import (
	"net/http"

	"legaltech-backend/internal/models"
	"legaltech-backend/internal/services"
	"legaltech-backend/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DocumentHandler struct {
	service *services.DocumentService
}

func NewDocumentHandler() *DocumentHandler {
	return &DocumentHandler{service: services.NewDocumentService()}
}

func (h *DocumentHandler) Create(c *gin.Context) {
	userID, _ := currentUserID(c)
	var req models.DocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	doc, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to register document")
		return
	}
	response.Success(c, http.StatusCreated, "Document registered", doc)
}

func (h *DocumentHandler) List(c *gin.Context) {
	userID, _ := currentUserID(c)
	docs, err := h.service.List(userID, c.Query("case_id"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to fetch documents")
		return
	}
	response.Success(c, http.StatusOK, "Documents fetched", docs)
}

func (h *DocumentHandler) Get(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid document id")
		return
	}
	doc, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Document not found")
		return
	}
	response.Success(c, http.StatusOK, "Document fetched", doc)
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	userID, _ := currentUserID(c)
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided")
		return
	}
	title := c.PostForm("title")
	doc, err := h.service.UploadForCase(userID, caseID, title, fileHeader)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "Document uploaded", doc)
}

func (h *DocumentHandler) ListForCase(c *gin.Context) {
	userID, _ := currentUserID(c)
	caseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid case id")
		return
	}
	docs, err := h.service.ListForCase(caseID, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Case not found")
		return
	}
	response.Success(c, http.StatusOK, "Documents fetched", docs)
}

func (h *DocumentHandler) Download(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid document id")
		return
	}
	doc, err := h.service.Get(id, userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Document not found")
		return
	}
	if doc.FilePath == "" {
		response.Error(c, http.StatusNotFound, "File not available for download")
		return
	}
	c.FileAttachment(doc.FilePath, doc.FileName)
}

func (h *DocumentHandler) Sign(c *gin.Context) {
	userID, _ := currentUserID(c)
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "No file provided")
		return
	}
	if _, err := h.service.SaveSignature(userID, file); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "Signature saved", gin.H{"signature_url": "/api/v1/documents/signature"})
}

func (h *DocumentHandler) GetSignature(c *gin.Context) {
	userID, _ := currentUserID(c)
	path, err := h.service.SignaturePath(userID)
	if err != nil || path == "" {
		response.Error(c, http.StatusNotFound, "No signature uploaded yet")
		return
	}
	c.File(path)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	userID, _ := currentUserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid document id")
		return
	}
	if err := h.service.Delete(id, userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to delete document")
		return
	}
	response.Success(c, http.StatusOK, "Document deleted", nil)
}
