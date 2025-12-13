package provider

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MockProvider implements a mock AI provider for development
type MockProvider struct {
	logger       *zap.Logger
	jobs         map[string]*mockJob
	simulateTime bool
}

type mockJob struct {
	id        string
	progress  int
	status    string
	startTime time.Time
}

// NewMockProvider creates a new MockProvider
func NewMockProvider(logger *zap.Logger, simulateTime bool) service.VideoProvider {
	return &MockProvider{
		logger:       logger,
		jobs:         make(map[string]*mockJob),
		simulateTime: simulateTime,
	}
}

// GetName returns the provider name
func (p *MockProvider) GetName() entity.AIProvider {
	return entity.ProviderMock
}

// GenerateVideo initiates mock video generation
func (p *MockProvider) GenerateVideo(ctx context.Context, req service.GenerationRequest) (*entity.GenerationResult, error) {
	jobID := uuid.New().String()

	p.jobs[jobID] = &mockJob{
		id:        jobID,
		progress:  0,
		status:    "processing",
		startTime: time.Now(),
	}

	p.logger.Info("Mock video generation started",
		zap.String("job_id", jobID),
		zap.String("prompt", req.Prompt),
	)

	// Simulate instant completion for fast development
	if !p.simulateTime {
		return &entity.GenerationResult{
			ProviderJobID: jobID,
			VideoURL:      fmt.Sprintf("https://cdn.arabella.app/videos/%s.mp4", jobID),
			ThumbnailURL:  fmt.Sprintf("https://cdn.arabella.app/thumbnails/%s.jpg", jobID),
			Duration:      req.Params.Duration,
		}, nil
	}

	// Return job ID for async processing
	return &entity.GenerationResult{
		ProviderJobID: jobID,
	}, nil
}

// GetProgress retrieves generation progress
func (p *MockProvider) GetProgress(ctx context.Context, providerJobID string) (*entity.Progress, error) {
	job, ok := p.jobs[providerJobID]
	if !ok {
		return nil, entity.ErrJobNotFound
	}

	// Simulate progress
	elapsed := time.Since(job.startTime)
	progress := int(elapsed.Seconds() / 30 * 100) // 30 seconds to complete
	if progress > 100 {
		progress = 100
	}

	job.progress = progress

	stage := "PROCESSING"
	if progress > 30 && progress < 80 {
		stage = "DIFFUSING_FRAMES"
	} else if progress >= 80 && progress < 100 {
		stage = "UPLOADING"
	} else if progress >= 100 {
		stage = "COMPLETED"
	}

	return &entity.Progress{
		Percent: progress,
		Stage:   stage,
		Message: fmt.Sprintf("Progress: %d%%", progress),
	}, nil
}

// CancelGeneration cancels an ongoing generation
func (p *MockProvider) CancelGeneration(ctx context.Context, providerJobID string) error {
	if _, ok := p.jobs[providerJobID]; !ok {
		return entity.ErrJobNotFound
	}

	delete(p.jobs, providerJobID)
	p.logger.Info("Mock generation cancelled", zap.String("job_id", providerJobID))

	return nil
}

// GetCapabilities returns provider capabilities
func (p *MockProvider) GetCapabilities() service.ProviderCapabilities {
	return service.ProviderCapabilities{
		Name:            "Mock Provider",
		MaxDuration:     60,
		MaxResolution:   entity.Resolution1080p,
		SupportedRatios: []entity.AspectRatio{entity.AspectRatio16x9, entity.AspectRatio9x16, entity.AspectRatio1x1},
		EstimatedTime:   5, // 5 seconds per second of video
		QualityTier:     "standard",
		SupportsStyles:  true,
		CostPerSecond:   0.0,
	}
}

// HealthCheck performs a health check
func (p *MockProvider) HealthCheck(ctx context.Context) (*service.ProviderHealth, error) {
	// Simulate occasional unhealthy status (5% chance)
	isHealthy := rand.Float32() > 0.05

	return &service.ProviderHealth{
		IsHealthy:    isHealthy,
		QueueDepth:   rand.Intn(10),
		ResponseTime: int64(rand.Intn(100) + 50),
		ErrorRate:    rand.Float64() * 0.05,
		LastChecked:  time.Now().Unix(),
	}, nil
}

