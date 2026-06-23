package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"ragserver/backend/internal/dto"
	"ragserver/backend/internal/model"
)

const CurrentUserKey = "current_user"

type AuthService interface {
	ParseToken(token string) (*model.User, error)
}

func JWTAuth(auth AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
			return
		}
		user, err := auth.ParseToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
			return
		}
		c.Set(CurrentUserKey, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (*model.User, bool) {
	value, ok := c.Get(CurrentUserKey)
	if !ok {
		return nil, false
	}
	user, ok := value.(*model.User)
	return user, ok
}

func RequireTeacher(c *gin.Context) (*model.User, bool) {
	user, ok := CurrentUser(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized"})
		return nil, false
	}
	if user.Role != model.RoleTeacher {
		c.AbortWithStatusJSON(http.StatusForbidden, dto.ErrorResponse{Error: "teacher role required"})
		return nil, false
	}
	return user, true
}
