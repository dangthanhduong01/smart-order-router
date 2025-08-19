package providers

import (
	"context"
	"fmt"
	"go-fetcher/types"
)

// FallbackSubgraphProvider tries multiple providers in order until one succeeds
type FallbackSubgraphProvider struct {
	BaseProvider
	providers []SubgraphProvider
}

// NewFallbackSubgraphProvider creates a new fallback provider
func NewFallbackSubgraphProvider(chainID types.ChainID, protocol types.Protocol, providers ...SubgraphProvider) *FallbackSubgraphProvider {
	return &FallbackSubgraphProvider{
		BaseProvider: NewBaseProvider(chainID, protocol),
		providers:    providers,
	}
}

// GetPools tries each provider in order until one succeeds
func (p *FallbackSubgraphProvider) GetPools(ctx context.Context, config *ProviderConfig) ([]types.Pool, error) {
	var lastErr error

	for i, provider := range p.providers {
		pools, err := provider.GetPools(ctx, config)
		if err == nil {
			return pools, nil
		}

		lastErr = fmt.Errorf("provider %d failed: %w", i, err)
	}

	return nil, fmt.Errorf("all providers failed. Last error: %w", lastErr)
}

// GetV2Pools tries each provider for V2 pools
func (p *FallbackSubgraphProvider) GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error) {
	if p.protocol != types.V2 {
		return nil, fmt.Errorf("provider is not configured for V2 protocol")
	}

	var lastErr error

	for i, provider := range p.providers {
		if v2Provider, ok := provider.(V2SubgraphProvider); ok {
			pools, err := v2Provider.GetV2Pools(ctx, config)
			if err == nil {
				return pools, nil
			}
			lastErr = fmt.Errorf("V2 provider %d failed: %w", i, err)
		}
	}

	return nil, fmt.Errorf("all V2 providers failed. Last error: %w", lastErr)
}

// GetV3Pools tries each provider for V3 pools
func (p *FallbackSubgraphProvider) GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error) {
	if p.protocol != types.V3 {
		return nil, fmt.Errorf("provider is not configured for V3 protocol")
	}

	var lastErr error

	for i, provider := range p.providers {
		if v3Provider, ok := provider.(V3SubgraphProvider); ok {
			pools, err := v3Provider.GetV3Pools(ctx, config)
			if err == nil {
				return pools, nil
			}
			lastErr = fmt.Errorf("V3 provider %d failed: %w", i, err)
		}
	}

	return nil, fmt.Errorf("all V3 providers failed. Last error: %w", lastErr)
}

// GetV4Pools tries each provider for V4 pools
func (p *FallbackSubgraphProvider) GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error) {
	if p.protocol != types.V4 {
		return nil, fmt.Errorf("provider is not configured for V4 protocol")
	}

	var lastErr error

	for i, provider := range p.providers {
		if v4Provider, ok := provider.(V4SubgraphProvider); ok {
			pools, err := v4Provider.GetV4Pools(ctx, config)
			if err == nil {
				return pools, nil
			}
			lastErr = fmt.Errorf("V4 provider %d failed: %w", i, err)
		}
	}

	return nil, fmt.Errorf("all V4 providers failed. Last error: %w", lastErr)
}
