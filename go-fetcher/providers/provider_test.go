package providers

import (
	"context"
	"testing"
	"go-fetcher/types"
	"time"
)

func TestStaticProvider(t *testing.T) {
	provider := NewStaticSubgraphProvider(types.Mainnet, types.V3)

	// Test GetPools
	pools, err := provider.GetPools(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to get pools: %v", err)
	}

	if len(pools) == 0 {
		t.Error("Expected non-empty pools list")
	}

	// Test GetV3Pools
	v3Pools, err := provider.GetV3Pools(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to get V3 pools: %v", err)
	}

	if len(v3Pools) == 0 {
		t.Error("Expected non-empty V3 pools list")
	}

	// Test that pools implement the interface correctly
	for _, pool := range pools {
		if pool.GetID() == "" {
			t.Error("Pool ID should not be empty")
		}
		if pool.GetTVLUSD() <= 0 {
			t.Error("Pool TVL should be positive")
		}
		if pool.GetProtocol() != types.V3 {
			t.Errorf("Expected protocol V3, got %s", pool.GetProtocol())
		}
	}
}

func TestFallbackProvider(t *testing.T) {
	staticProvider := NewStaticSubgraphProvider(types.Mainnet, types.V3)
	fallbackProvider := NewFallbackSubgraphProvider(types.Mainnet, types.V3, staticProvider)

	pools, err := fallbackProvider.GetPools(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to get pools from fallback provider: %v", err)
	}

	if len(pools) == 0 {
		t.Error("Expected non-empty pools list from fallback provider")
	}
}

func TestCachingProvider(t *testing.T) {
	staticProvider := NewStaticSubgraphProvider(types.Mainnet, types.V3)
	cachingProvider := NewCachingSubgraphProvider(types.Mainnet, types.V3, staticProvider, 5*time.Minute)

	// First request
	pools1, err := cachingProvider.GetPools(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to get pools (first request): %v", err)
	}

	// Second request (should be cached)
	pools2, err := cachingProvider.GetPools(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to get pools (second request): %v", err)
	}

	if len(pools1) != len(pools2) {
		t.Errorf("Pool count mismatch: first=%d, second=%d", len(pools1), len(pools2))
	}

	// Test cache stats
	stats := cachingProvider.GetCacheStats()
	if stats["itemCount"] == nil {
		t.Error("Cache stats should contain itemCount")
	}
}

func TestProviderConfig(t *testing.T) {
	config := &ProviderConfig{
		Timeout: 30,
		Retries: 2,
	}

	if config.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", config.Timeout)
	}

	if config.Retries != 2 {
		t.Errorf("Expected retries 2, got %d", config.Retries)
	}
}
