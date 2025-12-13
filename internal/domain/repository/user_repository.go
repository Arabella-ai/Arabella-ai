package repository

import (
	"context"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	// Create creates a new user
	Create(ctx context.Context, user *entity.User) error

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// GetByGoogleID retrieves a user by Google ID
	GetByGoogleID(ctx context.Context, googleID string) (*entity.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *entity.User) error

	// UpdateCredits updates user credits atomically
	UpdateCredits(ctx context.Context, id uuid.UUID, delta int) error

	// Delete deletes a user (soft delete)
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists users with pagination
	List(ctx context.Context, offset, limit int) ([]*entity.User, int64, error)

	// GetByTier retrieves users by tier
	GetByTier(ctx context.Context, tier entity.UserTier, offset, limit int) ([]*entity.User, int64, error)
}

