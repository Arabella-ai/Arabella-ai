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
	ProviderGeminiVEO AIProvider = "gemini_veo"
	ProviderOpenAISora AIProvider = "openai_sora"
	ProviderRunway     AIProvider = "runway"
	ProviderPikaLabs   AIProvider = "pika_labs"
	ProviderMock       AIProvider = "mock" // For development
)

// VideoJob represents a video generation job
type VideoJob struct {
	ID             uuid.UUID   `json:"id"`
	UserID         uuid.UUID   `json:"user_id"`
	TemplateID     uuid.UUID   `json:"template_id"`
	Prompt         string      `json:"prompt"`
	Params         VideoParams `json:"params"`
	Status         JobStatus   `json:"status"`
	Progress       int         `json:"progress"` // 0-100
	Provider       AIProvider  `json:"provider"`
	ProviderJobID  *string     `json:"provider_job_id,omitempty"`
	VideoURL       *string     `json:"video_url,omitempty"`
	ThumbnailURL   *string     `json:"thumbnail_url,omitempty"`
	DurationSeconds int        `json:"duration_seconds,omitempty"`
	CreditsCharged int         `json:"credits_charged"`
	ErrorMessage   *string     `json:"error_message,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	StartedAt      *time.Time  `json:"started_at,omitempty"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty"`
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

