package entity

import "errors"

// Domain errors
var (
	// User errors
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInsufficientCredits = errors.New("insufficient credits")
	ErrInvalidUserTier     = errors.New("invalid user tier")

	// Template errors
	ErrTemplateNotFound     = errors.New("template not found")
	ErrTemplateNotActive    = errors.New("template is not active")
	ErrTemplatePremiumOnly  = errors.New("template is premium only")

	// Video job errors
	ErrJobNotFound          = errors.New("video job not found")
	ErrJobAlreadyCompleted  = errors.New("video job already completed")
	ErrJobCannotBeCancelled = errors.New("video job cannot be cancelled")
	ErrJobAlreadyCancelled  = errors.New("video job already cancelled")

	// Provider errors
	ErrProviderUnavailable  = errors.New("AI provider unavailable")
	ErrProviderRateLimited  = errors.New("AI provider rate limited")
	ErrProviderTimeout      = errors.New("AI provider timeout")
	ErrGenerationFailed     = errors.New("video generation failed")

	// Validation errors
	ErrInvalidInput         = errors.New("invalid input")
	ErrInvalidPrompt        = errors.New("invalid prompt")
	ErrInvalidParams        = errors.New("invalid video parameters")
	ErrPromptTooLong        = errors.New("prompt exceeds maximum length")

	// Authentication errors
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrUnauthorized         = errors.New("unauthorized")

	// Rate limiting errors
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")

	// Storage errors
	ErrStorageUploadFailed  = errors.New("storage upload failed")
	ErrStorageDownloadFailed = errors.New("storage download failed")
)

// DomainError represents a domain-level error with additional context
type DomainError struct {
	Code    string
	Message string
	Err     error
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

