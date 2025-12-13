package handler

import (
	"net/http"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TemplateHandler handles template endpoints
type TemplateHandler struct {
	templateUseCase *usecase.TemplateUseCase
}

// NewTemplateHandler creates a new TemplateHandler
func NewTemplateHandler(templateUseCase *usecase.TemplateUseCase) *TemplateHandler {
	return &TemplateHandler{
		templateUseCase: templateUseCase,
	}
}

// ListTemplates retrieves a paginated list of templates
// @Summary List templates
// @Description Get a paginated list of video templates
// @Tags templates
// @Accept json
// @Produce json
// @Param category query string false "Filter by category"
// @Param search query string false "Search in name and description"
// @Param premium query bool false "Filter by premium status"
// @Param sort_by query string false "Sort field (name, created_at, usage_count)"
// @Param order query string false "Sort order (asc, desc)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} usecase.TemplateListResponse
// @Failure 400 {object} ErrorResponse
// @Router /templates [get]
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	var req usecase.TemplateListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid query parameters",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	response, err := h.templateUseCase.GetTemplates(c.Request.Context(), req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTemplate retrieves a template by ID
// @Summary Get template
// @Description Get a video template by ID
// @Tags templates
// @Accept json
// @Produce json
// @Param id path string true "Template ID" format(uuid)
// @Success 200 {object} entity.Template
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /templates/{id} [get]
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid template ID",
			Code:  "INVALID_ID",
		})
		return
	}

	template, err := h.templateUseCase.GetTemplateByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

// GetTemplatesByCategory retrieves templates by category
// @Summary Get templates by category
// @Description Get video templates filtered by category
// @Tags templates
// @Accept json
// @Produce json
// @Param category path string true "Category name"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} usecase.TemplateListResponse
// @Failure 400 {object} ErrorResponse
// @Router /templates/category/{category} [get]
func (h *TemplateHandler) GetTemplatesByCategory(c *gin.Context) {
	category := c.Param("category")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("page_size", "20")

	pageInt := 1
	pageSizeInt := 20
	if _, err := c.GetQuery("page"); err {
		if p, err := parsePositiveInt(page); err == nil {
			pageInt = p
		}
	}
	if _, err := c.GetQuery("page_size"); err {
		if ps, err := parsePositiveInt(pageSize); err == nil {
			pageSizeInt = ps
		}
	}

	response, err := h.templateUseCase.GetTemplatesByCategory(
		c.Request.Context(),
		entity.TemplateCategory(category),
		pageInt,
		pageSizeInt,
	)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPopularTemplates retrieves popular templates
// @Summary Get popular templates
// @Description Get the most popular video templates
// @Tags templates
// @Accept json
// @Produce json
// @Param limit query int false "Number of templates to return" default(10)
// @Success 200 {array} entity.Template
// @Router /templates/popular [get]
func (h *TemplateHandler) GetPopularTemplates(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if l, err := parsePositiveInt(limitStr); err == nil {
		limit = l
	}

	templates, err := h.templateUseCase.GetPopularTemplates(c.Request.Context(), limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": templates,
	})
}

// GetCategories retrieves all categories
// @Summary Get categories
// @Description Get all available template categories
// @Tags templates
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Router /templates/categories [get]
func (h *TemplateHandler) GetCategories(c *gin.Context) {
	categories, err := h.templateUseCase.GetCategories(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"categories": categories,
	})
}

// parsePositiveInt parses a string to a positive integer
func parsePositiveInt(s string) (int, error) {
	var i int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, entity.ErrInvalidInput
		}
		i = i*10 + int(c-'0')
	}
	if i <= 0 {
		return 0, entity.ErrInvalidInput
	}
	return i, nil
}

