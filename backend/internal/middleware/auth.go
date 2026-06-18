package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/dto"
)

func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if adminToken == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "admin token is not configured"})
			return
		}
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token != adminToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
			return
		}
		c.Next()
	}
}
