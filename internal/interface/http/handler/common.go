package handler

import (
	"errors"
	"net/http"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// handleError handles domain errors and returns appropriate HTTP responses
func handleError(c *gin.Context, err error) {
	// Check for domain errors
	var domainErr *entity.DomainError
	if errors.As(err, &domainErr) {
		c.JSON(getStatusCode(domainErr.Code), ErrorResponse{
			Error:   domainErr.Message,
			Code:    domainErr.Code,
			Details: "",
		})
		return
	}

	// Map known errors to status codes
	switch {
	case errors.Is(err, entity.ErrUserNotFound),
		errors.Is(err, entity.ErrTemplateNotFound),
		errors.Is(err, entity.ErrJobNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error: err.Error(),
			Code:  "NOT_FOUND",
		})

	case errors.Is(err, entity.ErrUnauthorized),
		errors.Is(err, entity.ErrInvalidToken),
		errors.Is(err, entity.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: err.Error(),
			Code:  "UNAUTHORIZED",
		})

	case errors.Is(err, entity.ErrInsufficientCredits),
		errors.Is(err, entity.ErrTemplatePremiumOnly):
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: err.Error(),
			Code:  "FORBIDDEN",
		})

	case errors.Is(err, entity.ErrRateLimitExceeded):
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Error: err.Error(),
			Code:  "RATE_LIMIT_EXCEEDED",
		})

	case errors.Is(err, entity.ErrInvalidInput),
		errors.Is(err, entity.ErrInvalidPrompt),
		errors.Is(err, entity.ErrInvalidParams):
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
			Code:  "BAD_REQUEST",
		})

	case errors.Is(err, entity.ErrProviderUnavailable),
		errors.Is(err, entity.ErrProviderTimeout):
		c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: err.Error(),
			Code:  "SERVICE_UNAVAILABLE",
		})

	case errors.Is(err, entity.ErrJobCannotBeCancelled),
		errors.Is(err, entity.ErrJobAlreadyCompleted),
		errors.Is(err, entity.ErrJobAlreadyCancelled):
		c.JSON(http.StatusConflict, ErrorResponse{
			Error: err.Error(),
			Code:  "CONFLICT",
		})

	default:
		// Log the error and return a generic error
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "An unexpected error occurred",
			Code:  "INTERNAL_ERROR",
		})
	}
}

// getStatusCode maps domain error codes to HTTP status codes
func getStatusCode(code string) int {
	switch code {
	case "NOT_FOUND":
		return http.StatusNotFound
	case "UNAUTHORIZED", "INVALID_TOKEN", "TOKEN_EXPIRED":
		return http.StatusUnauthorized
	case "FORBIDDEN", "PREMIUM_REQUIRED", "INSUFFICIENT_CREDITS":
		return http.StatusForbidden
	case "BAD_REQUEST", "INVALID_INPUT", "VALIDATION_ERROR":
		return http.StatusBadRequest
	case "CONFLICT":
		return http.StatusConflict
	case "RATE_LIMIT_EXCEEDED":
		return http.StatusTooManyRequests
	case "SERVICE_UNAVAILABLE":
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

