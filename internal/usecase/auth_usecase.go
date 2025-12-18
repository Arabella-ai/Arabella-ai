package usecase

import (
	"context"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/google/uuid"
)

// AuthTokens represents access and refresh tokens
type AuthTokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// GoogleAuthRequest represents a Google OAuth authentication request
type GoogleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// GoogleUserInfo represents user info from Google
type GoogleUserInfo struct {
	GoogleID  string
	Email     string
	Name      string
	AvatarURL string
}

// TokenGenerator interface for generating JWT tokens
type TokenGenerator interface {
	GenerateAccessToken(userID uuid.UUID, email string, tier entity.UserTier) (string, time.Time, error)
	GenerateRefreshToken(userID uuid.UUID) (string, time.Time, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (*RefreshTokenClaims, error)
}

// TokenClaims represents the claims in an access token
type TokenClaims struct {
	UserID uuid.UUID
	Email  string
	Tier   entity.UserTier
}

// RefreshTokenClaims represents the claims in a refresh token
type RefreshTokenClaims struct {
	UserID uuid.UUID
}

// GoogleAuthVerifier interface for verifying Google tokens
type GoogleAuthVerifier interface {
	VerifyIDToken(ctx context.Context, idToken string) (*GoogleUserInfo, error)
}

// AuthUseCase handles authentication-related business logic
type AuthUseCase struct {
	userRepo       repository.UserRepository
	tokenGenerator TokenGenerator
	googleVerifier GoogleAuthVerifier
}

// NewAuthUseCase creates a new AuthUseCase
func NewAuthUseCase(
	userRepo repository.UserRepository,
	tokenGenerator TokenGenerator,
	googleVerifier GoogleAuthVerifier,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:       userRepo,
		tokenGenerator: tokenGenerator,
		googleVerifier: googleVerifier,
	}
}

// AuthenticateWithGoogle authenticates a user with Google OAuth
func (uc *AuthUseCase) AuthenticateWithGoogle(ctx context.Context, idToken string) (*entity.User, *AuthTokens, error) {
	// Verify the Google ID token
	googleInfo, err := uc.googleVerifier.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, nil, entity.NewDomainError("GOOGLE_AUTH_FAILED", "Failed to verify Google token", err)
	}

	// Try to find existing user by Google ID
	user, err := uc.userRepo.GetByGoogleID(ctx, googleInfo.GoogleID)
	if err != nil && err != entity.ErrUserNotFound {
		return nil, nil, err
	}

	// If no user found by Google ID, try by email
	if user == nil {
		user, err = uc.userRepo.GetByEmail(ctx, googleInfo.Email)
		if err != nil && err != entity.ErrUserNotFound {
			return nil, nil, err
		}
	}

	// Create new user if not found
	if user == nil {
		user = entity.NewUserFromGoogle(
			googleInfo.GoogleID,
			googleInfo.Email,
			googleInfo.Name,
			googleInfo.AvatarURL,
		)
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, nil, err
		}
	} else if user.GoogleID == nil {
		// Link Google account to existing user
		user.GoogleID = &googleInfo.GoogleID
		if googleInfo.AvatarURL != "" {
			user.AvatarURL = &googleInfo.AvatarURL
		}
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return nil, nil, err
		}
	}

	// Generate tokens
	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

// RefreshAccessToken refreshes an access token using a refresh token
func (uc *AuthUseCase) RefreshAccessToken(ctx context.Context, refreshToken string) (*AuthTokens, error) {
	// Validate refresh token
	claims, err := uc.tokenGenerator.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	return uc.generateTokens(user)
}

// ValidateToken validates an access token and returns the user
func (uc *AuthUseCase) ValidateToken(ctx context.Context, accessToken string) (*entity.User, error) {
	// Validate access token
	claims, err := uc.tokenGenerator.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// generateTokens generates access and refresh tokens for a user
func (uc *AuthUseCase) generateTokens(user *entity.User) (*AuthTokens, error) {
	accessToken, expiresAt, err := uc.tokenGenerator.GenerateAccessToken(user.ID, user.Email, user.Tier)
	if err != nil {
		return nil, err
	}

	refreshToken, _, err := uc.tokenGenerator.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
		TokenType:    "Bearer",
	}, nil
}

// AuthenticateAsTestUser creates or gets a test user without Google OAuth
// This is for development/demo purposes
func (uc *AuthUseCase) AuthenticateAsTestUser(ctx context.Context, email, name string) (*entity.User, *AuthTokens, error) {
	// Try to find existing user by email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil && err != entity.ErrUserNotFound {
		return nil, nil, err
	}

	// Create new user if not found
	if user == nil {
		user = entity.NewUser(email, name)
		if err := uc.userRepo.Create(ctx, user); err != nil {
			return nil, nil, err
		}
	}

	// Generate tokens
	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, nil, err
	}

	return user, tokens, nil
}

