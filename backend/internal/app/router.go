package app

import (
	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/api"
	"ragserver/backend/internal/middleware"
)

func registerRoutes(r *gin.Engine, h *api.Handler, adminToken string) {
	v1 := r.Group("/api/v1")
	v1.GET("/health", h.Health)

	admin := v1.Group("")
	admin.Use(middleware.AdminAuth(adminToken))
	admin.POST("/kbs", h.CreateKB)
	admin.GET("/kbs", h.ListKBs)
	admin.GET("/kbs/:kb_id", h.GetKB)
	admin.PATCH("/kbs/:kb_id", h.UpdateKB)
	admin.DELETE("/kbs/:kb_id", h.DeleteKB)
	admin.POST("/kbs/:kb_id/documents/upload", h.UploadDocument)
	admin.GET("/kbs/:kb_id/documents", h.ListDocuments)
	admin.POST("/kbs/:kb_id/search", h.SearchKB)
	admin.DELETE("/documents/:document_id", h.DeleteDocument)
	admin.GET("/documents/:document_id/status", h.DocumentStatus)
	admin.POST("/api-keys", h.CreateAPIKey)
	admin.GET("/api-keys", h.ListAPIKeys)
	admin.GET("/api-keys/:key_id", h.GetAPIKey)
	admin.POST("/api-keys/:key_id/reveal", h.RevealAPIKey)
	admin.POST("/api-keys/:key_id/disable", h.DisableAPIKey)
	admin.DELETE("/api-keys/:key_id", h.DeleteAPIKey)
}
