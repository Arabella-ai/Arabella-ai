package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/service"
	"go.uber.org/zap"
)

const (
	geminiBaseURL = "https://generativelanguage.googleapis.com/v1beta"
)

// GeminiProvider implements the Gemini VEO video generation provider
type GeminiProvider struct {
	*BaseProvider
}

// GeminiGenerateRequest represents a Gemini video generation request
type GeminiGenerateRequest struct {
	Model    string         `json:"model"`
	Contents []GeminiContent `json:"contents"`
	Config   GeminiConfig   `json:"generationConfig"`
}

// GeminiContent represents content in a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

// GeminiPart represents a part of content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiConfig represents generation configuration
type GeminiConfig struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
}

// GeminiGenerateResponse represents a Gemini generation response
type GeminiGenerateResponse struct {
	Name      string `json:"name"`
	Done      bool   `json:"done"`
	Error     *GeminiError `json:"error,omitempty"`
	Response  *GeminiVideoResponse `json:"response,omitempty"`
}

// GeminiError represents an error from Gemini
type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// GeminiVideoResponse represents the video response
type GeminiVideoResponse struct {
	VideoURL     string `json:"videoUri"`
	ThumbnailURL string `json:"thumbnailUri"`
	Duration     int    `json:"durationSeconds"`
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, logger *zap.Logger) service.VideoProvider {
	return &GeminiProvider{
		BaseProvider: NewBaseProvider(apiKey, geminiBaseURL, 5*time.Minute, logger),
	}
}

// GetName returns the provider name
func (p *GeminiProvider) GetName() entity.AIProvider {
	return entity.ProviderGeminiVEO
}

// GenerateVideo initiates video generation with Gemini VEO
func (p *GeminiProvider) GenerateVideo(ctx context.Context, req service.GenerationRequest) (*entity.GenerationResult, error) {
	// Build the request
	geminiReq := GeminiGenerateRequest{
		Model: "gemini-2.0-flash-exp",
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: req.Prompt},
				},
			},
		},
		Config: GeminiConfig{
			Temperature: 0.7,
		},
	}

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the API request
	url := fmt.Sprintf("%s/models/gemini-2.0-flash-exp:generateContent?key=%s", p.baseURL, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		p.logger.Error("Gemini API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(bodyBytes)),
		)
		return nil, fmt.Errorf("Gemini API error: %d", resp.StatusCode)
	}

	var geminiResp GeminiGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini error: %s", geminiResp.Error.Message)
	}

	return &entity.GenerationResult{
		ProviderJobID: geminiResp.Name,
		VideoURL:      geminiResp.Response.VideoURL,
		ThumbnailURL:  geminiResp.Response.ThumbnailURL,
		Duration:      geminiResp.Response.Duration,
	}, nil
}

// GetProgress retrieves generation progress
func (p *GeminiProvider) GetProgress(ctx context.Context, providerJobID string) (*entity.Progress, error) {
	// Make the API request to check operation status
	url := fmt.Sprintf("%s/operations/%s?key=%s", p.baseURL, providerJobID, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var opResp GeminiGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&opResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if opResp.Done {
		return &entity.Progress{
			Percent: 100,
			Stage:   "COMPLETED",
			Message: "Video generation completed",
		}, nil
	}

	return &entity.Progress{
		Percent: 50, // Estimate
		Stage:   "PROCESSING",
		Message: "Video generation in progress",
	}, nil
}

// CancelGeneration cancels an ongoing generation
func (p *GeminiProvider) CancelGeneration(ctx context.Context, providerJobID string) error {
	// Make the API request to cancel
	url := fmt.Sprintf("%s/operations/%s:cancel?key=%s", p.baseURL, providerJobID, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to cancel: status %d", resp.StatusCode)
	}

	return nil
}

// GetCapabilities returns provider capabilities
func (p *GeminiProvider) GetCapabilities() service.ProviderCapabilities {
	return service.ProviderCapabilities{
		Name:            "Gemini VEO",
		MaxDuration:     120,
		MaxResolution:   entity.Resolution4K,
		SupportedRatios: []entity.AspectRatio{entity.AspectRatio16x9, entity.AspectRatio9x16, entity.AspectRatio1x1},
		EstimatedTime:   30,
		QualityTier:     "premium",
		SupportsStyles:  true,
		CostPerSecond:   0.05,
	}
}

// HealthCheck performs a health check
func (p *GeminiProvider) HealthCheck(ctx context.Context) (*service.ProviderHealth, error) {
	// Make a simple request to check if the API is available
	url := fmt.Sprintf("%s/models?key=%s", p.baseURL, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	responseTime := time.Since(start).Milliseconds()

	if err != nil {
		return &service.ProviderHealth{
			IsHealthy:    false,
			ResponseTime: responseTime,
			ErrorRate:    1.0,
			LastChecked:  time.Now().Unix(),
		}, nil
	}
	defer resp.Body.Close()

	isHealthy := resp.StatusCode == http.StatusOK

	return &service.ProviderHealth{
		IsHealthy:    isHealthy,
		QueueDepth:   0, // Would need to be tracked separately
		ResponseTime: responseTime,
		ErrorRate:    0.0,
		LastChecked:  time.Now().Unix(),
	}, nil
}

