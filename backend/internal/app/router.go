package app

import (
	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/api"
	"ragserver/backend/internal/middleware"
)

func registerRoutes(r *gin.Engine, h *api.Handler, auth middleware.AuthService) {
	v1 := r.Group("/api/v1")
	v1.GET("/health", h.Health)
	v1.POST("/auth/register", h.Register)
	v1.POST("/auth/login", h.Login)
	v1.POST("/mcp/files/upload", h.TempUpload)

	protected := v1.Group("")
	protected.Use(middleware.JWTAuth(auth))
	protected.POST("/kbs", h.CreateKB)
	protected.GET("/kbs", h.ListKBs)
	protected.GET("/kbs/:kb_id", h.GetKB)
	protected.PATCH("/kbs/:kb_id", h.UpdateKB)
	protected.DELETE("/kbs/:kb_id", h.DeleteKB)
	protected.POST("/kbs/:kb_id/documents/upload", h.UploadDocument)
	protected.GET("/kbs/:kb_id/documents", h.ListDocuments)
	protected.POST("/kbs/:kb_id/search", h.SearchKB)
	protected.DELETE("/documents/:document_id", h.DeleteDocument)
	protected.GET("/documents/:document_id/status", h.DocumentStatus)
	protected.GET("/users", h.ListUsers)
	protected.POST("/api-keys", h.CreateAPIKey)
	protected.GET("/api-keys", h.ListAPIKeys)
	protected.GET("/api-keys/:key_id", h.GetAPIKey)
	protected.POST("/api-keys/:key_id/reveal", h.RevealAPIKey)
	protected.POST("/api-keys/:key_id/disable", h.DisableAPIKey)
	protected.POST("/api-keys/:key_id/enable", h.EnableAPIKey)
	protected.DELETE("/api-keys/:key_id", h.DeleteAPIKey)
}
