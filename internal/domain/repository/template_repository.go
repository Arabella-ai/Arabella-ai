package repository

import (
	"context"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/google/uuid"
)

// TemplateFilter defines filters for template queries
type TemplateFilter struct {
	Category  *entity.TemplateCategory
	IsPremium *bool
	IsActive  *bool
	Tags      []string
	Search    string
}

// TemplateSort defines sorting options for templates
type TemplateSort struct {
	Field     string // name, created_at, usage_count
	Direction string // asc, desc
}

// TemplateRepository defines the interface for template data access
type TemplateRepository interface {
	// Create creates a new template
	Create(ctx context.Context, template *entity.Template) error

	// GetByID retrieves a template by ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Template, error)

	// Update updates an existing template
	Update(ctx context.Context, template *entity.Template) error

	// Delete deletes a template (soft delete by setting IsActive to false)
	Delete(ctx context.Context, id uuid.UUID) error

	// List lists templates with filtering and pagination
	List(ctx context.Context, filter TemplateFilter, sort TemplateSort, offset, limit int) ([]*entity.Template, int64, error)

	// ListByCategory lists templates by category
	ListByCategory(ctx context.Context, category entity.TemplateCategory, offset, limit int) ([]*entity.Template, int64, error)

	// GetPopular retrieves the most popular templates
	GetPopular(ctx context.Context, limit int) ([]*entity.Template, error)

	// IncrementUsage increments the usage count for a template
	IncrementUsage(ctx context.Context, id uuid.UUID) error

	// GetCategories retrieves all unique categories
	GetCategories(ctx context.Context) ([]entity.TemplateCategory, error)
}

