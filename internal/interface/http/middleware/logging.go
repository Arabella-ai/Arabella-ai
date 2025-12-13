package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
	// RequestIDHeader is the header key for request ID
	RequestIDHeader = "X-Request-ID"
)

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
	logger *zap.Logger
}

// NewLoggingMiddleware creates a new LoggingMiddleware
func NewLoggingMiddleware(logger *zap.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// RequestID middleware adds a unique request ID to each request
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set(RequestIDKey, requestID)
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// Logger middleware logs request details
func (m *LoggingMiddleware) Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Get request ID
		requestID, _ := c.Get(RequestIDKey)

		// Get user ID if available
		var userID string
		if uid, exists := c.Get(UserIDKey); exists {
			userID = uid.(uuid.UUID).String()
		}

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID.(string)),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Log based on status code
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				fields = append(fields, zap.String("error", e.Error()))
			}
			m.logger.Error("Request completed with errors", fields...)
		} else if statusCode >= 500 {
			m.logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			m.logger.Warn("Client error", fields...)
		} else {
			m.logger.Info("Request completed", fields...)
		}
	}
}

// Recovery middleware recovers from panics
func (m *LoggingMiddleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get(RequestIDKey)

				m.logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("request_id", requestID.(string)),
					zap.String("path", c.Request.URL.Path),
				)

				c.AbortWithStatusJSON(500, gin.H{
					"error":      "Internal server error",
					"code":       "INTERNAL_ERROR",
					"request_id": requestID,
				})
			}
		}()

		c.Next()
	}
}

