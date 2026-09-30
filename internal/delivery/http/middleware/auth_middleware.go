package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"projectgo/internal/domain"
	"projectgo/pkg/response"
	"projectgo/pkg/token"
)

const (
	ContextUserIDKey = "user_id"
	ContextEmailKey  = "email"
	ContextRoleKey   = "role"
)

func AuthMiddleware(jwtManager *token.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization format. Format must be 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtManager.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token: "+err.Error())
			c.Abort()
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextEmailKey, claims.Email)
		c.Set(ContextRoleKey, domain.Role(claims.Role))

		c.Next()
	}
}

func RequireRole(allowedRoles ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(ContextRoleKey)
		if !exists {
			response.Forbidden(c, "Role information missing from request context")
			c.Abort()
			return
		}

		userRole := roleVal.(domain.Role)
		for _, r := range allowedRoles {
			if userRole == r {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "You do not have permission to access this resource")
		c.Abort()
	}
}

func GetCurrentUserID(c *gin.Context) uuid.UUID {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return uuid.Nil
	}
	return val.(uuid.UUID)
}

func GetCurrentUserRole(c *gin.Context) domain.Role {
	val, exists := c.Get(ContextRoleKey)
	if !exists {
		return ""
	}
	return val.(domain.Role)
}
