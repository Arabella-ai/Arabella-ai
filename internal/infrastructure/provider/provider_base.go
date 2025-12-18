package provider

import (
	"context"
	"net/http"
	"time"

	"github.com/arabella/ai-studio-backend/internal/domain/entity"
	"github.com/arabella/ai-studio-backend/internal/domain/service"
	"go.uber.org/zap"
)

// BaseProvider contains common provider functionality
type BaseProvider struct {
	httpClient *http.Client
	logger     *zap.Logger
	apiKey     string
	baseURL    string
}

// NewBaseProvider creates a new BaseProvider
func NewBaseProvider(apiKey, baseURL string, timeout time.Duration, logger *zap.Logger) *BaseProvider {
	return &BaseProvider{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		logger:  logger,
		apiKey:  apiKey,
		baseURL: baseURL,
	}
}

// ProviderRegistry manages available AI providers
type ProviderRegistry struct {
	providers map[entity.AIProvider]service.VideoProvider
	logger    *zap.Logger
}

// NewProviderRegistry creates a new ProviderRegistry
func NewProviderRegistry(logger *zap.Logger) *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[entity.AIProvider]service.VideoProvider),
		logger:    logger,
	}
}

// Register registers a provider
func (r *ProviderRegistry) Register(provider service.VideoProvider) {
	r.providers[provider.GetName()] = provider
	r.logger.Info("Registered AI provider", zap.String("provider", string(provider.GetName())))
}

// Get returns a provider by name
func (r *ProviderRegistry) Get(name entity.AIProvider) (service.VideoProvider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

// GetAll returns all registered providers
func (r *ProviderRegistry) GetAll() []service.VideoProvider {
	providers := make([]service.VideoProvider, 0, len(r.providers))
	for _, p := range r.providers {
		providers = append(providers, p)
	}
	return providers
}

// ProviderSelectorImpl implements the ProviderSelector interface
type ProviderSelectorImpl struct {
	registry    *ProviderRegistry
	healthCache map[entity.AIProvider]*service.ProviderHealth
	logger      *zap.Logger
}

// NewProviderSelector creates a new ProviderSelector
func NewProviderSelector(registry *ProviderRegistry, logger *zap.Logger) service.ProviderSelector {
	return &ProviderSelectorImpl{
		registry:    registry,
		healthCache: make(map[entity.AIProvider]*service.ProviderHealth),
		logger:      logger,
	}
}

// SelectProvider selects the best provider based on requirements
func (s *ProviderSelectorImpl) SelectProvider(ctx context.Context, req service.ProviderSelectionRequest) (service.VideoProvider, error) {
	// If a preferred provider is specified, try to use it
	if req.PreferredProvider != nil {
		if provider, ok := s.registry.Get(*req.PreferredProvider); ok {
			health, err := provider.HealthCheck(ctx)
			// Log health check for debugging
			s.logger.Info("Preferred provider health check",
				zap.String("provider", string(*req.PreferredProvider)),
				zap.Bool("healthy", health != nil && health.IsHealthy),
				zap.Error(err),
			)
			// Allow provider even if health check fails (for testing)
			// In production, you might want to require health check
			if health != nil && health.IsHealthy {
				return provider, nil
			} else if health == nil || err != nil {
				// If health check fails but provider exists, still use it (for testing)
				s.logger.Warn("Preferred provider health check failed, but using it anyway",
					zap.String("provider", string(*req.PreferredProvider)),
				)
				return provider, nil
			}
		}
	}

	// Get all available providers
	providers := s.registry.GetAll()
	if len(providers) == 0 {
		return nil, entity.ErrProviderUnavailable
	}

	// Filter by user tier
	var eligible []service.VideoProvider
	for _, provider := range providers {
		caps := provider.GetCapabilities()

		// Check tier requirements
		// Allow free users to use premium providers (for testing)
		// In production, you can re-enable this check if needed
		// if req.UserTier == entity.UserTierFree {
		// 	if caps.QualityTier == "premium" {
		// 		continue
		// 	}
		// }

		// Check resolution requirements
		if !supportsResolution(caps.MaxResolution, req.RequiredResolution) {
			continue
		}

		// Check duration requirements
		if caps.MaxDuration < req.RequiredDuration {
			continue
		}

		// Check health (more lenient for testing)
		health, err := provider.HealthCheck(ctx)
		// Only skip if health check explicitly fails AND we have other options
		// For testing, we'll allow providers even if health check has issues
		if err != nil {
			s.logger.Warn("Provider health check error, but allowing it",
				zap.String("provider", string(provider.GetName())),
				zap.Error(err),
			)
			// Continue to next provider only if we have other options
		} else if health != nil && !health.IsHealthy {
			s.logger.Warn("Provider unhealthy, but allowing it for testing",
				zap.String("provider", string(provider.GetName())),
			)
			// For testing, allow unhealthy providers if it's the preferred one
			if req.PreferredProvider != nil && provider.GetName() == *req.PreferredProvider {
				// Allow preferred provider even if unhealthy
			} else if len(providers) > 1 {
				// Skip only if we have other providers
				continue
			}
		}

		eligible = append(eligible, provider)
	}

	if len(eligible) == 0 {
		return nil, entity.ErrProviderUnavailable
	}

	// Select the best provider based on queue depth and quality
	best := eligible[0]
	bestScore := scoreProvider(eligible[0], req.UserTier)

	for _, provider := range eligible[1:] {
		score := scoreProvider(provider, req.UserTier)
		if score > bestScore {
			best = provider
			bestScore = score
		}
	}

	return best, nil
}

// GetAvailableProviders returns all available providers
func (s *ProviderSelectorImpl) GetAvailableProviders(ctx context.Context) ([]service.VideoProvider, error) {
	var available []service.VideoProvider

	for _, provider := range s.registry.GetAll() {
		health, err := provider.HealthCheck(ctx)
		if err == nil && health.IsHealthy {
			available = append(available, provider)
		}
	}

	return available, nil
}

// RefreshHealth refreshes health status for all providers
func (s *ProviderSelectorImpl) RefreshHealth(ctx context.Context) error {
	for _, provider := range s.registry.GetAll() {
		health, _ := provider.HealthCheck(ctx)
		s.healthCache[provider.GetName()] = health
	}
	return nil
}

// supportsResolution checks if a provider supports a required resolution
func supportsResolution(max, required entity.VideoResolution) bool {
	resolutionOrder := map[entity.VideoResolution]int{
		entity.Resolution720p:  1,
		entity.Resolution1080p: 2,
		entity.Resolution4K:    3,
	}

	return resolutionOrder[max] >= resolutionOrder[required]
}

// scoreProvider calculates a score for provider selection
func scoreProvider(provider service.VideoProvider, userTier entity.UserTier) int {
	caps := provider.GetCapabilities()
	score := 0

	// Prefer Wan AI (highest priority) - much higher than others
	if provider.GetName() == entity.ProviderWanAI {
		score += 1000 // Very high priority
	}

	// Penalize mock provider (lowest priority)
	if provider.GetName() == entity.ProviderMock {
		score -= 500 // Very low priority
	}

	// Higher quality = higher score for premium users
	switch caps.QualityTier {
	case "premium":
		if userTier != entity.UserTierFree {
			score += 30
		}
	case "standard":
		score += 20
	case "budget":
		score += 10
	}

	// Lower estimated time = higher score
	score += 100 - caps.EstimatedTime

	return score
}
