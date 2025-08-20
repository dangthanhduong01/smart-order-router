package main

import (
	"context"
	"fmt"
	"go-fetcher/config"
	"go-fetcher/providers"
	"go-fetcher/types"
	"time"

	"github.com/sirupsen/logrus"
)

// PoolFetcher manages all pool providers and provides a unified interface
type PoolFetcher struct {
	chainID    types.ChainID
	v2Provider providers.SubgraphProvider
	v3Provider providers.SubgraphProvider
	v4Provider providers.SubgraphProvider
	logger     *logrus.Logger
}

// NewPoolFetcher creates a new pool fetcher for a specific chain
func NewPoolFetcher(chainID types.ChainID) (*PoolFetcher, error) {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fetcher := &PoolFetcher{
		chainID: chainID,
		logger:  logger,
	}

	// Initialize V2 provider
	if err := fetcher.initV2Provider(); err != nil {
		return nil, fmt.Errorf("failed to initialize V2 provider: %w", err)
	}

	// Initialize V3 provider
	if err := fetcher.initV3Provider(); err != nil {
		return nil, fmt.Errorf("failed to initialize V3 provider: %w", err)
	}

	// Initialize V4 provider
	if err := fetcher.initV4Provider(); err != nil {
		return nil, fmt.Errorf("failed to initialize V4 provider: %w", err)
	}

	return fetcher, nil
}

// initV2Provider initializes the V2 provider with fallback chain
func (f *PoolFetcher) initV2Provider() error {
	var v2Providers []providers.SubgraphProvider

	// Prefer GraphQL for completeness
	if subgraphURL, exists := config.GetSubgraphURL(f.chainID, types.V2); exists {
		graphqlProvider := providers.NewGraphQLSubgraphProvider(
			f.chainID,
			types.V2,
			subgraphURL,
			30*time.Second,
			2,
		)
		v2Providers = append(v2Providers, graphqlProvider)
	}

	// Then IPFS cache (fastest)
	if ipfsURL, exists := config.GetIPFSCacheURL(f.chainID, types.V2); exists {
		uriProvider := providers.NewURISubgraphProvider(
			f.chainID,
			types.V2,
			ipfsURL,
			30*time.Second,
			2,
		)
		cachingProvider := providers.NewCachingSubgraphProvider(
			f.chainID,
			types.V2,
			uriProvider,
			5*time.Minute,
		)
		v2Providers = append(v2Providers, cachingProvider)
	}

	// Add static provider as last resort
	staticProvider := providers.NewStaticSubgraphProvider(f.chainID, types.V2)
	v2Providers = append(v2Providers, staticProvider)

	if len(v2Providers) == 0 {
		return fmt.Errorf("no V2 providers available for chain %d", f.chainID)
	}

	f.v2Provider = providers.NewFallbackSubgraphProvider(
		f.chainID,
		types.V2,
		v2Providers...,
	)

	return nil
}

// initV3Provider initializes the V3 provider with fallback chain
func (f *PoolFetcher) initV3Provider() error {
	var v3Providers []providers.SubgraphProvider

	// Prefer GraphQL for completeness
	if subgraphURL, exists := config.GetSubgraphURL(f.chainID, types.V3); exists {
		graphqlProvider := providers.NewGraphQLSubgraphProvider(
			f.chainID,
			types.V3,
			subgraphURL,
			30*time.Second,
			2,
		)
		v3Providers = append(v3Providers, graphqlProvider)
	}

	// Then IPFS cache (fastest)
	if ipfsURL, exists := config.GetIPFSCacheURL(f.chainID, types.V3); exists {
		uriProvider := providers.NewURISubgraphProvider(
			f.chainID,
			types.V3,
			ipfsURL,
			30*time.Second,
			2,
		)
		cachingProvider := providers.NewCachingSubgraphProvider(
			f.chainID,
			types.V3,
			uriProvider,
			5*time.Minute,
		)
		v3Providers = append(v3Providers, cachingProvider)
	}

	// Add static provider as last resort
	staticProvider := providers.NewStaticSubgraphProvider(f.chainID, types.V3)
	v3Providers = append(v3Providers, staticProvider)

	if len(v3Providers) == 0 {
		return fmt.Errorf("no V3 providers available for chain %d", f.chainID)
	}

	f.v3Provider = providers.NewFallbackSubgraphProvider(
		f.chainID,
		types.V3,
		v3Providers...,
	)

	return nil
}

// initV4Provider initializes the V4 provider with fallback chain
func (f *PoolFetcher) initV4Provider() error {
	var v4Providers []providers.SubgraphProvider

	// Prefer GraphQL for completeness
	if subgraphURL, exists := config.GetSubgraphURL(f.chainID, types.V4); exists {
		graphqlProvider := providers.NewGraphQLSubgraphProvider(
			f.chainID,
			types.V4,
			subgraphURL,
			30*time.Second,
			2,
		)
		v4Providers = append(v4Providers, graphqlProvider)
	}

	// Then IPFS cache (fastest)
	if ipfsURL, exists := config.GetIPFSCacheURL(f.chainID, types.V4); exists {
		uriProvider := providers.NewURISubgraphProvider(
			f.chainID,
			types.V4,
			ipfsURL,
			30*time.Second,
			2,
		)
		cachingProvider := providers.NewCachingSubgraphProvider(
			f.chainID,
			types.V4,
			uriProvider,
			5*time.Minute,
		)
		v4Providers = append(v4Providers, cachingProvider)
	}

	// Add static provider as last resort
	staticProvider := providers.NewStaticSubgraphProvider(f.chainID, types.V4)
	v4Providers = append(v4Providers, staticProvider)

	if len(v4Providers) == 0 {
		return fmt.Errorf("no V4 providers available for chain %d", f.chainID)
	}

	f.v4Provider = providers.NewFallbackSubgraphProvider(
		f.chainID,
		types.V4,
		v4Providers...,
	)

	return nil
}

// GetPools fetches all pools for a specific protocol
func (f *PoolFetcher) GetPools(protocol types.Protocol, config *providers.ProviderConfig) ([]types.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var provider providers.SubgraphProvider

	switch protocol {
	case types.V2:
		provider = f.v2Provider
	case types.V3:
		provider = f.v3Provider
	case types.V4:
		provider = f.v4Provider
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}

	f.logger.Infof("Fetching %s pools for chain %d", protocol, f.chainID)

	start := time.Now()
	pools, err := provider.GetPools(ctx, config)
	duration := time.Since(start)

	if err != nil {
		f.logger.Errorf("Failed to fetch %s pools: %v", protocol, err)
		return nil, err
	}

	f.logger.Infof("Successfully fetched %d %s pools in %v", len(pools), protocol, duration)
	return pools, nil
}

// GetAllPools fetches pools from all protocols
func (f *PoolFetcher) GetAllPools(config *providers.ProviderConfig) (map[types.Protocol][]types.Pool, error) {
	result := make(map[types.Protocol][]types.Pool)

	protocols := []types.Protocol{types.V2, types.V3, types.V4}

	for _, protocol := range protocols {
		pools, err := f.GetPools(protocol, config)
		if err != nil {
			f.logger.Warnf("Failed to fetch %s pools: %v", protocol, err)
			continue
		}
		result[protocol] = pools
	}

	return result, nil
}

// GetV2Pools fetches V2 pools specifically
func (f *PoolFetcher) GetV2Pools(config *providers.ProviderConfig) ([]types.V2Pool, error) {
	if v2Provider, ok := f.v2Provider.(providers.V2SubgraphProvider); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		return v2Provider.GetV2Pools(ctx, config)
	}
	return nil, fmt.Errorf("V2 provider not available")
}

// GetV3Pools fetches V3 pools specifically
func (f *PoolFetcher) GetV3Pools(config *providers.ProviderConfig) ([]types.V3Pool, error) {
	if v3Provider, ok := f.v3Provider.(providers.V3SubgraphProvider); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		return v3Provider.GetV3Pools(ctx, config)
	}
	return nil, fmt.Errorf("V3 provider not available")
}

// GetV4Pools fetches V4 pools specifically
func (f *PoolFetcher) GetV4Pools(config *providers.ProviderConfig) ([]types.V4Pool, error) {
	if v4Provider, ok := f.v4Provider.(providers.V4SubgraphProvider); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		return v4Provider.GetV4Pools(ctx, config)
	}
	return nil, fmt.Errorf("V4 provider not available")
}

// GetChainID returns the chain ID
func (f *PoolFetcher) GetChainID() types.ChainID {
	return f.chainID
}

// SetLogLevel sets the log level
func (f *PoolFetcher) SetLogLevel(level logrus.Level) {
	f.logger.SetLevel(level)
}
