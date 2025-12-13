package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arabella/ai-studio-backend/internal/usecase"
	"go.uber.org/zap"
)

const (
	googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"
	googleUserInfoURL  = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// GoogleAuthConfig holds Google OAuth configuration
type GoogleAuthConfig struct {
	ClientID     string
	ClientSecret string
}

// GoogleTokenInfo represents the response from Google token info endpoint
type GoogleTokenInfo struct {
	Audience      string `json:"aud"`
	UserID        string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Expiry        int64  `json:"exp,string"`
}

// GoogleAuthVerifier implements the GoogleAuthVerifier interface
type GoogleAuthVerifier struct {
	config     GoogleAuthConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewGoogleAuthVerifier creates a new GoogleAuthVerifier
func NewGoogleAuthVerifier(config GoogleAuthConfig, logger *zap.Logger) usecase.GoogleAuthVerifier {
	return &GoogleAuthVerifier{
		config: config,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// VerifyIDToken verifies a Google ID token and returns user info
func (v *GoogleAuthVerifier) VerifyIDToken(ctx context.Context, idToken string) (*usecase.GoogleUserInfo, error) {
	// Verify the token with Google
	url := fmt.Sprintf("%s?id_token=%s", googleTokenInfoURL, idToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		v.logger.Error("Google token verification failed",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(bodyBytes)),
		)
		return nil, fmt.Errorf("token verification failed: %d", resp.StatusCode)
	}

	var tokenInfo GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Verify audience matches our client ID
	if tokenInfo.Audience != v.config.ClientID {
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", v.config.ClientID, tokenInfo.Audience)
	}

	// Check if token is expired
	if time.Now().Unix() > tokenInfo.Expiry {
		return nil, fmt.Errorf("token expired")
	}

	return &usecase.GoogleUserInfo{
		GoogleID:  tokenInfo.UserID,
		Email:     tokenInfo.Email,
		Name:      tokenInfo.Name,
		AvatarURL: tokenInfo.Picture,
	}, nil
}

