package entity

import (
	"time"

	"github.com/google/uuid"
)

// TemplateCategory represents the category of a video template
type TemplateCategory string

const (
	TemplateCategoryCyberpunk      TemplateCategory = "cyberpunk_intro"
	TemplateCategoryProductShowcase TemplateCategory = "product_showcase"
	TemplateCategoryDailyVlog       TemplateCategory = "daily_vlog"
	TemplateCategoryTechReview      TemplateCategory = "tech_review"
	TemplateCategoryNature          TemplateCategory = "nature"
	TemplateCategoryAbstract        TemplateCategory = "abstract"
	TemplateCategoryBusiness        TemplateCategory = "business"
	TemplateCategoryEducation       TemplateCategory = "education"
)

// VideoResolution represents video output resolution
type VideoResolution string

const (
	Resolution720p  VideoResolution = "720p"
	Resolution1080p VideoResolution = "1080p"
	Resolution4K    VideoResolution = "4k"
)

// AspectRatio represents video aspect ratio
type AspectRatio string

const (
	AspectRatio16x9 AspectRatio = "16:9"
	AspectRatio9x16 AspectRatio = "9:16"
	AspectRatio1x1  AspectRatio = "1:1"
	AspectRatio4x3  AspectRatio = "4:3"
)

// VideoParams contains configurable video generation parameters
type VideoParams struct {
	Duration       int             `json:"duration"`        // Duration in seconds
	Resolution     VideoResolution `json:"resolution"`      // Output resolution
	AspectRatio    AspectRatio     `json:"aspect_ratio"`    // Aspect ratio
	FPS            int             `json:"fps"`             // Frames per second
	Style          string          `json:"style,omitempty"` // Visual style modifier
	NegativePrompt string          `json:"negative_prompt,omitempty"`
}

// DefaultVideoParams returns default video parameters
func DefaultVideoParams() VideoParams {
	return VideoParams{
		Duration:    15,
		Resolution:  Resolution1080p,
		AspectRatio: AspectRatio16x9,
		FPS:         30,
	}
}

// Template represents a pre-built AI video template
type Template struct {
	ID                uuid.UUID        `json:"id"`
	Name              string           `json:"name"`
	Category          TemplateCategory `json:"category"`
	Description       string           `json:"description"`
	ThumbnailURL      string           `json:"thumbnail_url"`
	PreviewVideoURL   *string          `json:"preview_video_url,omitempty"`
	BasePrompt        string           `json:"base_prompt"`
	DefaultParams     VideoParams      `json:"default_params"`
	CreditCost        int              `json:"credit_cost"`
	EstimatedTime     time.Duration    `json:"estimated_time"` // In seconds
	IsPremium         bool             `json:"is_premium"`
	IsActive          bool             `json:"is_active"`
	PreferredProvider *string          `json:"preferred_provider,omitempty"`
	Tags              []string         `json:"tags"`
	UsageCount        int64            `json:"usage_count"`
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

// NewTemplate creates a new template
func NewTemplate(name string, category TemplateCategory, description, basePrompt string) *Template {
	now := time.Now()
	return &Template{
		ID:            uuid.New(),
		Name:          name,
		Category:      category,
		Description:   description,
		BasePrompt:    basePrompt,
		DefaultParams: DefaultVideoParams(),
		CreditCost:    1,
		EstimatedTime: 60 * time.Second,
		IsPremium:     false,
		IsActive:      true,
		Tags:          []string{},
		UsageCount:    0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// CanBeUsedBy checks if a template can be used by a user
func (t *Template) CanBeUsedBy(user *User) bool {
	if !t.IsActive {
		return false
	}
	if t.IsPremium && !user.IsPremium() {
		return false
	}
	return user.HasSufficientCredits(t.CreditCost)
}

// IncrementUsage increments the usage count
func (t *Template) IncrementUsage() {
	t.UsageCount++
	t.UpdatedAt = time.Now()
}

