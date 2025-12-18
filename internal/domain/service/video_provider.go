package service

import (
	"context"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
)

// GenerationRequest represents a video generation request to an AI provider
type GenerationRequest struct {
	JobID        string
	Prompt       string
	Params       entity.VideoParams
	TemplateID   string
	BasePrompt   string
	ThumbnailURL string // Template thumbnail URL for image-to-video
	UserTier     entity.UserTier
}

// ProviderCapabilities describes what a provider can do
type ProviderCapabilities struct {
	Name            string
	MaxDuration     int // Maximum video duration in seconds
	MaxResolution   entity.VideoResolution
	SupportedRatios []entity.AspectRatio
	EstimatedTime   int    // Estimated time per second of video
	QualityTier     string // budget, standard, premium
	SupportsStyles  bool
	CostPerSecond   float64
}

// ProviderHealth represents the health status of a provider
type ProviderHealth struct {
	IsHealthy    bool
	QueueDepth   int
	ResponseTime int64 // in milliseconds
	ErrorRate    float64
	LastChecked  int64 // unix timestamp
}

// VideoProvider defines the contract for AI video generation providers
type VideoProvider interface {
	// GetName returns the provider name
	GetName() entity.AIProvider

	// GenerateVideo initiates video generation
	GenerateVideo(ctx context.Context, req GenerationRequest) (*entity.GenerationResult, error)

	// GetProgress retrieves generation progress
	GetProgress(ctx context.Context, providerJobID string) (*entity.Progress, error)

	// CancelGeneration cancels an ongoing generation
	CancelGeneration(ctx context.Context, providerJobID string) error

	// GetCapabilities returns provider capabilities
	GetCapabilities() ProviderCapabilities

	// HealthCheck performs a health check
	HealthCheck(ctx context.Context) (*ProviderHealth, error)
}

// ProviderSelector defines the interface for selecting the best provider
type ProviderSelector interface {
	// SelectProvider selects the best provider based on requirements
	SelectProvider(ctx context.Context, req ProviderSelectionRequest) (VideoProvider, error)

	// GetAvailableProviders returns all available providers
	GetAvailableProviders(ctx context.Context) ([]VideoProvider, error)

	// RefreshHealth refreshes health status for all providers
	RefreshHealth(ctx context.Context) error
}

// ProviderSelectionRequest contains criteria for provider selection
type ProviderSelectionRequest struct {
	UserTier           entity.UserTier
	PreferredProvider  *entity.AIProvider
	RequiredResolution entity.VideoResolution
	RequiredDuration   int
	AspectRatio        entity.AspectRatio
}
