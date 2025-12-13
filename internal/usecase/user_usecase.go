package usecase

import (
	"context"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/google/uuid"
)

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	User            *entity.User `json:"user"`
	ActiveJobsCount int          `json:"active_jobs_count"`
	TotalJobs       int64        `json:"total_jobs"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

// SubscriptionRequest represents a subscription upgrade request
type SubscriptionRequest struct {
	Plan     string `json:"plan" binding:"required,oneof=premium pro"`
	Duration int    `json:"duration" binding:"required,oneof=1 3 6 12"` // months
}

// UserUseCase handles user-related business logic
type UserUseCase struct {
	userRepo repository.UserRepository
	jobRepo  repository.VideoJobRepository
}

// NewUserUseCase creates a new UserUseCase
func NewUserUseCase(
	userRepo repository.UserRepository,
	jobRepo repository.VideoJobRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo: userRepo,
		jobRepo:  jobRepo,
	}
}

// GetProfile retrieves the user profile with additional stats
func (uc *UserUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*UserProfileResponse, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get active jobs count
	activeJobsCount, err := uc.jobRepo.GetActiveJobsCount(ctx, userID)
	if err != nil {
		activeJobsCount = 0
	}

	// Get total jobs count
	filter := repository.VideoJobFilter{UserID: &userID}
	_, totalJobs, err := uc.jobRepo.List(ctx, filter, 0, 1)
	if err != nil {
		totalJobs = 0
	}

	return &UserProfileResponse{
		User:            user,
		ActiveJobsCount: activeJobsCount,
		TotalJobs:       totalJobs,
	}, nil
}

// UpdateProfile updates the user profile
func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.AvatarURL != "" {
		user.AvatarURL = &req.AvatarURL
	}

	user.UpdatedAt = time.Now()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// UpgradeSubscription upgrades the user subscription
func (uc *UserUseCase) UpgradeSubscription(ctx context.Context, userID uuid.UUID, req SubscriptionRequest) (*entity.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Determine tier
	var tier entity.UserTier
	switch req.Plan {
	case "premium":
		tier = entity.UserTierPremium
	case "pro":
		tier = entity.UserTierPro
	default:
		return nil, entity.ErrInvalidInput
	}

	// Calculate expiration
	expiresAt := time.Now().AddDate(0, req.Duration, 0)

	// If already subscribed, extend from current expiration
	if user.SubscriptionExpiresAt != nil && user.SubscriptionExpiresAt.After(time.Now()) {
		expiresAt = user.SubscriptionExpiresAt.AddDate(0, req.Duration, 0)
	}

	// Upgrade tier
	user.UpgradeTier(tier, expiresAt)

	// Add bonus credits based on plan
	bonusCredits := 0
	switch tier {
	case entity.UserTierPremium:
		bonusCredits = 50 * req.Duration
	case entity.UserTierPro:
		bonusCredits = 200 * req.Duration
	}
	user.AddCredits(bonusCredits)

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetCredits retrieves the user's current credit balance
func (uc *UserUseCase) GetCredits(ctx context.Context, userID uuid.UUID) (int, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return 0, err
	}

	return user.Credits, nil
}

// AddCredits adds credits to the user's account
func (uc *UserUseCase) AddCredits(ctx context.Context, userID uuid.UUID, amount int) error {
	return uc.userRepo.UpdateCredits(ctx, userID, amount)
}

// DeductCredits deducts credits from the user's account
func (uc *UserUseCase) DeductCredits(ctx context.Context, userID uuid.UUID, amount int) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.HasSufficientCredits(amount) {
		return entity.ErrInsufficientCredits
	}

	return uc.userRepo.UpdateCredits(ctx, userID, -amount)
}

// DeleteAccount deletes the user account
func (uc *UserUseCase) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	return uc.userRepo.Delete(ctx, userID)
}

