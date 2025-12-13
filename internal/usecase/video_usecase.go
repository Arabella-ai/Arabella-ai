package usecase

import (
	"context"
	"fmt"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/arabella/ai-studio-backend/internal/domain/service"
	"github.com/google/uuid"
)

// VideoGenerationRequest represents a video generation request
type VideoGenerationRequest struct {
	TemplateID uuid.UUID          `json:"template_id" binding:"required"`
	Prompt     string             `json:"prompt" binding:"required,min=10,max=2000"`
	Params     *entity.VideoParams `json:"params,omitempty"`
}

// VideoGenerationResponse represents the response after initiating generation
type VideoGenerationResponse struct {
	JobID         uuid.UUID   `json:"job_id"`
	Status        string      `json:"status"`
	EstimatedTime int         `json:"estimated_time"` // in seconds
	QueuePosition int         `json:"queue_position"`
}

// VideoJobListRequest represents a request to list video jobs
type VideoJobListRequest struct {
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// VideoJobListResponse represents the paginated job list response
type VideoJobListResponse struct {
	Jobs       []*entity.VideoJob `json:"jobs"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// JobQueueService interface for job queue operations
type JobQueueService interface {
	Enqueue(ctx context.Context, job *entity.VideoJob) error
	GetQueuePosition(ctx context.Context, jobID uuid.UUID) (int, error)
	GetQueueDepth(ctx context.Context) (int, error)
}

// VideoUseCase handles video generation business logic
type VideoUseCase struct {
	jobRepo          repository.VideoJobRepository
	templateRepo     repository.TemplateRepository
	userRepo         repository.UserRepository
	providerSelector service.ProviderSelector
	jobQueue         JobQueueService
	wsHub            WebSocketHub
}

// WebSocketHub interface for real-time updates
type WebSocketHub interface {
	BroadcastToJob(jobID uuid.UUID, eventType string, payload interface{})
	BroadcastToUser(userID uuid.UUID, eventType string, payload interface{})
}

// NewVideoUseCase creates a new VideoUseCase
func NewVideoUseCase(
	jobRepo repository.VideoJobRepository,
	templateRepo repository.TemplateRepository,
	userRepo repository.UserRepository,
	providerSelector service.ProviderSelector,
	jobQueue JobQueueService,
	wsHub WebSocketHub,
) *VideoUseCase {
	return &VideoUseCase{
		jobRepo:          jobRepo,
		templateRepo:     templateRepo,
		userRepo:         userRepo,
		providerSelector: providerSelector,
		jobQueue:         jobQueue,
		wsHub:            wsHub,
	}
}

// GenerateVideo initiates a video generation job
func (uc *VideoUseCase) GenerateVideo(ctx context.Context, userID uuid.UUID, req VideoGenerationRequest) (*VideoGenerationResponse, error) {
	// Get the user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get the template
	template, err := uc.templateRepo.GetByID(ctx, req.TemplateID)
	if err != nil {
		return nil, err
	}

	// Validate template access
	if !template.IsActive {
		return nil, entity.ErrTemplateNotActive
	}

	if template.IsPremium && !user.IsPremium() {
		return nil, entity.ErrTemplatePremiumOnly
	}

	if !user.HasSufficientCredits(template.CreditCost) {
		return nil, entity.ErrInsufficientCredits
	}

	// Merge params with template defaults
	params := template.DefaultParams
	if req.Params != nil {
		if req.Params.Duration > 0 {
			params.Duration = req.Params.Duration
		}
		if req.Params.Resolution != "" {
			params.Resolution = req.Params.Resolution
		}
		if req.Params.AspectRatio != "" {
			params.AspectRatio = req.Params.AspectRatio
		}
		if req.Params.FPS > 0 {
			params.FPS = req.Params.FPS
		}
		if req.Params.Style != "" {
			params.Style = req.Params.Style
		}
		if req.Params.NegativePrompt != "" {
			params.NegativePrompt = req.Params.NegativePrompt
		}
	}

	// Combine base prompt with user prompt
	fullPrompt := fmt.Sprintf("%s. %s", template.BasePrompt, req.Prompt)

	// Create the job
	job := entity.NewVideoJob(userID, template.ID, fullPrompt, params, template.CreditCost)

	// Deduct credits
	if err := uc.userRepo.UpdateCredits(ctx, userID, -template.CreditCost); err != nil {
		return nil, err
	}

	// Save the job
	if err := uc.jobRepo.Create(ctx, job); err != nil {
		// Refund credits on failure
		_ = uc.userRepo.UpdateCredits(ctx, userID, template.CreditCost)
		return nil, err
	}

	// Increment template usage
	_ = uc.templateRepo.IncrementUsage(ctx, template.ID)

	// Enqueue the job for processing
	if err := uc.jobQueue.Enqueue(ctx, job); err != nil {
		// Mark job as failed but keep credits deducted (will be refunded by cleanup)
		job.Fail("Failed to enqueue job")
		_ = uc.jobRepo.Update(ctx, job)
		return nil, err
	}

	// Get queue position
	queuePosition, _ := uc.jobQueue.GetQueuePosition(ctx, job.ID)

	// Calculate estimated time
	estimatedTime := int(template.EstimatedTime.Seconds()) + (queuePosition * 30)

	return &VideoGenerationResponse{
		JobID:         job.ID,
		Status:        string(job.Status),
		EstimatedTime: estimatedTime,
		QueuePosition: queuePosition,
	}, nil
}

// GetJobStatus retrieves the status of a video job
func (uc *VideoUseCase) GetJobStatus(ctx context.Context, userID, jobID uuid.UUID) (*entity.VideoJob, error) {
	job, err := uc.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if job.UserID != userID {
		return nil, entity.ErrUnauthorized
	}

	return job, nil
}

// GetUserJobs retrieves all jobs for a user
func (uc *VideoUseCase) GetUserJobs(ctx context.Context, userID uuid.UUID, req VideoJobListRequest) (*VideoJobListResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize

	// Build filter
	filter := repository.VideoJobFilter{
		UserID: &userID,
	}

	if req.Status != "" {
		status := entity.JobStatus(req.Status)
		filter.Status = &status
	}

	// Get jobs
	jobs, total, err := uc.jobRepo.List(ctx, filter, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &VideoJobListResponse{
		Jobs:       jobs,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// CancelJob cancels a video generation job
func (uc *VideoUseCase) CancelJob(ctx context.Context, userID, jobID uuid.UUID) error {
	job, err := uc.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}

	// Verify ownership
	if job.UserID != userID {
		return entity.ErrUnauthorized
	}

	// Check if job can be cancelled
	if !job.CanBeCancelled() {
		return entity.ErrJobCannotBeCancelled
	}

	// Cancel the job
	job.Cancel()
	if err := uc.jobRepo.Update(ctx, job); err != nil {
		return err
	}

	// Refund credits
	_ = uc.userRepo.UpdateCredits(ctx, userID, job.CreditsCharged)

	// Broadcast cancellation
	if uc.wsHub != nil {
		uc.wsHub.BroadcastToJob(jobID, "cancelled", map[string]interface{}{
			"job_id": jobID.String(),
			"status": "cancelled",
		})
	}

	return nil
}

// GetRecentJobs retrieves recent jobs for a user
func (uc *VideoUseCase) GetRecentJobs(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.VideoJob, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	return uc.jobRepo.GetRecentByUser(ctx, userID, limit)
}

// GetVideo retrieves a completed video
func (uc *VideoUseCase) GetVideo(ctx context.Context, userID, jobID uuid.UUID) (*entity.VideoJob, error) {
	job, err := uc.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if job.UserID != userID {
		return nil, entity.ErrUnauthorized
	}

	if job.Status != entity.JobStatusCompleted {
		return nil, entity.NewDomainError("VIDEO_NOT_READY", "Video is not ready yet", nil)
	}

	return job, nil
}

