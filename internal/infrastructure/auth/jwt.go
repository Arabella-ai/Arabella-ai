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
		SecretKey:            secretKey,
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
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
func (g *JWTTokenGenerator) ValidateAccessToken(tokenString string) (*usecase.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(g.config.SecretKey), nil
	})

	if err != nil {
		return nil, entity.ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, entity.ErrInvalidToken
	}

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

