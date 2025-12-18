package usecase

import (
	"context"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/repository"
	"github.com/google/uuid"
)

// TemplateListRequest represents a request to list templates
type TemplateListRequest struct {
	Category string `form:"category"`
	Search   string `form:"search"`
	Premium  *bool  `form:"premium"`
	Tags     string `form:"tags"`
	SortBy   string `form:"sort_by"`
	Order    string `form:"order"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// TemplateListResponse represents the paginated template list response
type TemplateListResponse struct {
	Templates  []*entity.Template `json:"templates"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// TemplateUseCase handles template-related business logic
type TemplateUseCase struct {
	templateRepo repository.TemplateRepository
	cache        CacheService
}

// CacheService interface for caching
type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// NewTemplateUseCase creates a new TemplateUseCase
func NewTemplateUseCase(
	templateRepo repository.TemplateRepository,
	cache CacheService,
) *TemplateUseCase {
	return &TemplateUseCase{
		templateRepo: templateRepo,
		cache:        cache,
	}
}

// GetTemplates retrieves a paginated list of templates
func (uc *TemplateUseCase) GetTemplates(ctx context.Context, req TemplateListRequest) (*TemplateListResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.Order == "" {
		req.Order = "desc"
	}

	// Build filter
	filter := repository.TemplateFilter{
		IsActive: boolPtr(true),
	}

	if req.Category != "" {
		category := entity.TemplateCategory(req.Category)
		filter.Category = &category
	}

	if req.Premium != nil {
		filter.IsPremium = req.Premium
	}

	if req.Search != "" {
		filter.Search = req.Search
	}

	// Build sort
	sort := repository.TemplateSort{
		Field:     req.SortBy,
		Direction: req.Order,
	}

	// Calculate offset
	offset := (req.Page - 1) * req.PageSize

	// Get templates
	templates, total, err := uc.templateRepo.List(ctx, filter, sort, offset, req.PageSize)
	if err != nil {
		return nil, err
	}

	// Calculate total pages
	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	return &TemplateListResponse{
		Templates:  templates,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetTemplateByID retrieves a template by ID
func (uc *TemplateUseCase) GetTemplateByID(ctx context.Context, id uuid.UUID) (*entity.Template, error) {
	template, err := uc.templateRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !template.IsActive {
		return nil, entity.ErrTemplateNotActive
	}

	return template, nil
}

// GetTemplatesByCategory retrieves templates by category
func (uc *TemplateUseCase) GetTemplatesByCategory(ctx context.Context, category entity.TemplateCategory, page, pageSize int) (*TemplateListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	templates, total, err := uc.templateRepo.ListByCategory(ctx, category, offset, pageSize)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &TemplateListResponse{
		Templates:  templates,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetPopularTemplates retrieves the most popular templates
func (uc *TemplateUseCase) GetPopularTemplates(ctx context.Context, limit int) ([]*entity.Template, error) {
	if limit < 1 || limit > 50 {
		limit = 10
	}

	return uc.templateRepo.GetPopular(ctx, limit)
}

// GetCategories retrieves all available categories
func (uc *TemplateUseCase) GetCategories(ctx context.Context) ([]entity.TemplateCategory, error) {
	return uc.templateRepo.GetCategories(ctx)
}

// ValidateTemplateAccess validates if a user can use a template
func (uc *TemplateUseCase) ValidateTemplateAccess(ctx context.Context, templateID uuid.UUID, user *entity.User) (*entity.Template, error) {
	template, err := uc.templateRepo.GetByID(ctx, templateID)
	if err != nil {
		return nil, err
	}

	if !template.IsActive {
		return nil, entity.ErrTemplateNotActive
	}

	if template.IsPremium && !user.IsPremium() {
		return nil, entity.ErrTemplatePremiumOnly
	}

	if !user.HasSufficientCredits(template.CreditCost) {
		return nil, entity.ErrInsufficientCredits
	}

	return template, nil
}

// CreateTemplate creates a new template
func (uc *TemplateUseCase) CreateTemplate(ctx context.Context, template *entity.Template) error {
	if template.Name == "" {
		return entity.ErrInvalidInput
	}
	if template.Category == "" {
		return entity.ErrInvalidInput
	}
	if template.Description == "" {
		return entity.ErrInvalidInput
	}
	if template.BasePrompt == "" {
		return entity.ErrInvalidInput
	}
	return uc.templateRepo.Create(ctx, template)
}

// UpdateTemplate updates an existing template
func (uc *TemplateUseCase) UpdateTemplate(ctx context.Context, template *entity.Template) error {
	// Verify template exists
	_, err := uc.templateRepo.GetByID(ctx, template.ID)
	if err != nil {
		return err
	}
	return uc.templateRepo.Update(ctx, template)
}

// DeleteTemplate deletes a template (soft delete)
func (uc *TemplateUseCase) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	return uc.templateRepo.Delete(ctx, id)
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}
