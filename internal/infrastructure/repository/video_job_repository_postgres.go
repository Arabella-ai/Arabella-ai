package repository

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VideoJobRepositoryPostgres implements VideoJobRepository for PostgreSQL
type VideoJobRepositoryPostgres struct {
	pool *pgxpool.Pool
}

// NewVideoJobRepositoryPostgres creates a new VideoJobRepositoryPostgres
func NewVideoJobRepositoryPostgres(pool *pgxpool.Pool) repository.VideoJobRepository {
	return &VideoJobRepositoryPostgres{pool: pool}
}

// Create creates a new video job
func (r *VideoJobRepositoryPostgres) Create(ctx context.Context, job *entity.VideoJob) error {
	query := `
		INSERT INTO video_jobs (id, user_id, template_id, prompt, params, status, progress,
		                        provider, provider_job_id, video_url, thumbnail_url, duration_seconds,
		                        credits_charged, error_message, created_at, started_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	paramsJSON, err := json.Marshal(job.Params)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query,
		job.ID,
		job.UserID,
		job.TemplateID,
		job.Prompt,
		paramsJSON,
		job.Status,
		job.Progress,
		job.Provider,
		job.ProviderJobID,
		job.VideoURL,
		job.ThumbnailURL,
		job.DurationSeconds,
		job.CreditsCharged,
		job.ErrorMessage,
		job.CreatedAt,
		job.StartedAt,
		job.CompletedAt,
	)

	return err
}

// GetByID retrieves a video job by ID
func (r *VideoJobRepositoryPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.VideoJob, error) {
	query := `
		SELECT id, user_id, template_id, prompt, params, status, progress,
		       provider, provider_job_id, video_url, thumbnail_url, duration_seconds,
		       credits_charged, error_message, created_at, started_at, completed_at
		FROM video_jobs
		WHERE id = $1
	`

	job := &entity.VideoJob{}
	var paramsJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.UserID,
		&job.TemplateID,
		&job.Prompt,
		&paramsJSON,
		&job.Status,
		&job.Progress,
		&job.Provider,
		&job.ProviderJobID,
		&job.VideoURL,
		&job.ThumbnailURL,
		&job.DurationSeconds,
		&job.CreditsCharged,
		&job.ErrorMessage,
		&job.CreatedAt,
		&job.StartedAt,
		&job.CompletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrJobNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(paramsJSON, &job.Params); err != nil {
		return nil, err
	}

	return job, nil
}

// Update updates an existing video job
func (r *VideoJobRepositoryPostgres) Update(ctx context.Context, job *entity.VideoJob) error {
	query := `
		UPDATE video_jobs
		SET status = $2, progress = $3, provider = $4, provider_job_id = $5,
		    video_url = $6, thumbnail_url = $7, duration_seconds = $8,
		    error_message = $9, started_at = $10, completed_at = $11
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		job.ID,
		job.Status,
		job.Progress,
		job.Provider,
		job.ProviderJobID,
		job.VideoURL,
		job.ThumbnailURL,
		job.DurationSeconds,
		job.ErrorMessage,
		job.StartedAt,
		job.CompletedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrJobNotFound
	}

	return nil
}

// UpdateStatus updates the status and progress of a job
func (r *VideoJobRepositoryPostgres) UpdateStatus(ctx context.Context, id uuid.UUID, status entity.JobStatus, progress int) error {
	query := `UPDATE video_jobs SET status = $2, progress = $3 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, status, progress)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrJobNotFound
	}

	return nil
}

// Complete marks a job as completed
func (r *VideoJobRepositoryPostgres) Complete(ctx context.Context, id uuid.UUID, videoURL, thumbnailURL string, duration int) error {
	query := `
		UPDATE video_jobs
		SET status = $2, progress = 100, video_url = $3, thumbnail_url = $4,
		    duration_seconds = $5, completed_at = $6
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, entity.JobStatusCompleted, videoURL, thumbnailURL, duration, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrJobNotFound
	}

	return nil
}

// Fail marks a job as failed
func (r *VideoJobRepositoryPostgres) Fail(ctx context.Context, id uuid.UUID, errorMessage string) error {
	query := `
		UPDATE video_jobs
		SET status = $2, error_message = $3, completed_at = $4
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, id, entity.JobStatusFailed, errorMessage, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrJobNotFound
	}

	return nil
}

// List lists video jobs with filtering and pagination
func (r *VideoJobRepositoryPostgres) List(ctx context.Context, filter repository.VideoJobFilter, offset, limit int) ([]*entity.VideoJob, int64, error) {
	// Build WHERE clause
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = $1")
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Status != nil {
		conditions = append(conditions, "status = $"+string(rune('0'+argIndex)))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Provider != nil {
		conditions = append(conditions, "provider = $"+string(rune('0'+argIndex)))
		args = append(args, *filter.Provider)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE "
		for i, cond := range conditions {
			if i > 0 {
				whereClause += " AND "
			}
			whereClause += cond
		}
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM video_jobs " + whereClause
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Get jobs
	query := `
		SELECT id, user_id, template_id, prompt, params, status, progress,
		       provider, provider_job_id, video_url, thumbnail_url, duration_seconds,
		       credits_charged, error_message, created_at, started_at, completed_at
		FROM video_jobs
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT $` + string(rune('0'+argIndex)) + ` OFFSET $` + string(rune('0'+argIndex+1))

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	jobs, err := r.scanJobs(rows)
	if err != nil {
		return nil, 0, err
	}

	return jobs, total, nil
}

// GetByUserID retrieves all jobs for a user
func (r *VideoJobRepositoryPostgres) GetByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*entity.VideoJob, int64, error) {
	filter := repository.VideoJobFilter{UserID: &userID}
	return r.List(ctx, filter, offset, limit)
}

// GetPendingJobs retrieves pending jobs for processing
func (r *VideoJobRepositoryPostgres) GetPendingJobs(ctx context.Context, limit int) ([]*entity.VideoJob, error) {
	query := `
		SELECT id, user_id, template_id, prompt, params, status, progress,
		       provider, provider_job_id, video_url, thumbnail_url, duration_seconds,
		       credits_charged, error_message, created_at, started_at, completed_at
		FROM video_jobs
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, entity.JobStatusPending, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// GetActiveJobsCount returns the count of active jobs for a user
func (r *VideoJobRepositoryPostgres) GetActiveJobsCount(ctx context.Context, userID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM video_jobs
		WHERE user_id = $1 AND status IN ($2, $3, $4, $5)
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID,
		entity.JobStatusPending,
		entity.JobStatusProcessing,
		entity.JobStatusDiffusing,
		entity.JobStatusUploading,
	).Scan(&count)

	return count, err
}

// GetRecentByUser retrieves recent jobs for a user
func (r *VideoJobRepositoryPostgres) GetRecentByUser(ctx context.Context, userID uuid.UUID, limit int) ([]*entity.VideoJob, error) {
	query := `
		SELECT id, user_id, template_id, prompt, params, status, progress,
		       provider, provider_job_id, video_url, thumbnail_url, duration_seconds,
		       credits_charged, error_message, created_at, started_at, completed_at
		FROM video_jobs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanJobs(rows)
}

// CountByStatus returns the count of jobs by status
func (r *VideoJobRepositoryPostgres) CountByStatus(ctx context.Context, status entity.JobStatus) (int64, error) {
	query := `SELECT COUNT(*) FROM video_jobs WHERE status = $1`

	var count int64
	err := r.pool.QueryRow(ctx, query, status).Scan(&count)

	return count, err
}

// scanJobs scans rows into video jobs
func (r *VideoJobRepositoryPostgres) scanJobs(rows pgx.Rows) ([]*entity.VideoJob, error) {
	var jobs []*entity.VideoJob
	for rows.Next() {
		job := &entity.VideoJob{}
		var paramsJSON []byte

		err := rows.Scan(
			&job.ID,
			&job.UserID,
			&job.TemplateID,
			&job.Prompt,
			&paramsJSON,
			&job.Status,
			&job.Progress,
			&job.Provider,
			&job.ProviderJobID,
			&job.VideoURL,
			&job.ThumbnailURL,
			&job.DurationSeconds,
			&job.CreditsCharged,
			&job.ErrorMessage,
			&job.CreatedAt,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(paramsJSON, &job.Params); err != nil {
			return nil, err
		}

		jobs = append(jobs, job)
	}

	return jobs, nil
}

