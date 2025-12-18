package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authUseCase *usecase.AuthUseCase
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authUseCase *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// GoogleAuthRequest represents the Google OAuth request body
type GoogleAuthRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// GoogleAuthResponse represents the Google OAuth response
type GoogleAuthResponse struct {
	User   *entity.User       `json:"user"`
	Tokens *usecase.AuthTokens `json:"tokens"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// GoogleAuth handles Google OAuth authentication
// @Summary Authenticate with Google
// @Description Authenticate a user using Google OAuth ID token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body GoogleAuthRequest true "Google ID Token"
// @Success 200 {object} GoogleAuthResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/google [post]
func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	var req GoogleAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	user, tokens, err := h.authUseCase.AuthenticateWithGoogle(c.Request.Context(), req.IDToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, GoogleAuthResponse{
		User:   user,
		Tokens: tokens,
	})
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Refresh an access token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh Token"
// @Success 200 {object} usecase.AuthTokens
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	tokens, err := h.authUseCase.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// TestLogin handles test/demo login without Google OAuth
// @Summary Test login (skip Google auth)
// @Description Create or get a test user for development/demo purposes
// @Tags auth
// @Accept json
// @Produce json
// @Param request body TestLoginRequest true "Test Login Request"
// @Success 200 {object} GoogleAuthResponse
// @Failure 400 {object} ErrorResponse
// @Router /auth/test [post]
func (h *AuthHandler) TestLogin(c *gin.Context) {
	var req TestLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// Use email as identifier, or generate a test email
	email := req.Email
	if email == "" {
		email = fmt.Sprintf("test-%d@arabella.test", time.Now().Unix())
	}

	name := req.Name
	if name == "" {
		name = "Test User"
	}

	// Create or get test user
	user, tokens, err := h.authUseCase.AuthenticateAsTestUser(c.Request.Context(), email, name)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, GoogleAuthResponse{
		User:   user,
		Tokens: tokens,
	})
}

// TestLoginRequest represents the test login request body
type TestLoginRequest struct {
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate the current session
// @Tags auth
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// In a stateless JWT system, logout is typically handled client-side
	// by removing the tokens. However, we can implement token blacklisting
	// if needed.
	c.Status(http.StatusNoContent)
}

