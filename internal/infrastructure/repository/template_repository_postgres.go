package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TemplateRepositoryPostgres implements TemplateRepository for PostgreSQL
type TemplateRepositoryPostgres struct {
	pool *pgxpool.Pool
}

// NewTemplateRepositoryPostgres creates a new TemplateRepositoryPostgres
func NewTemplateRepositoryPostgres(pool *pgxpool.Pool) repository.TemplateRepository {
	return &TemplateRepositoryPostgres{pool: pool}
}

// Create creates a new template
func (r *TemplateRepositoryPostgres) Create(ctx context.Context, template *entity.Template) error {
	query := `
		INSERT INTO templates (id, name, category, description, thumbnail_url, preview_video_url,
		                       base_prompt, default_params, credit_cost, estimated_time_seconds,
		                       is_premium, is_active, preferred_provider, tags, usage_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`

	paramsJSON, err := json.Marshal(template.DefaultParams)
	if err != nil {
		return err
	}

	_, err = r.pool.Exec(ctx, query,
		template.ID,
		template.Name,
		template.Category,
		template.Description,
		template.ThumbnailURL,
		template.PreviewVideoURL,
		template.BasePrompt,
		paramsJSON,
		template.CreditCost,
		int(template.EstimatedTime.Seconds()),
		template.IsPremium,
		template.IsActive,
		template.PreferredProvider,
		template.Tags,
		template.UsageCount,
		template.CreatedAt,
		template.UpdatedAt,
	)

	return err
}

// GetByID retrieves a template by ID
func (r *TemplateRepositoryPostgres) GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error) {
	query := `
		SELECT id, name, category, description, thumbnail_url, preview_video_url,
		       base_prompt, default_params, credit_cost, estimated_time_seconds,
		       is_premium, is_active, preferred_provider, tags, usage_count, created_at, updated_at
		FROM templates
		WHERE id = $1
	`

	template := &entity.Template{}
	var paramsJSON []byte
	var estimatedTimeSeconds int

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Category,
		&template.Description,
		&template.ThumbnailURL,
		&template.PreviewVideoURL,
		&template.BasePrompt,
		&paramsJSON,
		&template.CreditCost,
		&estimatedTimeSeconds,
		&template.IsPremium,
		&template.IsActive,
		&template.PreferredProvider,
		&template.Tags,
		&template.UsageCount,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, entity.ErrTemplateNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(paramsJSON, &template.DefaultParams); err != nil {
		return nil, err
	}

	template.EstimatedTime = time.Duration(estimatedTimeSeconds) * time.Second

	return template, nil
}

// Update updates an existing template
func (r *TemplateRepositoryPostgres) Update(ctx context.Context, template *entity.Template) error {
	query := `
		UPDATE templates
		SET name = $2, category = $3, description = $4, thumbnail_url = $5, preview_video_url = $6,
		    base_prompt = $7, default_params = $8, credit_cost = $9, estimated_time_seconds = $10,
		    is_premium = $11, is_active = $12, preferred_provider = $13, tags = $14, 
		    usage_count = $15, updated_at = $16
		WHERE id = $1
	`

	paramsJSON, err := json.Marshal(template.DefaultParams)
	if err != nil {
		return err
	}

	template.UpdatedAt = time.Now()

	result, err := r.pool.Exec(ctx, query,
		template.ID,
		template.Name,
		template.Category,
		template.Description,
		template.ThumbnailURL,
		template.PreviewVideoURL,
		template.BasePrompt,
		paramsJSON,
		template.CreditCost,
		int(template.EstimatedTime.Seconds()),
		template.IsPremium,
		template.IsActive,
		template.PreferredProvider,
		template.Tags,
		template.UsageCount,
		template.UpdatedAt,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrTemplateNotFound
	}

	return nil
}

// Delete soft deletes a template
func (r *TemplateRepositoryPostgres) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE templates SET is_active = false, updated_at = $2 WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id, time.Now())
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return entity.ErrTemplateNotFound
	}

	return nil
}

// List lists templates with filtering and pagination
func (r *TemplateRepositoryPostgres) List(ctx context.Context, filter repository.TemplateFilter, sort repository.TemplateSort, offset, limit int) ([]*entity.Template, int64, error) {
	// Build WHERE clause
	conditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.Category != nil {
		conditions = append(conditions, "category = $"+string(rune('0'+argIndex)))
		args = append(args, *filter.Category)
		argIndex++
	}

	if filter.IsPremium != nil {
		conditions = append(conditions, "is_premium = $"+string(rune('0'+argIndex)))
		args = append(args, *filter.IsPremium)
		argIndex++
	}

	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = $"+string(rune('0'+argIndex)))
		args = append(args, *filter.IsActive)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, "(name ILIKE $"+string(rune('0'+argIndex))+" OR description ILIKE $"+string(rune('0'+argIndex))+")")
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := "SELECT COUNT(*) FROM templates " + whereClause
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Build ORDER BY clause
	orderBy := "created_at DESC"
	if sort.Field != "" {
		direction := "ASC"
		if sort.Direction == "desc" {
			direction = "DESC"
		}
		orderBy = sort.Field + " " + direction
	}

	// Get templates
	query := `
		SELECT id, name, category, description, thumbnail_url, preview_video_url,
		       base_prompt, default_params, credit_cost, estimated_time_seconds,
		       is_premium, is_active, preferred_provider, tags, usage_count, created_at, updated_at
		FROM templates
		` + whereClause + `
		ORDER BY ` + orderBy + `
		LIMIT $` + string(rune('0'+argIndex)) + ` OFFSET $` + string(rune('0'+argIndex+1))

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	return r.scanTemplates(rows)
}

// ListByCategory lists templates by category
func (r *TemplateRepositoryPostgres) ListByCategory(ctx context.Context, category entity.TemplateCategory, offset, limit int) ([]*entity.Template, int64, error) {
	filter := repository.TemplateFilter{
		Category: &category,
		IsActive: boolPtr(true),
	}
	sort := repository.TemplateSort{Field: "created_at", Direction: "desc"}
	return r.List(ctx, filter, sort, offset, limit)
}

// GetPopular retrieves the most popular templates
func (r *TemplateRepositoryPostgres) GetPopular(ctx context.Context, limit int) ([]*entity.Template, error) {
	query := `
		SELECT id, name, category, description, thumbnail_url, preview_video_url,
		       base_prompt, default_params, credit_cost, estimated_time_seconds,
		       is_premium, is_active, preferred_provider, tags, usage_count, created_at, updated_at
		FROM templates
		WHERE is_active = true
		ORDER BY usage_count DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates, _, err := r.scanTemplates(rows)
	return templates, err
}

// IncrementUsage increments the usage count
func (r *TemplateRepositoryPostgres) IncrementUsage(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE templates SET usage_count = usage_count + 1 WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

// GetCategories retrieves all unique categories
func (r *TemplateRepositoryPostgres) GetCategories(ctx context.Context) ([]entity.TemplateCategory, error) {
	query := `SELECT DISTINCT category FROM templates WHERE is_active = true ORDER BY category`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []entity.TemplateCategory
	for rows.Next() {
		var category entity.TemplateCategory
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// scanTemplates scans rows into templates
func (r *TemplateRepositoryPostgres) scanTemplates(rows pgx.Rows) ([]*entity.Template, int64, error) {
	var templates []*entity.Template
	for rows.Next() {
		template := &entity.Template{}
		var paramsJSON []byte
		var estimatedTimeSeconds int

		err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Category,
			&template.Description,
			&template.ThumbnailURL,
			&template.PreviewVideoURL,
			&template.BasePrompt,
			&paramsJSON,
			&template.CreditCost,
			&estimatedTimeSeconds,
			&template.IsPremium,
			&template.IsActive,
			&template.PreferredProvider,
			&template.Tags,
			&template.UsageCount,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if err := json.Unmarshal(paramsJSON, &template.DefaultParams); err != nil {
			return nil, 0, err
		}

		template.EstimatedTime = time.Duration(estimatedTimeSeconds) * time.Second
		templates = append(templates, template)
	}

	return templates, int64(len(templates)), nil
}

func boolPtr(b bool) *bool {
	return &b
}

