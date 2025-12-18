package auth

import (
	"context"
	"fmt"

	"github.com/arabella/ai-studio-backend/internal/usecase"
	"google.golang.org/api/idtoken"
	"go.uber.org/zap"
)

// GoogleAuthConfig holds Google OAuth configuration
type GoogleAuthConfig struct {
	ClientID     string
	ClientSecret string
}

// GoogleAuthVerifier implements the GoogleAuthVerifier interface
type GoogleAuthVerifier struct {
	config GoogleAuthConfig
	logger *zap.Logger
}

// NewGoogleAuthVerifier creates a new GoogleAuthVerifier
func NewGoogleAuthVerifier(config GoogleAuthConfig, logger *zap.Logger) usecase.GoogleAuthVerifier {
	return &GoogleAuthVerifier{
		config: config,
		logger: logger,
	}
}

// VerifyIDToken verifies a Google ID token and returns user info
func (v *GoogleAuthVerifier) VerifyIDToken(ctx context.Context, idToken string) (*usecase.GoogleUserInfo, error) {
	// Verify the token using Google's official library
	payload, err := idtoken.Validate(ctx, idToken, v.config.ClientID)
	if err != nil {
		v.logger.Error("Google token verification failed",
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to verify Google token: %w", err)
	}

	// Extract user information from the token payload
	googleID, ok := payload.Claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing sub claim")
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	return &usecase.GoogleUserInfo{
		GoogleID:  googleID,
		Email:     email,
		Name:      name,
		AvatarURL: picture,
	}, nil
}

