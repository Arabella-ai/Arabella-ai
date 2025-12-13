package handler

import (
	"net/http"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/interface/http/middleware"
	"github.com/arabella/ai-studio-backend/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// VideoHandler handles video generation endpoints
type VideoHandler struct {
	videoUseCase *usecase.VideoUseCase
}

// NewVideoHandler creates a new VideoHandler
func NewVideoHandler(videoUseCase *usecase.VideoUseCase) *VideoHandler {
	return &VideoHandler{
		videoUseCase: videoUseCase,
	}
}

// GenerateVideoRequest represents the video generation request
type GenerateVideoRequest struct {
	TemplateID string              `json:"template_id" binding:"required,uuid"`
	Prompt     string              `json:"prompt" binding:"required,min=10,max=2000"`
	Params     *VideoParamsRequest `json:"params,omitempty"`
}

// VideoParamsRequest represents video generation parameters
type VideoParamsRequest struct {
	Duration       int    `json:"duration,omitempty"`
	Resolution     string `json:"resolution,omitempty"`
	AspectRatio    string `json:"aspect_ratio,omitempty"`
	FPS            int    `json:"fps,omitempty"`
	Style          string `json:"style,omitempty"`
	NegativePrompt string `json:"negative_prompt,omitempty"`
}

// GenerateVideo initiates video generation
// @Summary Generate video
// @Description Start a new AI video generation job
// @Tags videos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body GenerateVideoRequest true "Video Generation Request"
// @Success 201 {object} usecase.VideoGenerationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 429 {object} ErrorResponse
// @Router /videos/generate [post]
func (h *VideoHandler) GenerateVideo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	var req GenerateVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	templateID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid template ID",
			Code:  "INVALID_ID",
		})
		return
	}

	// Build use case request
	useCaseReq := usecase.VideoGenerationRequest{
		TemplateID: templateID,
		Prompt:     req.Prompt,
	}

	if req.Params != nil {
		useCaseReq.Params = convertVideoParams(req.Params)
	} else {
		useCaseReq.Params = nil
	}

	response, err := h.videoUseCase.GenerateVideo(c.Request.Context(), userID, useCaseReq)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// GetJobStatus retrieves the status of a video job
// @Summary Get job status
// @Description Get the current status of a video generation job
// @Tags videos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID" format(uuid)
// @Success 200 {object} entity.VideoJob
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /videos/{id}/status [get]
func (h *VideoHandler) GetJobStatus(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid job ID",
			Code:  "INVALID_ID",
		})
		return
	}

	job, err := h.videoUseCase.GetJobStatus(c.Request.Context(), userID, jobID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, job)
}

// GetVideo retrieves a completed video
// @Summary Get video
// @Description Get details and URL of a completed video
// @Tags videos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID" format(uuid)
// @Success 200 {object} entity.VideoJob
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /videos/{id} [get]
func (h *VideoHandler) GetVideo(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid job ID",
			Code:  "INVALID_ID",
		})
		return
	}

	job, err := h.videoUseCase.GetVideo(c.Request.Context(), userID, jobID)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, job)
}

// ListUserVideos retrieves all videos for the authenticated user
// @Summary List user videos
// @Description Get a paginated list of videos for the authenticated user
// @Tags videos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Items per page" default(20)
// @Success 200 {object} usecase.VideoJobListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /videos [get]
func (h *VideoHandler) ListUserVideos(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	var req usecase.VideoJobListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid query parameters",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	response, err := h.videoUseCase.GetUserJobs(c.Request.Context(), userID, req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CancelJob cancels a video generation job
// @Summary Cancel job
// @Description Cancel a pending or processing video generation job
// @Tags videos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Job ID" format(uuid)
// @Success 204
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Router /videos/{id}/cancel [post]
func (h *VideoHandler) CancelJob(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	jobID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid job ID",
			Code:  "INVALID_ID",
		})
		return
	}

	if err := h.videoUseCase.CancelJob(c.Request.Context(), userID, jobID); err != nil {
		handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetRecentVideos retrieves recent videos for the authenticated user
// @Summary Get recent videos
// @Description Get the most recent videos for the authenticated user
// @Tags videos
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param limit query int false "Number of videos to return" default(10)
// @Success 200 {array} entity.VideoJob
// @Failure 401 {object} ErrorResponse
// @Router /videos/recent [get]
func (h *VideoHandler) GetRecentVideos(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "Authentication required",
			Code:  "UNAUTHORIZED",
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit := 10
	if l, err := parsePositiveInt(limitStr); err == nil {
		limit = l
	}

	jobs, err := h.videoUseCase.GetRecentJobs(c.Request.Context(), userID, limit)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"videos": jobs,
	})
}

// convertVideoParams converts request params to entity params
func convertVideoParams(req *VideoParamsRequest) *entity.VideoParams {
	if req == nil {
		return nil
	}
	return &entity.VideoParams{
		Duration:       req.Duration,
		Resolution:     entity.VideoResolution(req.Resolution),
		AspectRatio:    entity.AspectRatio(req.AspectRatio),
		FPS:            req.FPS,
		Style:          req.Style,
		NegativePrompt: req.NegativePrompt,
	}
}

