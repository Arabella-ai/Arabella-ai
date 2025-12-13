package entity

import (
	"time"

	"github.com/google/uuid"
)

// UserTier represents the subscription tier of a user
type UserTier string

const (
	UserTierFree    UserTier = "free"
	UserTierPremium UserTier = "premium"
	UserTierPro     UserTier = "pro"
)

// User represents a platform user
type User struct {
	ID                    uuid.UUID  `json:"id"`
	Email                 string     `json:"email"`
	GoogleID              *string    `json:"google_id,omitempty"`
	Name                  string     `json:"name"`
	AvatarURL             *string    `json:"avatar_url,omitempty"`
	Credits               int        `json:"credits"`
	Tier                  UserTier   `json:"tier"`
	SubscriptionExpiresAt *time.Time `json:"subscription_expires_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// NewUser creates a new user with default values
func NewUser(email, name string) *User {
	now := time.Now()
	return &User{
		ID:        uuid.New(),
		Email:     email,
		Name:      name,
		Credits:   5, // Default free credits
		Tier:      UserTierFree,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewUserFromGoogle creates a new user from Google OAuth data
func NewUserFromGoogle(googleID, email, name, avatarURL string) *User {
	user := NewUser(email, name)
	user.GoogleID = &googleID
	if avatarURL != "" {
		user.AvatarURL = &avatarURL
	}
	return user
}

// IsPremium checks if the user has premium access
func (u *User) IsPremium() bool {
	if u.Tier == UserTierFree {
		return false
	}
	if u.SubscriptionExpiresAt != nil && u.SubscriptionExpiresAt.Before(time.Now()) {
		return false
	}
	return true
}

// HasSufficientCredits checks if user has enough credits
func (u *User) HasSufficientCredits(required int) bool {
	return u.Credits >= required
}

// DeductCredits removes credits from user account
func (u *User) DeductCredits(amount int) error {
	if !u.HasSufficientCredits(amount) {
		return ErrInsufficientCredits
	}
	u.Credits -= amount
	u.UpdatedAt = time.Now()
	return nil
}

// AddCredits adds credits to user account
func (u *User) AddCredits(amount int) {
	u.Credits += amount
	u.UpdatedAt = time.Now()
}

// UpgradeTier upgrades user to a new tier
func (u *User) UpgradeTier(tier UserTier, expiresAt time.Time) {
	u.Tier = tier
	u.SubscriptionExpiresAt = &expiresAt
	u.UpdatedAt = time.Now()
}

