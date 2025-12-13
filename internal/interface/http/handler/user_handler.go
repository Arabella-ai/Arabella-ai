package handler

import (
	"net/http"

	"github.com/arabella/ai-studio-backend/internal/interface/http/middleware"
	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user endpoints
type UserHandler struct {
	userUseCase *usecase.UserUseCase
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userUseCase *usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

// GetProfile retrieves the authenticated user's profile
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} usecase.UserProfileResponse
// @Failure 401 {object} ErrorResponse
// @Router /user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	profile, err := h.userUseCase.GetProfile(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile updates the authenticated user's profile
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body usecase.UpdateProfileRequest true "Profile Update Request"
// @Success 200 {object} entity.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	var req usecase.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	user, err := h.userUseCase.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetCredits retrieves the authenticated user's credit balance
// @Summary Get user credits
// @Description Get the credit balance of the authenticated user
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]int
// @Failure 401 {object} ErrorResponse
// @Router /user/credits [get]
func (h *UserHandler) GetCredits(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	credits, err := h.userUseCase.GetCredits(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"credits": credits,
	})
}

// UpgradeSubscription upgrades the user's subscription
// @Summary Upgrade subscription
// @Description Upgrade the authenticated user's subscription
// @Tags subscriptions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body usecase.SubscriptionRequest true "Subscription Request"
// @Success 200 {object} entity.User
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /subscriptions [post]
func (h *UserHandler) UpgradeSubscription(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	var req usecase.SubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	user, err := h.userUseCase.UpgradeSubscription(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteAccount deletes the authenticated user's account
// @Summary Delete account
// @Description Permanently delete the authenticated user's account
// @Tags user
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 204
// @Failure 401 {object} ErrorResponse
// @Router /user/account [delete]
func (h *UserHandler) DeleteAccount(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	if err := h.userUseCase.DeleteAccount(c.Request.Context(), userID); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

