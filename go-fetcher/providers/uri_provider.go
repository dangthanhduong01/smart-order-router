package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"go-fetcher/types"
	"io"
	"net/http"
	"time"
)

// URISubgraphProvider fetches pools from a URI (IPFS cache)
type URISubgraphProvider struct {
	BaseProvider
	uri     string
	timeout time.Duration
	retries int
}

// NewURISubgraphProvider creates a new URI subgraph provider
func NewURISubgraphProvider(chainID types.ChainID, protocol types.Protocol, uri string, timeout time.Duration, retries int) *URISubgraphProvider {
	return &URISubgraphProvider{
		BaseProvider: NewBaseProvider(chainID, protocol),
		uri:          uri,
		timeout:      timeout,
		retries:      retries,
	}
}

// GetPools fetches pools from the URI
func (p *URISubgraphProvider) GetPools(ctx context.Context, config *ProviderConfig) ([]types.Pool, error) {
	switch p.protocol {
	case types.V2:
		pools, err := p.GetV2Pools(ctx, config)
		if err != nil {
			return nil, err
		}
		result := make([]types.Pool, len(pools))
		for i, pool := range pools {
			result[i] = pool
		}
		return result, nil
	case types.V3:
		pools, err := p.GetV3Pools(ctx, config)
		if err != nil {
			return nil, err
		}
		result := make([]types.Pool, len(pools))
		for i, pool := range pools {
			result[i] = pool
		}
		return result, nil
	case types.V4:
		pools, err := p.GetV4Pools(ctx, config)
		if err != nil {
			return nil, err
		}
		result := make([]types.Pool, len(pools))
		for i, pool := range pools {
			result[i] = pool
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", p.protocol)
	}
}

// GetV2Pools fetches V2 pools from the URI
func (p *URISubgraphProvider) GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error) {
	data, err := p.fetchData(ctx)
	if err != nil {
		return nil, err
	}

	var pools []types.V2Pool
	if err := json.Unmarshal(data, &pools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal V2 pools: %w", err)
	}

	return pools, nil
}

// GetV3Pools fetches V3 pools from the URI
func (p *URISubgraphProvider) GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error) {
	data, err := p.fetchData(ctx)
	if err != nil {
		return nil, err
	}

	var pools []types.V3Pool
	if err := json.Unmarshal(data, &pools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal V3 pools: %w", err)
	}

	return pools, nil
}

// GetV4Pools fetches V4 pools from the URI
func (p *URISubgraphProvider) GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error) {
	data, err := p.fetchData(ctx)
	if err != nil {
		return nil, err
	}

	var pools []types.V4Pool
	if err := json.Unmarshal(data, &pools); err != nil {
		return nil, fmt.Errorf("failed to unmarshal V4 pools: %w", err)
	}

	return pools, nil
}

// fetchData fetches data from the URI with retry logic
func (p *URISubgraphProvider) fetchData(ctx context.Context) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= p.retries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		data, err := p.makeRequest(ctx)
		if err == nil {
			return data, nil
		}

		lastErr = err
	}

	return nil, fmt.Errorf("failed to fetch data after %d attempts: %w", p.retries+1, lastErr)
}

// makeRequest makes a single HTTP request to the URI
func (p *URISubgraphProvider) makeRequest(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", p.uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: p.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}
