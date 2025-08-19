package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"go-fetcher/types"

	"github.com/hasura/go-graphql-client"
)

// GraphQLSubgraphProvider fetches pools directly from subgraph using GraphQL
type GraphQLSubgraphProvider struct {
	BaseProvider
	client  *graphql.Client
	timeout time.Duration
	retries int
}

// NewGraphQLSubgraphProvider creates a new GraphQL subgraph provider
func NewGraphQLSubgraphProvider(chainID types.ChainID, protocol types.Protocol, subgraphURL string, timeout time.Duration, retries int) *GraphQLSubgraphProvider {
	client := graphql.NewClient(subgraphURL, nil)
	return &GraphQLSubgraphProvider{
		BaseProvider: NewBaseProvider(chainID, protocol),
		client:       client,
		timeout:      timeout,
		retries:      retries,
	}
}

// GetPools fetches pools using GraphQL
func (p *GraphQLSubgraphProvider) GetPools(ctx context.Context, config *ProviderConfig) ([]types.Pool, error) {
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

// GetV2Pools fetches V2 pools using GraphQL
func (p *GraphQLSubgraphProvider) GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error) {
	pageSize := 1000
	threshold := "0.025"
	fetchAll := false
	if config != nil {
		if config.PageSize > 0 {
			pageSize = config.PageSize
		}
		if config.MinTrackedReserve > 0 {
			threshold = fmt.Sprintf("%f", config.MinTrackedReserve)
		}
		fetchAll = config.FetchAll
	}

	var pools []types.V2Pool
	var lastErr error

	for attempt := 0; attempt <= p.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		lastID := ""
		for {
			var q struct {
				Pairs []struct {
					ID              string `graphql:"id"`
					Token0          struct {
						ID     string `graphql:"id"`
						Symbol string `graphql:"symbol"`
					} `graphql:"token0"`
					Token1 struct {
						ID     string `graphql:"id"`
						Symbol string `graphql:"symbol"`
					} `graphql:"token1"`
					TotalSupply       string `graphql:"totalSupply"`
					TrackedReserveETH string `graphql:"trackedReserveETH"`
					ReserveETH        string `graphql:"reserveETH"`
					ReserveUSD        string `graphql:"reserveUSD"`
				} `graphql:"pairs(first: $pageSize, where: {trackedReserveETH_gt: $threshold, id_gt: $lastID}, orderBy: id, orderDirection: asc)"`
			}

			vars := map[string]interface{}{
				"pageSize":  pageSize,
				"threshold": threshold,
				"lastID":    lastID,
			}

			err := p.client.Query(ctx, &q, vars)
			if err != nil {
				lastErr = err
				break
			}

			if len(q.Pairs) == 0 {
				return pools, nil
			}

			for _, pair := range q.Pairs {
				reserve, _ := parseFloat(pair.ReserveETH)
				reserveUSD, _ := parseFloat(pair.ReserveUSD)

				pool := types.V2Pool{
					ID:          pair.ID,
					Token0:      types.Token{ID: pair.Token0.ID, Symbol: pair.Token0.Symbol},
					Token1:      types.Token{ID: pair.Token1.ID, Symbol: pair.Token1.Symbol},
					Reserve:     reserve,
					ReserveUSD:  reserveUSD,
					TotalSupply: pair.TotalSupply,
				}
				pools = append(pools, pool)
				lastID = pair.ID
			}

			if !fetchAll {
				return pools, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to fetch V2 pools after %d attempts: %w", p.retries+1, lastErr)
}

// GetV3Pools fetches V3 pools using GraphQL
func (p *GraphQLSubgraphProvider) GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error) {
	pageSize := 1000
	threshold := "0.01"
	fetchAll := false
	if config != nil {
		if config.PageSize > 0 {
			pageSize = config.PageSize
		}
		if config.MinTVLETH > 0 {
			threshold = fmt.Sprintf("%f", config.MinTVLETH)
		}
		fetchAll = config.FetchAll
	}

	var pools []types.V3Pool
	var lastErr error

	for attempt := 0; attempt <= p.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		lastID := ""
		for {
			var q struct {
				Pools []struct {
					ID string `graphql:"id"`
					Token0 struct {
						ID string `graphql:"id"`
						Symbol string `graphql:"symbol"`
						Name string `graphql:"name"`
						Decimals int `graphql:"decimals"`
					} `graphql:"token0"`
					Token1 struct {
						ID string `graphql:"id"`
						Symbol string `graphql:"symbol"`
						Name string `graphql:"name"`
						Decimals int `graphql:"decimals"`
					} `graphql:"token1"`
					FeeTier string `graphql:"feeTier"`
					Liquidity string `graphql:"liquidity"`
					TotalValueLockedUSD string `graphql:"totalValueLockedUSD"`
					TotalValueLockedETH string `graphql:"totalValueLockedETH"`
					TickSpacing string `graphql:"tickSpacing"`
				} `graphql:"pools(first: $pageSize, where: {totalValueLockedETH_gt: $threshold, id_gt: $lastID}, orderBy: id, orderDirection: asc)"`
			}

			vars := map[string]interface{}{
				"pageSize":  pageSize,
				"threshold": threshold,
				"lastID":    lastID,
			}

			err := p.client.Query(ctx, &q, vars)
			if err != nil {
				lastErr = err
				break
			}

			if len(q.Pools) == 0 {
				return pools, nil
			}

			for _, pool := range q.Pools {
				tvlUSD, _ := parseFloat(pool.TotalValueLockedUSD)
				tvlETH, _ := parseFloat(pool.TotalValueLockedETH)

				v3Pool := types.V3Pool{
					ID: pool.ID,
					Token0: types.Token{ID: pool.Token0.ID, Symbol: pool.Token0.Symbol, Name: pool.Token0.Name, Decimals: pool.Token0.Decimals},
					Token1: types.Token{ID: pool.Token1.ID, Symbol: pool.Token1.Symbol, Name: pool.Token1.Name, Decimals: pool.Token1.Decimals},
					FeeTier: pool.FeeTier,
					Liquidity: pool.Liquidity,
					TotalValueLockedUSD: tvlUSD,
					TotalValueLockedETH: tvlETH,
					TickSpacing: pool.TickSpacing,
				}
				pools = append(pools, v3Pool)
				lastID = pool.ID
			}

			if !fetchAll {
				return pools, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to fetch V3 pools after %d attempts: %w", p.retries+1, lastErr)
}

// GetV4Pools fetches V4 pools using GraphQL
func (p *GraphQLSubgraphProvider) GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error) {
	pageSize := 1000
	threshold := "0.01"
	fetchAll := false
	if config != nil {
		if config.PageSize > 0 {
			pageSize = config.PageSize
		}
		if config.MinTVLETH > 0 {
			threshold = fmt.Sprintf("%f", config.MinTVLETH)
		}
		fetchAll = config.FetchAll
	}

	var pools []types.V4Pool
	var lastErr error

	for attempt := 0; attempt <= p.retries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		lastID := ""
		for {
			var q struct {
				Pools []struct {
					ID string `graphql:"id"`
					Token0 struct {
						ID string `graphql:"id"`
						Symbol string `graphql:"symbol"`
						Name string `graphql:"name"`
						Decimals int `graphql:"decimals"`
					} `graphql:"token0"`
					Token1 struct {
						ID string `graphql:"id"`
						Symbol string `graphql:"symbol"`
						Name string `graphql:"name"`
						Decimals int `graphql:"decimals"`
					} `graphql:"token1"`
					FeeTier string `graphql:"feeTier"`
					TickSpacing string `graphql:"tickSpacing"`
					Hooks string `graphql:"hooks"`
					Liquidity string `graphql:"liquidity"`
					TotalValueLockedUSD string `graphql:"totalValueLockedUSD"`
					TotalValueLockedETH string `graphql:"totalValueLockedETH"`
				} `graphql:"pools(first: $pageSize, where: {totalValueLockedETH_gt: $threshold, id_gt: $lastID}, orderBy: id, orderDirection: asc)"`
			}

			vars := map[string]interface{}{
				"pageSize":  pageSize,
				"threshold": threshold,
				"lastID":    lastID,
			}

			err := p.client.Query(ctx, &q, vars)
			if err != nil {
				lastErr = err
				break
			}

			if len(q.Pools) == 0 {
				return pools, nil
			}

			for _, pool := range q.Pools {
				tvlUSD, _ := parseFloat(pool.TotalValueLockedUSD)
				tvlETH, _ := parseFloat(pool.TotalValueLockedETH)

				v4Pool := types.V4Pool{
					ID: pool.ID,
					Token0: types.Token{ID: pool.Token0.ID, Symbol: pool.Token0.Symbol, Name: pool.Token0.Name, Decimals: pool.Token0.Decimals},
					Token1: types.Token{ID: pool.Token1.ID, Symbol: pool.Token1.Symbol, Name: pool.Token1.Name, Decimals: pool.Token1.Decimals},
					FeeTier: pool.FeeTier,
					TickSpacing: pool.TickSpacing,
					Hooks: pool.Hooks,
					Liquidity: pool.Liquidity,
					TotalValueLockedUSD: tvlUSD,
					TotalValueLockedETH: tvlETH,
				}
				pools = append(pools, v4Pool)
				lastID = pool.ID
			}

			if !fetchAll {
				return pools, nil
			}
		}
	}

	return nil, fmt.Errorf("failed to fetch V4 pools after %d attempts: %w", p.retries+1, lastErr)
}

// parseFloat safely parses a string to float64
func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	
	var f float64
	err := json.Unmarshal([]byte(s), &f)
	return f, err
}
