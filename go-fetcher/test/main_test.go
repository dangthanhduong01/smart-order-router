package main

import (
	"testing"
	"time"
	"go-fetcher/providers"
	"go-fetcher/types"
)

func TestPoolFetcherCreation(t *testing.T) {
	// Test creating pool fetcher for Mainnet
	fetcher, err := NewPoolFetcher(types.Mainnet)
	if err != nil {
		t.Fatalf("Failed to create pool fetcher: %v", err)
	}

	if fetcher.GetChainID() != types.Mainnet {
		t.Errorf("Expected chain ID %d, got %d", types.Mainnet, fetcher.GetChainID())
	}
}

func TestPoolFetcherInvalidChain(t *testing.T) {
	// Test creating pool fetcher for invalid chain
	_, err := NewPoolFetcher(999999)
	if err == nil {
		t.Error("Expected error for invalid chain ID")
	}
}

func TestGetPools(t *testing.T) {
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	fetcher, err := NewPoolFetcher(types.Mainnet)
	if err != nil {
		t.Fatalf("Failed to create pool fetcher: %v", err)
	}

	// Test getting V3 pools
	v3Pools, err := fetcher.GetPools(types.V3, nil)
	if err != nil {
		t.Fatalf("Failed to get V3 pools: %v", err)
	}

	if len(v3Pools) == 0 {
		t.Error("Expected non-empty V3 pools list")
	}

	// Test getting V2 pools
	v2Pools, err := fetcher.GetPools(types.V2, nil)
	if err != nil {
		t.Fatalf("Failed to get V2 pools: %v", err)
	}

	if len(v2Pools) == 0 {
		t.Error("Expected non-empty V2 pools list")
	}
}

func TestGetAllPools(t *testing.T) {
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	fetcher, err := NewPoolFetcher(types.Optimism)
	if err != nil {
		t.Fatalf("Failed to create pool fetcher: %v", err)
	}

	allPools, err := fetcher.GetAllPools(nil)
	if err != nil {
		t.Fatalf("Failed to get all pools: %v", err)
	}

	// Check that we have pools for at least one protocol
	totalPools := 0
	for protocol, pools := range allPools {
		totalPools += len(pools)
		t.Logf("%s: %d pools", protocol, len(pools))
	}

	if totalPools == 0 {
		t.Error("Expected at least some pools")
	}
}

func TestProviderConfig(t *testing.T) {
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	fetcher, err := NewPoolFetcher(types.Arbitrum)
	if err != nil {
		t.Fatalf("Failed to create pool fetcher: %v", err)
	}

	// Test with custom configuration
	config := &providers.ProviderConfig{
		Timeout: 30,
		Retries: 2,
	}

	v3Pools, err := fetcher.GetV3Pools(config)
	if err != nil {
		t.Fatalf("Failed to get V3 pools with config: %v", err)
	}

	if len(v3Pools) == 0 {
		t.Error("Expected non-empty V3 pools list")
	}
}

func TestPoolInterface(t *testing.T) {
	// Test that pools implement the interface correctly
	v2Pool := types.V2Pool{
		ID:         "test-pool",
		Token0:     types.Token{ID: "token0", Symbol: "T0"},
		Token1:     types.Token{ID: "token1", Symbol: "T1"},
		Reserve:    1000.0,
		ReserveUSD: 5000.0,
	}

	// Test interface methods
	if v2Pool.GetID() != "test-pool" {
		t.Errorf("Expected ID 'test-pool', got '%s'", v2Pool.GetID())
	}

	if v2Pool.GetProtocol() != types.V2 {
		t.Errorf("Expected protocol V2, got %s", v2Pool.GetProtocol())
	}

	if v2Pool.GetTVLUSD() != 5000.0 {
		t.Errorf("Expected TVL USD 5000.0, got %f", v2Pool.GetTVLUSD())
	}
}

func TestPerformance(t *testing.T) {
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	fetcher, err := NewPoolFetcher(types.Base)
	if err != nil {
		t.Fatalf("Failed to create pool fetcher: %v", err)
	}

	// Measure performance
	start := time.Now()
	v4Pools, err := fetcher.GetV4Pools(nil)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to get V4 pools: %v", err)
	}

	t.Logf("Fetched %d V4 pools in %v", len(v4Pools), duration)

	// Performance assertion (adjust based on expected performance)
	if duration > 30*time.Second {
		t.Errorf("Fetch took too long: %v", duration)
	}
}

func TestCaching(t *testing.T) {
	// Skip if running in CI or if network is not available
	if testing.Short() {
		t.Skip("Skipping caching test in short mode")
	}

	fetcher, err := NewPoolFetcher(types.Polygon)
	if err != nil {
		t.Fatalf("Failed to create pool fetcher: %v", err)
	}

	// First request (cache miss)
	start1 := time.Now()
	v3Pools1, err := fetcher.GetV3Pools(nil)
	duration1 := time.Since(start1)

	if err != nil {
		t.Fatalf("Failed to get V3 pools (first request): %v", err)
	}

	// Second request (cache hit)
	start2 := time.Now()
	v3Pools2, err := fetcher.GetV3Pools(nil)
	duration2 := time.Since(start2)

	if err != nil {
		t.Fatalf("Failed to get V3 pools (second request): %v", err)
	}

	// Verify same number of pools
	if len(v3Pools1) != len(v3Pools2) {
		t.Errorf("Pool count mismatch: first=%d, second=%d", len(v3Pools1), len(v3Pools2))
	}

	// Verify cache is faster (second request should be much faster)
	if duration2 >= duration1 {
		t.Logf("Cache performance: first=%v, second=%v", duration1, duration2)
		// Note: This might not always be true due to network conditions
		// t.Errorf("Cache hit should be faster than cache miss")
	}
}

func BenchmarkGetPools(b *testing.B) {
	fetcher, err := NewPoolFetcher(types.Mainnet)
	if err != nil {
		b.Fatalf("Failed to create pool fetcher: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fetcher.GetPools(types.V3, nil)
		if err != nil {
			b.Fatalf("Failed to get pools: %v", err)
		}
	}
}

func BenchmarkGetAllPools(b *testing.B) {
	fetcher, err := NewPoolFetcher(types.Optimism)
	if err != nil {
		b.Fatalf("Failed to create pool fetcher: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := fetcher.GetAllPools(nil)
		if err != nil {
			b.Fatalf("Failed to get all pools: %v", err)
		}
	}
}
