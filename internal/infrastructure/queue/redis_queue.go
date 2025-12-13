package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	jobQueueKey     = "arabella:jobs:queue"
	jobDataPrefix   = "arabella:jobs:data:"
	jobStatusPrefix = "arabella:jobs:status:"
)

// RedisQueue implements job queue using Redis
type RedisQueue struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisQueue creates a new Redis-based job queue
func NewRedisQueue(client *redis.Client, logger *zap.Logger) *RedisQueue {
	return &RedisQueue{
		client: client,
		logger: logger,
	}
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *entity.VideoJob) error {
	// Store job data
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	dataKey := jobDataPrefix + job.ID.String()
	if err := q.client.Set(ctx, dataKey, jobData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store job data: %w", err)
	}

	// Add to queue with priority score (based on timestamp and user tier)
	score := float64(time.Now().UnixNano())

	if err := q.client.ZAdd(ctx, jobQueueKey, redis.Z{
		Score:  score,
		Member: job.ID.String(),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add to queue: %w", err)
	}

	q.logger.Info("Job enqueued",
		zap.String("job_id", job.ID.String()),
		zap.String("user_id", job.UserID.String()),
	)

	return nil
}

// Dequeue retrieves and removes the next job from the queue
func (q *RedisQueue) Dequeue(ctx context.Context) (*entity.VideoJob, error) {
	// Get the first job from the queue
	result, err := q.client.ZPopMin(ctx, jobQueueKey, 1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue: %w", err)
	}

	if len(result) == 0 {
		return nil, nil // Queue is empty
	}

	jobID := result[0].Member

	// Get job data
	dataKey := jobDataPrefix + jobID
	jobData, err := q.client.Get(ctx, dataKey).Result()
	if err == redis.Nil {
		return nil, nil // Job data not found
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	var job entity.VideoJob
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

// GetQueuePosition returns the position of a job in the queue
func (q *RedisQueue) GetQueuePosition(ctx context.Context, jobID uuid.UUID) (int, error) {
	rank, err := q.client.ZRank(ctx, jobQueueKey, jobID.String()).Result()
	if err == redis.Nil {
		return -1, nil // Job not in queue
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get queue position: %w", err)
	}

	return int(rank), nil
}

// GetQueueDepth returns the current queue depth
func (q *RedisQueue) GetQueueDepth(ctx context.Context) (int, error) {
	count, err := q.client.ZCard(ctx, jobQueueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue depth: %w", err)
	}

	return int(count), nil
}

// RemoveJob removes a job from the queue
func (q *RedisQueue) RemoveJob(ctx context.Context, jobID uuid.UUID) error {
	// Remove from queue
	if err := q.client.ZRem(ctx, jobQueueKey, jobID.String()).Err(); err != nil {
		return fmt.Errorf("failed to remove from queue: %w", err)
	}

	// Remove job data
	dataKey := jobDataPrefix + jobID.String()
	if err := q.client.Del(ctx, dataKey).Err(); err != nil {
		return fmt.Errorf("failed to remove job data: %w", err)
	}

	return nil
}

// UpdateJobStatus updates the status of a job in cache
func (q *RedisQueue) UpdateJobStatus(ctx context.Context, jobID uuid.UUID, status entity.JobStatus, progress int) error {
	statusData := map[string]interface{}{
		"status":     string(status),
		"progress":   progress,
		"updated_at": time.Now().Unix(),
	}

	data, err := json.Marshal(statusData)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	statusKey := jobStatusPrefix + jobID.String()
	if err := q.client.Set(ctx, statusKey, data, 5*time.Minute).Err(); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

// GetJobStatus retrieves the cached status of a job
func (q *RedisQueue) GetJobStatus(ctx context.Context, jobID uuid.UUID) (entity.JobStatus, int, error) {
	statusKey := jobStatusPrefix + jobID.String()
	data, err := q.client.Get(ctx, statusKey).Result()
	if err == redis.Nil {
		return "", 0, nil
	}
	if err != nil {
		return "", 0, fmt.Errorf("failed to get status: %w", err)
	}

	var statusData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &statusData); err != nil {
		return "", 0, fmt.Errorf("failed to unmarshal status: %w", err)
	}

	status := entity.JobStatus(statusData["status"].(string))
	progress := int(statusData["progress"].(float64))

	return status, progress, nil
}

// PeekQueue returns jobs in the queue without removing them
func (q *RedisQueue) PeekQueue(ctx context.Context, count int) ([]string, error) {
	result, err := q.client.ZRange(ctx, jobQueueKey, 0, int64(count-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to peek queue: %w", err)
	}

	return result, nil
}

