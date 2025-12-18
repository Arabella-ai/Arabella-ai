package entity

import (
	"time"

	"github.com/google/uuid"
)

// JobStatus represents the current state of a video generation job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusDiffusing  JobStatus = "diffusing"
	JobStatusUploading  JobStatus = "uploading"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusCancelled  JobStatus = "cancelled"
)

// AIProvider represents available AI video generation providers
type AIProvider string

const (
	ProviderGeminiVEO  AIProvider = "gemini_veo"
	ProviderOpenAISora AIProvider = "openai_sora"
	ProviderRunway     AIProvider = "runway"
	ProviderPikaLabs   AIProvider = "pika_labs"
	ProviderWanAI      AIProvider = "wan_ai"
	ProviderMock       AIProvider = "mock" // For development
)

// VideoJob represents a video generation job
// @Description Video generation job with status, progress, and result URLs
type VideoJob struct {
	ID              uuid.UUID   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID          uuid.UUID   `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TemplateID      uuid.UUID   `json:"template_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Prompt          string      `json:"prompt" example:"A beautiful sunset over mountains"`
	Params          VideoParams `json:"params"`
	Status          JobStatus   `json:"status" example:"completed" enums:"pending,processing,diffusing,uploading,completed,failed,cancelled"`
	Progress        int         `json:"progress" example:"100" minimum:"0" maximum:"100"` // 0-100
	Provider        AIProvider  `json:"provider" example:"gemini_veo" enums:"gemini_veo,openai_sora,runway,pika_labs,wan_ai,mock"`
	ProviderJobID   *string     `json:"provider_job_id,omitempty" example:"gemini-job-123"`
	VideoURL        *string     `json:"video_url,omitempty" example:"https://storage.googleapis.com/gemini-videos/abc123.mp4"`
	ThumbnailURL    *string     `json:"thumbnail_url,omitempty" example:"https://cdn.arabella.app/thumbnails/abc123.jpg"`
	DurationSeconds int         `json:"duration_seconds,omitempty" example:"15"`
	CreditsCharged  int         `json:"credits_charged" example:"2"`
	ErrorMessage    *string     `json:"error_message,omitempty" example:"Generation failed"`
	CreatedAt       time.Time   `json:"created_at" example:"2025-12-13T16:00:00Z"`
	StartedAt       *time.Time  `json:"started_at,omitempty" example:"2025-12-13T16:00:05Z"`
	CompletedAt     *time.Time  `json:"completed_at,omitempty" example:"2025-12-13T16:02:00Z"`
}

// NewVideoJob creates a new video generation job
func NewVideoJob(userID, templateID uuid.UUID, prompt string, params VideoParams, creditCost int) *VideoJob {
	return &VideoJob{
		ID:             uuid.New(),
		UserID:         userID,
		TemplateID:     templateID,
		Prompt:         prompt,
		Params:         params,
		Status:         JobStatusPending,
		Progress:       0,
		CreditsCharged: creditCost,
		CreatedAt:      time.Now(),
	}
}

// StartProcessing marks the job as being processed
func (j *VideoJob) StartProcessing(provider AIProvider) {
	j.Status = JobStatusProcessing
	j.Provider = provider
	now := time.Now()
	j.StartedAt = &now
}

// UpdateProgress updates the job progress
func (j *VideoJob) UpdateProgress(progress int, status JobStatus) {
	j.Progress = progress
	j.Status = status
}

// SetProviderJobID sets the external provider job ID
func (j *VideoJob) SetProviderJobID(providerJobID string) {
	j.ProviderJobID = &providerJobID
}

// Complete marks the job as completed
func (j *VideoJob) Complete(videoURL, thumbnailURL string, duration int) {
	j.Status = JobStatusCompleted
	j.Progress = 100
	j.VideoURL = &videoURL
	j.ThumbnailURL = &thumbnailURL
	j.DurationSeconds = duration
	now := time.Now()
	j.CompletedAt = &now
}

// Fail marks the job as failed
func (j *VideoJob) Fail(errorMessage string) {
	j.Status = JobStatusFailed
	j.ErrorMessage = &errorMessage
	now := time.Now()
	j.CompletedAt = &now
}

// Cancel marks the job as cancelled
func (j *VideoJob) Cancel() {
	j.Status = JobStatusCancelled
	now := time.Now()
	j.CompletedAt = &now
}

// IsTerminal checks if the job is in a terminal state
func (j *VideoJob) IsTerminal() bool {
	return j.Status == JobStatusCompleted ||
		j.Status == JobStatusFailed ||
		j.Status == JobStatusCancelled
}

// CanBeCancelled checks if the job can be cancelled
func (j *VideoJob) CanBeCancelled() bool {
	return j.Status == JobStatusPending || j.Status == JobStatusProcessing
}

// GenerationResult represents the result from an AI provider
type GenerationResult struct {
	ProviderJobID string
	VideoURL      string
	ThumbnailURL  string
	Duration      int
}

// Progress represents generation progress from an AI provider
type Progress struct {
	Percent int
	Stage   string
	Message string
}
