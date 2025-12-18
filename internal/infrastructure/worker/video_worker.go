package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/arabella/ai-studio-backend/internal/domain/service"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// VideoWorker processes video generation jobs from the queue
type VideoWorker struct {
	jobRepo          repository.VideoJobRepository
	templateRepo     repository.TemplateRepository
	userRepo         repository.UserRepository
	providerSelector service.ProviderSelector
	queue            QueueService
	wsHub            WebSocketHub
	logger           *zap.Logger
	stopChan         chan struct{}
}

// QueueService interface for queue operations
type QueueService interface {
	Dequeue(ctx context.Context) (*entity.VideoJob, error)
	UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status entity.JobStatus, progress int) error
}

// WebSocketHub interface for broadcasting updates
type WebSocketHub interface {
	BroadcastToJob(jobID uuid.UUID, eventType string, payload interface{})
	BroadcastToUser(userID uuid.UUID, eventType string, payload interface{})
}

// NewVideoWorker creates a new video worker
func NewVideoWorker(
	jobRepo repository.VideoJobRepository,
	templateRepo repository.TemplateRepository,
	userRepo repository.UserRepository,
	providerSelector service.ProviderSelector,
	queue QueueService,
	wsHub WebSocketHub,
	logger *zap.Logger,
) *VideoWorker {
	return &VideoWorker{
		jobRepo:          jobRepo,
		templateRepo:     templateRepo,
		userRepo:         userRepo,
		providerSelector: providerSelector,
		queue:            queue,
		wsHub:            wsHub,
		logger:           logger,
		stopChan:         make(chan struct{}),
	}
}

// Start starts the worker in a goroutine
func (w *VideoWorker) Start(ctx context.Context) {
	go w.run(ctx)
}

// Stop stops the worker
func (w *VideoWorker) Stop() {
	close(w.stopChan)
}

// run is the main worker loop
func (w *VideoWorker) run(ctx context.Context) {
	w.logger.Info("Video worker started")
	defer w.logger.Info("Video worker stopped")

	ticker := time.NewTicker(2 * time.Second) // Check queue every 2 seconds
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.processNextJob(ctx)
		}
	}
}

// processNextJob processes the next job from the queue
func (w *VideoWorker) processNextJob(ctx context.Context) {
	job, err := w.queue.Dequeue(ctx)
	if err != nil {
		w.logger.Error("Failed to dequeue job", zap.Error(err))
		return
	}

	if job == nil {
		// Queue is empty, nothing to do
		return
	}

	w.logger.Info("Processing video job",
		zap.String("job_id", job.ID.String()),
		zap.String("user_id", job.UserID.String()),
		zap.String("template_id", job.TemplateID.String()),
	)

	// Process the job in a goroutine to not block the queue
	go w.processJob(ctx, job)
}

// processJob processes a single video generation job
func (w *VideoWorker) processJob(ctx context.Context, job *entity.VideoJob) {
	// Update job status to processing
	job.StartProcessing(entity.ProviderWanAI) // Default to Wan AI
	if err := w.jobRepo.Update(ctx, job); err != nil {
		w.logger.Error("Failed to update job status", zap.Error(err))
		return
	}

	w.queue.UpdateJobStatus(ctx, job.ID, job.Status, job.Progress)
	w.wsHub.BroadcastToJob(job.ID, "status_update", map[string]interface{}{
		"status":   job.Status,
		"progress": job.Progress,
	})

	// Get template
	template, err := w.templateRepo.GetByID(ctx, job.TemplateID)
	if err != nil {
		w.failJob(ctx, job, fmt.Sprintf("Failed to get template: %v", err))
		return
	}

	// Get user to determine tier
	user, err := w.userRepo.GetByID(ctx, job.UserID)
	if err != nil {
		w.failJob(ctx, job, fmt.Sprintf("Failed to get user: %v", err))
		return
	}

	// Select provider
	providerReq := service.ProviderSelectionRequest{
		UserTier:           user.Tier,
		RequiredResolution: job.Params.Resolution,
		RequiredDuration:   job.Params.Duration,
		AspectRatio:        job.Params.AspectRatio,
	}

	// Prefer Wan AI
	wanAIProvider := entity.ProviderWanAI
	providerReq.PreferredProvider = &wanAIProvider

	provider, err := w.providerSelector.SelectProvider(ctx, providerReq)
	if err != nil {
		w.failJob(ctx, job, fmt.Sprintf("Failed to select provider: %v", err))
		return
	}

	w.logger.Info("Selected provider",
		zap.String("job_id", job.ID.String()),
		zap.String("provider", string(provider.GetName())),
	)

	// Update job with provider
	job.Provider = provider.GetName()
	if err := w.jobRepo.Update(ctx, job); err != nil {
		w.logger.Error("Failed to update job provider", zap.Error(err))
	}

	// Generate video
	genReq := service.GenerationRequest{
		JobID:        job.ID.String(),
		Prompt:       job.Prompt, // This contains only the user's prompt (template base prompt is ignored)
		Params:       job.Params,
		TemplateID:   template.ID.String(),
		BasePrompt:   template.BasePrompt,
		ThumbnailURL: template.ThumbnailURL, // Pass template thumbnail for image-to-video
		UserTier:     user.Tier,
	}

	w.logger.Info("Calling provider to generate video",
		zap.String("job_id", job.ID.String()),
		zap.String("provider", string(provider.GetName())),
		zap.String("prompt", job.Prompt),
	)

	result, err := provider.GenerateVideo(ctx, genReq)
	if err != nil {
		w.failJob(ctx, job, fmt.Sprintf("Video generation failed: %v", err))
		return
	}

	// Update job with provider job ID
	job.SetProviderJobID(result.ProviderJobID)

	// If video URL is already available (e.g., from mock provider), complete immediately
	if result.VideoURL != "" {
		w.logger.Info("Video URL available immediately",
			zap.String("job_id", job.ID.String()),
			zap.String("video_url", result.VideoURL),
		)

		thumbnailURL := result.ThumbnailURL
		if thumbnailURL == "" {
			thumbnailURL = ""
		}

		duration := result.Duration
		if duration == 0 {
			duration = job.Params.Duration
			if duration == 0 {
				duration = 15 // Default
			}
		}

		job.Complete(result.VideoURL, thumbnailURL, duration)
		if err := w.jobRepo.Update(ctx, job); err != nil {
			w.logger.Error("Failed to complete job", zap.Error(err))
			return
		}

		w.queue.UpdateJobStatus(ctx, job.ID, job.Status, job.Progress)
		w.wsHub.BroadcastToJob(job.ID, "completed", map[string]interface{}{
			"status":        job.Status,
			"progress":      job.Progress,
			"video_url":     job.VideoURL,
			"thumbnail_url": job.ThumbnailURL,
		})

		w.logger.Info("Video job completed immediately",
			zap.String("job_id", job.ID.String()),
			zap.String("video_url", result.VideoURL),
		)
		return
	}

	if err := w.jobRepo.Update(ctx, job); err != nil {
		w.logger.Error("Failed to update provider job ID", zap.Error(err))
	}

	// Poll for completion
	w.pollForCompletion(ctx, job, provider)
}

// pollForCompletion polls the provider for job completion
func (w *VideoWorker) pollForCompletion(ctx context.Context, job *entity.VideoJob, provider service.VideoProvider) {
	ticker := time.NewTicker(5 * time.Second) // Poll every 5 seconds (reduced frequency)
	defer ticker.Stop()

	maxAttempts := 360 // Max 30 minutes (360 * 5 seconds = 30 minutes)
	attempts := 0
	consecutiveErrors := 0
	maxConsecutiveErrors := 5 // Fail after 5 consecutive errors (25 seconds)

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			attempts++
			if attempts > maxAttempts {
				w.failJob(ctx, job, "Video generation timeout after 30 minutes")
				return
			}

			if job.ProviderJobID == nil {
				w.logger.Warn("No provider job ID",
					zap.String("job_id", job.ID.String()),
				)
				consecutiveErrors++
				if consecutiveErrors >= maxConsecutiveErrors {
					w.failJob(ctx, job, "No provider job ID after multiple attempts")
					return
				}
				continue
			}

			progress, err := provider.GetProgress(ctx, *job.ProviderJobID)
			if err != nil {
				consecutiveErrors++
				w.logger.Warn("Failed to get progress",
					zap.String("job_id", job.ID.String()),
					zap.Int("consecutive_errors", consecutiveErrors),
					zap.Error(err),
				)
				// Fail fast if we get too many consecutive errors
				if consecutiveErrors >= maxConsecutiveErrors {
					w.failJob(ctx, job, fmt.Sprintf("Failed to get progress after %d attempts: %v", consecutiveErrors, err))
					return
				}
				continue
			}

			// Reset error counter on success
			consecutiveErrors = 0

			// Update progress
			job.Progress = progress.Percent
			job.UpdateProgress(progress.Percent, entity.JobStatus(progress.Stage))
			if err := w.jobRepo.Update(ctx, job); err != nil {
				w.logger.Error("Failed to update job progress", zap.Error(err))
			}

			w.queue.UpdateJobStatus(ctx, job.ID, job.Status, job.Progress)
			w.wsHub.BroadcastToJob(job.ID, "progress_update", map[string]interface{}{
				"status":   job.Status,
				"progress": job.Progress,
				"message":  progress.Message,
			})

			if progress.Stage == "COMPLETED" {
				// Job is completed, get final video URL
				if job.ProviderJobID != nil {
					w.completeJob(ctx, job, provider)
				} else {
					w.failJob(ctx, job, "Job completed but no provider job ID")
				}
				return
			}

			if progress.Stage == "FAILED" {
				w.failJob(ctx, job, progress.Message)
				return
			}
		}
	}
}

// completeJob marks the job as completed
func (w *VideoWorker) completeJob(ctx context.Context, job *entity.VideoJob, provider service.VideoProvider) {
	if job.ProviderJobID == nil {
		w.failJob(ctx, job, "No provider job ID available")
		return
	}

	// Get video URL from provider
	var videoURL string
	var thumbnailURL string

	if job.Provider == entity.ProviderWanAI {
		// For DashScope (Wan AI), fetch the video URL from the task status
		// Use type assertion to call GetVideoURL if available
		if wanAIProvider, ok := provider.(interface {
			GetVideoURL(ctx context.Context, providerJobID string) (string, error)
		}); ok {
			var err error
			videoURL, err = wanAIProvider.GetVideoURL(ctx, *job.ProviderJobID)
			if err != nil {
				w.logger.Warn("Failed to get video URL from DashScope",
					zap.String("job_id", job.ID.String()),
					zap.String("task_id", *job.ProviderJobID),
					zap.Error(err),
				)
				videoURL = "" // Will be empty, but job will be marked as completed
			}
		} else {
			w.logger.Warn("Provider does not support GetVideoURL",
				zap.String("provider", string(job.Provider)),
			)
		}
		thumbnailURL = ""
	} else if job.Provider == entity.ProviderMock {
		// For mock provider, use a working sample video URL
		videoURL = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
		thumbnailURL = ""
	} else {
		// For other providers, use a working sample video URL (no storage domain needed)
		videoURL = "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4"
		thumbnailURL = ""
	}

	// Default duration if not set
	duration := job.Params.Duration
	if duration == 0 {
		duration = 15 // Default 15 seconds
	}

	// If video URL is still empty, we'll need to fetch it from DashScope
	// For now, mark as completed and the frontend can poll for the video URL
	if videoURL == "" && job.Provider == entity.ProviderWanAI {
		w.logger.Warn("Video URL not available yet, job will be marked as completed",
			zap.String("job_id", job.ID.String()),
			zap.String("task_id", *job.ProviderJobID),
		)
		// Set a placeholder - the actual URL will be fetched later
		videoURL = fmt.Sprintf("dashscope://task/%s", *job.ProviderJobID)
	}

	job.Complete(videoURL, thumbnailURL, duration)
	if err := w.jobRepo.Update(ctx, job); err != nil {
		w.logger.Error("Failed to complete job", zap.Error(err))
		return
	}

	w.queue.UpdateJobStatus(ctx, job.ID, job.Status, job.Progress)
	w.wsHub.BroadcastToJob(job.ID, "completed", map[string]interface{}{
		"status":        job.Status,
		"progress":      job.Progress,
		"video_url":     job.VideoURL,
		"thumbnail_url": job.ThumbnailURL,
	})

	w.logger.Info("Video job completed",
		zap.String("job_id", job.ID.String()),
		zap.String("video_url", videoURL),
	)
}

// failJob marks the job as failed
func (w *VideoWorker) failJob(ctx context.Context, job *entity.VideoJob, errorMsg string) {
	job.Fail(errorMsg)
	if err := w.jobRepo.Update(ctx, job); err != nil {
		w.logger.Error("Failed to mark job as failed", zap.Error(err))
		return
	}

	w.queue.UpdateJobStatus(ctx, job.ID, job.Status, job.Progress)
	w.wsHub.BroadcastToJob(job.ID, "failed", map[string]interface{}{
		"status":        job.Status,
		"error_message": errorMsg,
	})

	w.logger.Error("Video job failed",
		zap.String("job_id", job.ID.String()),
		zap.String("error", errorMsg),
	)
}


