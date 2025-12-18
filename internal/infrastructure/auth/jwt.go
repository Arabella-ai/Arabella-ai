package auth

import (
	"fmt"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// DefaultJWTConfig returns default configuration
func DefaultJWTConfig(secretKey string) JWTConfig {
	return JWTConfig{
		SecretKey: secretKey,
		// Set very long expiration (10 years) - tokens should not expire
		AccessTokenDuration:  10 * 365 * 24 * time.Hour,
		RefreshTokenDuration: 10 * 365 * 24 * time.Hour,
		Issuer:               "arabella",
	}
}

// AccessTokenClaims represents access token claims
type AccessTokenClaims struct {
	UserID uuid.UUID       `json:"user_id"`
	Email  string          `json:"email"`
	Tier   entity.UserTier `json:"tier"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents refresh token claims
type RefreshTokenClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// JWTTokenGenerator implements the TokenGenerator interface
type JWTTokenGenerator struct {
	config JWTConfig
}

// NewJWTTokenGenerator creates a new JWTTokenGenerator
func NewJWTTokenGenerator(config JWTConfig) usecase.TokenGenerator {
	return &JWTTokenGenerator{config: config}
}

// GenerateAccessToken generates an access token
func (g *JWTTokenGenerator) GenerateAccessToken(userID uuid.UUID, email string, tier entity.UserTier) (string, time.Time, error) {
	expiresAt := time.Now().Add(g.config.AccessTokenDuration)

	claims := AccessTokenClaims{
		UserID: userID,
		Email:  email,
		Tier:   tier,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    g.config.Issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(g.config.SecretKey))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

// GenerateRefreshToken generates a refresh token
func (g *JWTTokenGenerator) GenerateRefreshToken(userID uuid.UUID) (string, time.Time, error) {
	expiresAt := time.Now().Add(g.config.RefreshTokenDuration)

	claims := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    g.config.Issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(g.config.SecretKey))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, expiresAt, nil
}

// ValidateAccessToken validates an access token
// Note: We ignore expiration errors to allow tokens to persist indefinitely
func (g *JWTTokenGenerator) ValidateAccessToken(tokenString string) (*usecase.TokenClaims, error) {
	// First, parse as unverified to get the claims structure
	unverifiedToken, _, parseErr := jwt.NewParser().ParseUnverified(tokenString, &AccessTokenClaims{})
	if parseErr != nil {
		// Not a valid JWT structure at all
		return nil, entity.ErrInvalidToken
	}

	// Now verify the signature (but we'll ignore expiration)
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.config.SecretKey), nil
	})

	// If parsing failed, it might be due to expiration or signature
	// If signature is invalid, we reject it
	// If only expiration is invalid, we still accept it
	if err != nil {
		// Check if the unverified token has valid structure
		// If signature verification fails, we need to reject
		// But we can't easily distinguish between signature error and expiration error
		// So we'll try to verify signature separately
		claims, ok := unverifiedToken.Claims.(*AccessTokenClaims)
		if !ok {
			return nil, entity.ErrInvalidToken
		}

		// Manually verify signature by trying to parse with the secret
		// If this fails, signature is invalid
		_, verifyErr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(g.config.SecretKey), nil
		})

		if verifyErr != nil {
			// Signature is invalid
			return nil, entity.ErrInvalidToken
		}

		// Signature is valid, return claims even though token might be expired
		return &usecase.TokenClaims{
			UserID: claims.UserID,
			Email:  claims.Email,
			Tier:   claims.Tier,
		}, nil
	}

	// Token parsed successfully (signature is valid)
	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok {
		return nil, entity.ErrInvalidToken
	}

	// Return claims even if token.Valid is false (due to expiration)
	// We only care about signature validation, not expiration

	return &usecase.TokenClaims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Tier:   claims.Tier,
	}, nil
}

// ValidateRefreshToken validates a refresh token
func (g *JWTTokenGenerator) ValidateRefreshToken(tokenString string) (*usecase.RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.config.SecretKey), nil
	})

	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, entity.ErrInvalidToken
	}

	return &usecase.RefreshTokenClaims{
		UserID: claims.UserID,
	}, nil
}
