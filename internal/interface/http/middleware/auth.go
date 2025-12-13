package middleware

import (
	"net/http"
	"strings"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// AuthorizationHeader is the header key for authorization
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
	// UserIDKey is the context key for user ID
	UserIDKey = "user_id"
	// UserKey is the context key for user
	UserKey = "user"
	// UserTierKey is the context key for user tier
	UserTierKey = "user_tier"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	authUseCase *usecase.AuthUseCase
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(authUseCase *usecase.AuthUseCase) *AuthMiddleware {
	return &AuthMiddleware{
		authUseCase: authUseCase,
	}
}

// RequireAuth middleware requires valid authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Missing authorization token",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		user, err := m.authUseCase.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set(UserIDKey, user.ID)
		c.Set(UserKey, user)
		c.Set(UserTierKey, user.Tier)

		c.Next()
	}
}

// OptionalAuth middleware allows optional authentication
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		user, err := m.authUseCase.ValidateToken(c.Request.Context(), token)
		if err == nil {
			c.Set(UserIDKey, user.ID)
			c.Set(UserKey, user)
			c.Set(UserTierKey, user.Tier)
		}

		c.Next()
	}
}

// RequirePremium middleware requires premium tier or higher
func (m *AuthMiddleware) RequirePremium() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get(UserKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authentication required",
				"code":  "UNAUTHORIZED",
			})
			c.Abort()
			return
		}

		u := user.(*entity.User)
		if !u.IsPremium() {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Premium subscription required",
				"code":  "PREMIUM_REQUIRED",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractToken extracts the JWT token from the request
func extractToken(c *gin.Context) string {
	// Check Authorization header
	authHeader := c.GetHeader(AuthorizationHeader)
	if authHeader != "" && strings.HasPrefix(authHeader, BearerPrefix) {
		return strings.TrimPrefix(authHeader, BearerPrefix)
	}

	// Check query parameter (for WebSocket connections)
	token := c.Query("token")
	if token != "" {
		return token
	}

	return ""
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

// GetUser extracts user from context
func GetUser(c *gin.Context) (*entity.User, bool) {
	user, exists := c.Get(UserKey)
	if !exists {
		return nil, false
	}
	return user.(*entity.User), true
}

// GetUserTier extracts user tier from context
func GetUserTier(c *gin.Context) (entity.UserTier, bool) {
	tier, exists := c.Get(UserTierKey)
	if !exists {
		return "", false
	}
	return tier.(entity.UserTier), true
}

