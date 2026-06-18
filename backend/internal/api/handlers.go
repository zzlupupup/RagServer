package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/service"
)

type Handler struct {
	KB       *service.KnowledgeBaseService
	Document *service.DocumentService
	APIKey   *service.APIKeyService
	Search   *service.SearchService
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) CreateKB(c *gin.Context) {
	var req dto.KnowledgeBaseCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.KB.Create(c.Request.Context(), req)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListKBs(c *gin.Context) {
	resp, err := h.KB.List(c.Request.Context())
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) GetKB(c *gin.Context) {
	id, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	resp, err := h.KB.Get(c.Request.Context(), id)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) UpdateKB(c *gin.Context) {
	id, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	var req dto.KnowledgeBaseUpdateRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.KB.Update(c.Request.Context(), id, req)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) DeleteKB(c *gin.Context) {
	id, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	write(c, gin.H{"deleted": true}, h.KB.Delete(c.Request.Context(), id), http.StatusOK)
}

func (h *Handler) UploadDocument(c *gin.Context) {
	kbID, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "file is required"})
		return
	}
	resp, err := h.Document.UploadMultipart(c.Request.Context(), kbID, file)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	kbID, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	resp, err := h.Document.ListByKB(c.Request.Context(), kbID)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	id, ok := parseID(c, "document_id")
	if !ok {
		return
	}
	write(c, gin.H{"deleted": true}, h.Document.Delete(c.Request.Context(), id), http.StatusOK)
}

func (h *Handler) DocumentStatus(c *gin.Context) {
	id, ok := parseID(c, "document_id")
	if !ok {
		return
	}
	resp, err := h.Document.GetStatus(c.Request.Context(), id)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) SearchKB(c *gin.Context) {
	kbID, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	var req dto.SearchRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.Search.Search(c.Request.Context(), kbID, req.Query, req.TopK)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	var req dto.APIKeyCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.APIKey.Create(c.Request.Context(), req.Name)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListAPIKeys(c *gin.Context) {
	resp, err := h.APIKey.List(c.Request.Context())
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) GetAPIKey(c *gin.Context) {
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	resp, err := h.APIKey.Get(c.Request.Context(), id)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) RevealAPIKey(c *gin.Context) {
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	key, err := h.APIKey.Reveal(c.Request.Context(), id)
	write(c, dto.RevealAPIKeyResponse{APIKey: key}, err, http.StatusOK)
}

func (h *Handler) DisableAPIKey(c *gin.Context) {
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	write(c, gin.H{"disabled": true}, h.APIKey.Disable(c.Request.Context(), id), http.StatusOK)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	write(c, gin.H{"deleted": true}, h.APIKey.Delete(c.Request.Context(), id), http.StatusOK)
}

func bindJSON(c *gin.Context, out any) bool {
	if err := c.ShouldBindJSON(out); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return false
	}
	return true
}

func parseID(c *gin.Context, name string) (uint64, bool) {
	value := c.Param(name)
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid " + name})
		return 0, false
	}
	return id, true
}

func write(c *gin.Context, payload any, err error, status int) {
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(status, payload)
}
