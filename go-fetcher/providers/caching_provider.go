package providers

import (
	"context"
	"fmt"
	"time"
	"go-fetcher/types"

	"github.com/patrickmn/go-cache"
)

// CachingSubgraphProvider wraps a provider with caching functionality
type CachingSubgraphProvider struct {
	BaseProvider
	provider SubgraphProvider
	cache    *cache.Cache
}

// NewCachingSubgraphProvider creates a new caching provider
func NewCachingSubgraphProvider(chainID types.ChainID, protocol types.Protocol, provider SubgraphProvider, defaultExpiration time.Duration) *CachingSubgraphProvider {
	return &CachingSubgraphProvider{
		BaseProvider: NewBaseProvider(chainID, protocol),
		provider:     provider,
		cache:        cache.New(defaultExpiration, time.Minute),
	}
}

// GetPools gets pools with caching
func (p *CachingSubgraphProvider) GetPools(ctx context.Context, config *ProviderConfig) ([]types.Pool, error) {
	cacheKey := p.generateCacheKey(config)
	
	// Try to get from cache first
	if cached, found := p.cache.Get(cacheKey); found {
		if pools, ok := cached.([]types.Pool); ok {
			return pools, nil
		}
	}

	// If not in cache, fetch from provider
	pools, err := p.provider.GetPools(ctx, config)
	if err != nil {
		return nil, err
	}

	// Store in cache
	p.cache.Set(cacheKey, pools, cache.DefaultExpiration)

	return pools, nil
}

// GetV2Pools gets V2 pools with caching
func (p *CachingSubgraphProvider) GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error) {
	if p.protocol != types.V2 {
		return nil, fmt.Errorf("provider is not configured for V2 protocol")
	}

	cacheKey := p.generateCacheKey(config) + "_v2"
	
	// Try to get from cache first
	if cached, found := p.cache.Get(cacheKey); found {
		if pools, ok := cached.([]types.V2Pool); ok {
			return pools, nil
		}
	}

	// If not in cache, fetch from provider
	if v2Provider, ok := p.provider.(V2SubgraphProvider); ok {
		pools, err := v2Provider.GetV2Pools(ctx, config)
		if err != nil {
			return nil, err
		}

		// Store in cache
		p.cache.Set(cacheKey, pools, cache.DefaultExpiration)

		return pools, nil
	}

	return nil, fmt.Errorf("underlying provider does not support V2 pools")
}

// GetV3Pools gets V3 pools with caching
func (p *CachingSubgraphProvider) GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error) {
	if p.protocol != types.V3 {
		return nil, fmt.Errorf("provider is not configured for V3 protocol")
	}

	cacheKey := p.generateCacheKey(config) + "_v3"
	
	// Try to get from cache first
	if cached, found := p.cache.Get(cacheKey); found {
		if pools, ok := cached.([]types.V3Pool); ok {
			return pools, nil
		}
	}

	// If not in cache, fetch from provider
	if v3Provider, ok := p.provider.(V3SubgraphProvider); ok {
		pools, err := v3Provider.GetV3Pools(ctx, config)
		if err != nil {
			return nil, err
		}

		// Store in cache
		p.cache.Set(cacheKey, pools, cache.DefaultExpiration)

		return pools, nil
	}

	return nil, fmt.Errorf("underlying provider does not support V3 pools")
}

// GetV4Pools gets V4 pools with caching
func (p *CachingSubgraphProvider) GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error) {
	if p.protocol != types.V4 {
		return nil, fmt.Errorf("provider is not configured for V4 protocol")
	}

	cacheKey := p.generateCacheKey(config) + "_v4"
	
	// Try to get from cache first
	if cached, found := p.cache.Get(cacheKey); found {
		if pools, ok := cached.([]types.V4Pool); ok {
			return pools, nil
		}
	}

	// If not in cache, fetch from provider
	if v4Provider, ok := p.provider.(V4SubgraphProvider); ok {
		pools, err := v4Provider.GetV4Pools(ctx, config)
		if err != nil {
			return nil, err
		}

		// Store in cache
		p.cache.Set(cacheKey, pools, cache.DefaultExpiration)

		return pools, nil
	}

	return nil, fmt.Errorf("underlying provider does not support V4 pools")
}

// generateCacheKey generates a cache key based on chain ID, protocol, and config
func (p *CachingSubgraphProvider) generateCacheKey(config *ProviderConfig) string {
	key := fmt.Sprintf("pools_%d_%s", p.chainID, p.protocol)
	
	if config != nil {
		if config.BlockNumber != nil {
			key += fmt.Sprintf("_block_%d", *config.BlockNumber)
		}
	}
	
	return key
}

// ClearCache clears the cache
func (p *CachingSubgraphProvider) ClearCache() {
	p.cache.Flush()
}

// GetCacheStats returns cache statistics
func (p *CachingSubgraphProvider) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["itemCount"] = p.cache.ItemCount()
	return stats
}
