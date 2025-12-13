package repository

import (
	"context"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/google/uuid"
)

// VideoJobFilter defines filters for video job queries
type VideoJobFilter struct {
	UserID   *uuid.UUID
	Status   *entity.JobStatus
	Provider *entity.AIProvider
}

// VideoJobRepository defines the interface for video job data access
type VideoJobRepository interface {
	// Create creates a new video job
	Create(ctx context.Context, job *entity.VideoJob) error

	// GetByID retrieves a video job by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.VideoJob, error)

	// Update updates an existing video job
	Update(ctx context.Context, job *entity.VideoJob) error

	// UpdateStatus updates the status and progress of a job
	UpdateStatus(ctx context.Context, id uuid.UUID, status entity.JobStatus, progress int) error

	// Complete marks a job as completed with video details
	Complete(ctx context.Context, id uuid.UUID, videoURL, thumbnailURL string, duration int) error

	// Fail marks a job as failed with error message
	Fail(ctx context.Context, id uuid.UUID, errorMessage string) error

	// List lists video jobs with filtering and pagination
	List(ctx context.Context, filter VideoJobFilter, offset, limit int) ([]*entity.VideoJob, int64, error)

	// GetByUserID retrieves all jobs for a user
	GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.VideoJob, int64, error)

	// GetPendingJobs retrieves pending jobs for processing
	GetPendingJobs(ctx context.Context, limit int) ([]*entity.VideoJob, error)

	// GetActiveJobsCount returns the count of active jobs for a user
	GetActiveJobsCount(ctx context.Context, userID uuid.UUID) (int, error)

	// GetRecentByUser retrieves recent jobs for a user
	GetRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.VideoJob, error)

	// CountByStatus returns the count of jobs by status
	CountByStatus(ctx context.Context, status entity.JobStatus) (int64, error)
}

