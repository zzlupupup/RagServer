package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/middleware"
	"ragserver/backend/internal/model"
	"ragserver/backend/internal/service"
)

type Handler struct {
	Auth        *service.AuthService
	Users       *service.UserService
	KB          *service.KnowledgeBaseService
	Document    *service.DocumentService
	APIKey      *service.APIKeyService
	Search      *service.SearchService
	TempUploads *service.TempUploadService
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.Auth.Register(c.Request.Context(), req)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.Auth.Login(c.Request.Context(), req)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) TempUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "file is required"})
		return
	}
	resp, err := h.TempUploads.Upload(c.Request.Context(), file)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListUsers(c *gin.Context) {
	if _, ok := middleware.RequireTeacher(c); !ok {
		return
	}
	resp, err := h.Users.ListActive(c.Request.Context())
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) CreateKB(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	var req dto.KnowledgeBaseCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.KB.Create(c.Request.Context(), user, req)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListKBs(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	resp, err := h.KB.List(c.Request.Context(), user)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) GetKB(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	resp, err := h.KB.Get(c.Request.Context(), user, id)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) UpdateKB(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	var req dto.KnowledgeBaseUpdateRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.KB.Update(c.Request.Context(), user, id, req)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) DeleteKB(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	write(c, gin.H{"deleted": true}, h.KB.Delete(c.Request.Context(), user, id), http.StatusOK)
}

func (h *Handler) UploadDocument(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "file is required"})
		return
	}
	resp, err := h.Document.UploadMultipart(c.Request.Context(), user, kbID, file)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	page, pageSize := parsePagination(c)
	resp, err := h.Document.ListByKB(c.Request.Context(), user, kbID, page, pageSize)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "document_id")
	if !ok {
		return
	}
	write(c, gin.H{"deleted": true}, h.Document.Delete(c.Request.Context(), user, id), http.StatusOK)
}

func (h *Handler) DocumentStatus(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "document_id")
	if !ok {
		return
	}
	resp, err := h.Document.GetStatus(c.Request.Context(), user, id)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) SearchKB(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		return
	}
	kbID, ok := parseID(c, "kb_id")
	if !ok {
		return
	}
	var req dto.SearchRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.Search.Search(c.Request.Context(), user, kbID, req.Query, req.TopK)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	var req dto.APIKeyCreateRequest
	if !bindJSON(c, &req) {
		return
	}
	resp, err := h.APIKey.Create(c.Request.Context(), teacher, req)
	write(c, resp, err, http.StatusCreated)
}

func (h *Handler) ListAPIKeys(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	page, pageSize := parsePagination(c)
	resp, err := h.APIKey.List(c.Request.Context(), teacher, page, pageSize)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) GetAPIKey(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	resp, err := h.APIKey.Get(c.Request.Context(), teacher, id)
	write(c, resp, err, http.StatusOK)
}

func (h *Handler) RevealAPIKey(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	key, err := h.APIKey.Reveal(c.Request.Context(), teacher, id)
	write(c, dto.RevealAPIKeyResponse{APIKey: key}, err, http.StatusOK)
}

func (h *Handler) DisableAPIKey(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	write(c, gin.H{"disabled": true}, h.APIKey.Disable(c.Request.Context(), teacher, id), http.StatusOK)
}

func (h *Handler) EnableAPIKey(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	write(c, gin.H{"enabled": true}, h.APIKey.Enable(c.Request.Context(), teacher, id), http.StatusOK)
}

func (h *Handler) DeleteAPIKey(c *gin.Context) {
	teacher, ok := middleware.RequireTeacher(c)
	if !ok {
		return
	}
	id, ok := parseID(c, "key_id")
	if !ok {
		return
	}
	write(c, gin.H{"deleted": true}, h.APIKey.Delete(c.Request.Context(), teacher, id), http.StatusOK)
}

func currentUser(c *gin.Context) (*model.User, bool) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
	}
	return user, ok
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

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	return page, pageSize
}

func write(c *gin.Context, payload any, err error, status int) {
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}
	c.JSON(status, payload)
}
