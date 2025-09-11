package providers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go-fetcher/types"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/hasura/go-graphql-client"
)

// GraphQLSubgraphProvider fetches pools directly from subgraph using GraphQL
type GraphQLSubgraphProvider struct {
	BaseProvider
	client        *graphql.Client
	timeout       time.Duration
	retries       int
	queryBuilder  *QueryBuilder
	poolBlacklist *PoolBlacklist // Add blacklist for pool tracking
	// concurrency control
	maxConcurrency int
	sem            chan struct{}
}

// NewGraphQLSubgraphProvider creates a new GraphQL subgraph provider
func NewGraphQLSubgraphProvider(chainID types.ChainID, protocol types.Protocol, subgraphURL string, timeout time.Duration, retries int) *GraphQLSubgraphProvider {
	client := graphql.NewClient(subgraphURL, nil)
	p := &GraphQLSubgraphProvider{
		BaseProvider:   NewBaseProvider(chainID, protocol),
		client:         client,
		timeout:        timeout,
		retries:        retries,
		queryBuilder:   NewQueryBuilder(protocol, chainID),
		poolBlacklist:  NewPoolBlacklist(3), // blacklist after 3 failures
		maxConcurrency: getMaxConcurrencyFromEnv(),
	}
	p.sem = make(chan struct{}, p.maxConcurrency)
	return p
}

// NewGraphQLSubgraphProviderWithAuth creates a new GraphQL subgraph provider with API key authentication
func NewGraphQLSubgraphProviderWithAuth(chainID types.ChainID, protocol types.Protocol, subgraphURL string, apiKey string, timeout time.Duration, retries int) *GraphQLSubgraphProvider {
	// Create HTTP client with Authorization header
	httpClient := &http.Client{
		Timeout: timeout,
	}

	// Create GraphQL client with auth headers
	client := graphql.NewClient(subgraphURL, httpClient).WithRequestModifier(func(req *http.Request) {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
		req.Header.Set("Content-Type", "application/json")
	})

	p := &GraphQLSubgraphProvider{
		BaseProvider:   NewBaseProvider(chainID, protocol),
		client:         client,
		timeout:        timeout,
		retries:        retries,
		queryBuilder:   NewQueryBuilder(protocol, chainID),
		poolBlacklist:  NewPoolBlacklist(3), // blacklist after 3 failures
		maxConcurrency: getMaxConcurrencyFromEnv(),
	}
	p.sem = make(chan struct{}, p.maxConcurrency)

	return p
}

// acquire a slot in the semaphore
func (p *GraphQLSubgraphProvider) acquire() {
	if p == nil || p.sem == nil {
		return
	}
	p.sem <- struct{}{}
}

// release a slot in the semaphore
func (p *GraphQLSubgraphProvider) release() {
	if p == nil || p.sem == nil {
		return
	}
	select {
	case <-p.sem:
	default:
	}
}

// getMaxConcurrencyFromEnv reads GRAPHQL_MAX_CONCURRENCY env or returns default 4
func getMaxConcurrencyFromEnv() int {
	v := os.Getenv("GRAPHQL_MAX_CONCURRENCY")
	if v == "" {
		return 4
	}
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return n
	}
	return 4
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

// QueryConfig represents a query configuration
type QueryConfig struct {
	Name  string
	Where string
}

// fetchPoolsForQuery fetches pools for a specific query
func (p *GraphQLSubgraphProvider) fetchPoolsForQuery(ctx context.Context, queryConfig QueryConfig, pageSize int, fetchAll bool) ([]types.V2Pool, error) {
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
					ID     string `graphql:"id"`
					Token0 struct {
						ID       string `graphql:"id"`
						Symbol   string `graphql:"symbol"`
						Name     string `graphql:"name"`
						Decimals string `graphql:"decimals"`
					} `graphql:"token0"`
					Token1 struct {
						ID       string `graphql:"id"`
						Symbol   string `graphql:"symbol"`
						Name     string `graphql:"name"`
						Decimals string `graphql:"decimals"`
					} `graphql:"token1"`
					TotalSupply       string `graphql:"totalSupply"`
					TrackedReserveETH string `graphql:"trackedReserveETH"`
					ReserveETH        string `graphql:"reserveETH"`
					ReserveUSD        string `graphql:"reserveUSD"`
				} `graphql:"pairs(first: $pageSize, where: {id_gt: $lastID}, orderBy: id, orderDirection: asc)"`
			}

			// Note: In real implementation, you'd need dynamic GraphQL query building
			vars := map[string]interface{}{
				"pageSize": pageSize,
				"lastID":   lastID,
			}

			p.acquire()
			err := p.client.Query(ctx, &q, vars)
			p.release()

			if err != nil {
				lastErr = err
				break
			}

			if len(q.Pairs) == 0 {
				return pools, nil
			}

			// collect tokens for batched metadata fetch
			pageTokens := []*types.Token{}

			for _, pair := range q.Pairs {
				reserve, _ := parseFloat(pair.TrackedReserveETH)
				reserveUSD, _ := parseFloat(pair.ReserveUSD)

				// parse decimals safely
				dec0 := 0
				if pair.Token0.Decimals != "" {
					if d, err := strconv.Atoi(pair.Token0.Decimals); err == nil {
						dec0 = d
					}
				}
				dec1 := 0
				if pair.Token1.Decimals != "" {
					if d, err := strconv.Atoi(pair.Token1.Decimals); err == nil {
						dec1 = d
					}
				}

				pPool := types.V2Pool{
					ID:          strings.ToLower(pair.ID),
					Token0:      types.Token{ID: strings.ToLower(pair.Token0.ID), Symbol: pair.Token0.Symbol, Name: pair.Token0.Name, Decimals: dec0},
					Token1:      types.Token{ID: strings.ToLower(pair.Token1.ID), Symbol: pair.Token1.Symbol, Name: pair.Token1.Name, Decimals: dec1},
					Reserve:     reserve,
					ReserveUSD:  reserveUSD,
					TotalSupply: pair.TotalSupply,
				}

				pools = append(pools, pPool)
				// pointers to slice elements for batch filling
				idx := len(pools) - 1
				pageTokens = append(pageTokens, &pools[idx].Token0, &pools[idx].Token1)

				lastID = pair.ID
			}

			// batch fill token metadata (best-effort)
			if len(pageTokens) > 0 {
				if err := p.EnsureTokensMetadataBatch(ctx, pageTokens); err != nil {
					fmt.Printf("warning: batch token metadata fetch failed: %v\n", err)
				}
			}

			if !fetchAll {
				break
			}
		}

		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to fetch pools for query %s after %d attempts: %w", queryConfig.Name, p.retries+1, lastErr)
	}

	return pools, nil
}

// shouldIncludeV2Pool determines if a V2 pool should be included based on smart-order-router logic
func (p *GraphQLSubgraphProvider) shouldIncludeV2Pool(pool types.V2Pool, feiToken, virtualToken string, threshold, untrackedUsdThreshold float64) bool {
	return p.queryBuilder.ShouldIncludeV2Pool(pool, threshold)
}

// GetV2Pools fetches V2 pools using GraphQL with multiple queries like smart-order-router
func (p *GraphQLSubgraphProvider) GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error) {
	pageSize := 1000
	threshold := 0.025
	untrackedUsdThreshold := float64(^uint(0) >> 1) // Max value
	fetchAll := false

	if config != nil {
		if config.PageSize > 0 {
			pageSize = config.PageSize
		}
		if config.MinTrackedReserve > 0 {
			threshold = config.MinTrackedReserve
		}
		fetchAll = config.FetchAll
	}

	// FEI token address (same as smart-order-router)
	feiToken := "0x956f47f50a910163d8bf957cf5846d573e7f87ca"
	virtualToken := "0x0b3e328455c4059eeb9e3f84b5543f74e24e7e1b"

	// Define queries similar to smart-order-router
	queries := []QueryConfig{
		// 1. FEI token pools (token0)
		{
			Name:  "FEI pools (token0)",
			Where: fmt.Sprintf("token0: \"%s\"", feiToken),
		},
		// 2. FEI token pools (token1)
		{
			Name:  "FEI pools (token1)",
			Where: fmt.Sprintf("token1: \"%s\"", feiToken),
		},
		// 3. High tracked reserve ETH pools
		{
			Name:  "High tracked reserve ETH pools",
			Where: fmt.Sprintf("trackedReserveETH_gt: \"%f\"", threshold),
		},
		// 4. High untracked USD pools
		{
			Name:  "High untracked USD pools",
			Where: fmt.Sprintf("reserveUSD_gt: \"%f\"", untrackedUsdThreshold),
		},
	}

	// Add virtual pair pools for BASE chain
	if p.chainID == types.Base {
		queries = append(queries, []QueryConfig{
			{
				Name:  "Virtual pair pools (token0)",
				Where: fmt.Sprintf("token0: \"%s\"", virtualToken),
			},
			{
				Name:  "Virtual pair pools (token1)",
				Where: fmt.Sprintf("token1: \"%s\"", virtualToken),
			},
		}...)
	}

	var (
		allPools []types.V2Pool
		poolMap  = make(map[string]types.V2Pool) // For deduplication
		mu       sync.Mutex
		wg       sync.WaitGroup
		resCh    = make(chan struct {
			name  string
			pools []types.V2Pool
			err   error
		}, len(queries))
	)

	// fetch queries in parallel
	for _, qc := range queries {
		wg.Add(1)
		qc := qc
		go func() {
			defer wg.Done()
			pools, err := p.fetchPoolsForQuery(ctx, qc, pageSize, fetchAll)
			resCh <- struct {
				name  string
				pools []types.V2Pool
				err   error
			}{qc.Name, pools, err}
		}()
	}

	// wait and close
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// process results
	for r := range resCh {
		if r.err != nil {
			return nil, fmt.Errorf("failed to fetch pools for query %s: %w", r.name, r.err)
		}
		// Save per-query results to file (streamed)
		sanitized := sanitizeFilename(r.name)
		qFilename := fmt.Sprintf("v2_query_%s.json", sanitized)
		if err := p.savePoolsToFile(qFilename, r.pools); err != nil {
			fmt.Printf("warning: failed to save V2 query '%s' to file: %v\n", r.name, err)
		}
		// aggregate
		mu.Lock()
		for _, pool := range r.pools {
			poolMap[pool.ID] = pool
		}
		mu.Unlock()
	}

	// Convert map to slice
	mu.Lock()
	for _, pool := range poolMap {
		allPools = append(allPools, pool)
	}
	mu.Unlock()

	// Apply filtering logic similar to smart-order-router
	var filteredPools []types.V2Pool
	for _, pool := range allPools {
		if p.shouldIncludeV2Pool(pool, feiToken, virtualToken, threshold, untrackedUsdThreshold) {
			filteredPools = append(filteredPools, pool)
		}
	}

	// Save aggregate to chain folder
	aggFilename := fmt.Sprintf("v2_chain_%d.json", p.chainID)
	if err := p.savePoolsToFile(aggFilename, filteredPools); err != nil {
		fmt.Printf("warning: failed to save aggregated V2 pools to file: %v\n", err)
	}

	return filteredPools, nil
}

// V3QueryConfig represents a V3 query configuration
type V3QueryConfig struct {
	Name  string
	Where string
}

// fetchV3PoolsForQuery fetches V3 pools for a specific query
func (p *GraphQLSubgraphProvider) fetchV3PoolsForQuery(ctx context.Context, queryConfig V3QueryConfig, pageSize int, fetchAll bool) ([]types.V3Pool, error) {
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
					ID     string `graphql:"id"`
					Token0 struct {
						ID       string `graphql:"id"`
						Symbol   string `graphql:"symbol"`
						Name     string `graphql:"name"`
						Decimals string `graphql:"decimals"`
					} `graphql:"token0"`
					Token1 struct {
						ID       string `graphql:"id"`
						Symbol   string `graphql:"symbol"`
						Name     string `graphql:"name"`
						Decimals string `graphql:"decimals"`
					} `graphql:"token1"`
					FeeTier             string `graphql:"feeTier"`
					Liquidity           string `graphql:"liquidity"`
					TotalValueLockedUSD string `graphql:"totalValueLockedUSD"`
					TotalValueLockedETH string `graphql:"totalValueLockedETH"`
					// removed tickSpacing: some subgraphs don't expose this field
				} `graphql:"pools(first: $pageSize, where: {id_gt: $lastID}, orderBy: id, orderDirection: asc)"`
			}

			// Note: In real implementation, you'd need dynamic GraphQL query building
			vars := map[string]interface{}{
				"pageSize": pageSize,
				"lastID":   lastID,
			}

			p.acquire()
			err := p.client.Query(ctx, &q, vars)
			p.release()

			if err != nil {
				lastErr = err
				break
			}

			if len(q.Pools) == 0 {
				return pools, nil
			}

			// collect tokens for batched metadata fetch
			pageTokens := []*types.Token{}

			for _, pool := range q.Pools {
				tvlUSD, _ := parseFloat(pool.TotalValueLockedUSD)
				tvlETH, _ := parseFloat(pool.TotalValueLockedETH)

				// parse decimals safely
				dec0 := 0
				if pool.Token0.Decimals != "" {
					if d, err := strconv.Atoi(pool.Token0.Decimals); err == nil {
						dec0 = d
					}
				}
				dec1 := 0
				if pool.Token1.Decimals != "" {
					if d, err := strconv.Atoi(pool.Token1.Decimals); err == nil {
						dec1 = d
					}
				}

				v3Pool := types.V3Pool{
					ID: strings.ToLower(pool.ID),
					Token0: types.Token{
						ID:       strings.ToLower(pool.Token0.ID),
						Symbol:   pool.Token0.Symbol,
						Name:     pool.Token0.Name,
						Decimals: dec0,
					},
					Token1: types.Token{
						ID:       strings.ToLower(pool.Token1.ID),
						Symbol:   pool.Token1.Symbol,
						Name:     pool.Token1.Name,
						Decimals: dec1,
					},
					FeeTier:             pool.FeeTier,
					Liquidity:           pool.Liquidity,
					TotalValueLockedUSD: tvlUSD,
					TotalValueLockedETH: tvlETH,
					TickSpacing:         "", // not requested from subgraph
				}

				pools = append(pools, v3Pool)
				// pointers to slice elements for batch filling
				idx := len(pools) - 1
				pageTokens = append(pageTokens, &pools[idx].Token0, &pools[idx].Token1)

				lastID = pool.ID
			}

			// batch fill token metadata (best-effort)
			if len(pageTokens) > 0 {
				if err := p.EnsureTokensMetadataBatch(ctx, pageTokens); err != nil {
					fmt.Printf("warning: batch token metadata fetch failed: %v\n", err)
				}
			}

			if !fetchAll {
				break
			}
		}

		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to fetch V3 pools for query %s after %d attempts: %w", queryConfig.Name, p.retries+1, lastErr)
	}

	return pools, nil
}

// shouldIncludeV3Pool determines if a V3 pool should be included based on smart-order-router V3 logic
func (p *GraphQLSubgraphProvider) shouldIncludeV3Pool(pool types.V3Pool, threshold float64) bool {
	return p.queryBuilder.ShouldIncludeV3Pool(pool, threshold)
}

// GetV3Pools fetches V3 pools using GraphQL with multiple queries like smart-order-router
func (p *GraphQLSubgraphProvider) GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error) {
	pageSize := 1000
	threshold := 0.01
	fetchAll := false

	if config != nil {
		if config.PageSize > 0 {
			pageSize = config.PageSize
		}
		if config.MinTVLETH > 0 {
			threshold = config.MinTVLETH
		}
		fetchAll = config.FetchAll
	}

	// Define queries similar to smart-order-router V3 logic
	queries := []V3QueryConfig{
		// 1. High TVL ETH pools
		{
			Name:  "High TVL ETH pools",
			Where: fmt.Sprintf("totalValueLockedETH_gt: \"%f\"", threshold),
		},
		// 2. Zero ETH pools (special case)
		{
			Name:  "Zero TVL ETH pools",
			Where: "totalValueLockedETH: \"0\"",
		},
	}

	var (
		allPools []types.V3Pool
		poolMap  = make(map[string]types.V3Pool) // For deduplication
		mu       sync.Mutex
		wg       sync.WaitGroup
		resCh    = make(chan struct {
			name  string
			pools []types.V3Pool
			err   error
		}, len(queries))
	)

	// fetch queries in parallel
	for _, qc := range queries {
		wg.Add(1)
		qc := qc
		go func() {
			defer wg.Done()
			pools, err := p.fetchV3PoolsForQuery(ctx, qc, pageSize, fetchAll)
			resCh <- struct {
				name  string
				pools []types.V3Pool
				err   error
			}{qc.Name, pools, err}
		}()
	}

	// wait and close
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// process results
	for r := range resCh {
		if r.err != nil {
			return nil, fmt.Errorf("failed to fetch V3 pools for query %s: %w", r.name, r.err)
		}
		// Save per-query results to file (streamed)
		sanitized := sanitizeFilename(r.name)
		qFilename := fmt.Sprintf("v3_query_%s.json", sanitized)
		if err := p.savePoolsToFile(qFilename, r.pools); err != nil {
			fmt.Printf("warning: failed to save V3 query '%s' to file: %v\n", r.name, err)
		}
		// aggregate
		mu.Lock()
		for _, pool := range r.pools {
			poolMap[pool.ID] = pool
		}
		mu.Unlock()
	}

	// Convert map to slice
	mu.Lock()
	for _, pool := range poolMap {
		allPools = append(allPools, pool)
	}
	mu.Unlock()

	// Apply filtering logic similar to smart-order-router
	var filteredPools []types.V3Pool
	for _, pool := range allPools {
		if p.shouldIncludeV3Pool(pool, threshold) {
			filteredPools = append(filteredPools, pool)
		}
	}

	// Save aggregate to chain folder
	aggFilename := fmt.Sprintf("v3_chain_%d.json", p.chainID)
	if err := p.savePoolsToFile(aggFilename, filteredPools); err != nil {
		fmt.Printf("warning: failed to save aggregated V3 pools to file: %v\n", err)
	}

	return filteredPools, nil
}

// V4QueryConfig represents a V4 query configuration
type V4QueryConfig struct {
	Name  string
	Where string
}

// fetchV4PoolsForQuery fetches V4 pools for a specific query
func (p *GraphQLSubgraphProvider) fetchV4PoolsForQuery(ctx context.Context, queryConfig V4QueryConfig, pageSize int, fetchAll bool) ([]types.V4Pool, error) {
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
					ID     string `graphql:"id"`
					Token0 struct {
						ID       string `graphql:"id"`
						Symbol   string `graphql:"symbol"`
						Name     string `graphql:"name"`
						Decimals string `graphql:"decimals"`
					} `graphql:"token0"`
					Token1 struct {
						ID       string `graphql:"id"`
						Symbol   string `graphql:"symbol"`
						Name     string `graphql:"name"`
						Decimals string `graphql:"decimals"`
					} `graphql:"token1"`
					FeeTier     string `graphql:"feeTier"`
					TickSpacing string `graphql:"tickSpacing"` // removed if absent
					// Hooks may not exist on all gateways; remove from query
					Liquidity           string `graphql:"liquidity"`
					TotalValueLockedUSD string `graphql:"totalValueLockedUSD"`
					TotalValueLockedETH string `graphql:"totalValueLockedETH"`
				} `graphql:"pools(first: $pageSize, where: {id_gt: $lastID}, orderBy: id, orderDirection: asc)"`
			}

			vars := map[string]interface{}{
				"pageSize": pageSize,
				"lastID":   lastID,
			}

			p.acquire()
			err := p.client.Query(ctx, &q, vars)
			p.release()

			if err != nil {
				lastErr = err
				break
			}

			if len(q.Pools) == 0 {
				return pools, nil
			}

			// collect tokens for batched metadata fetch
			pageTokens := []*types.Token{}

			for _, pool := range q.Pools {
				tvlUSD, _ := parseFloat(pool.TotalValueLockedUSD)
				tvlETH, _ := parseFloat(pool.TotalValueLockedETH)

				// parse decimals safely
				dec0 := 0
				if pool.Token0.Decimals != "" {
					if d, err := strconv.Atoi(pool.Token0.Decimals); err == nil {
						dec0 = d
					}
				}
				dec1 := 0
				if pool.Token1.Decimals != "" {
					if d, err := strconv.Atoi(pool.Token1.Decimals); err == nil {
						dec1 = d
					}
				}

				v4Pool := types.V4Pool{
					ID: strings.ToLower(pool.ID),
					Token0: types.Token{
						ID:       strings.ToLower(pool.Token0.ID),
						Symbol:   pool.Token0.Symbol,
						Name:     pool.Token0.Name,
						Decimals: dec0,
					},
					Token1: types.Token{
						ID:       strings.ToLower(pool.Token1.ID),
						Symbol:   pool.Token1.Symbol,
						Name:     pool.Token1.Name,
						Decimals: dec1,
					},
					FeeTier:             pool.FeeTier,
					TickSpacing:         "", // not requested to avoid schema mismatch
					Hooks:               "", // not requested to avoid schema mismatch
					Liquidity:           pool.Liquidity,
					TotalValueLockedUSD: tvlUSD,
					TotalValueLockedETH: tvlETH,
				}

				pools = append(pools, v4Pool)
				// pointers to slice elements for batch filling
				idx := len(pools) - 1
				pageTokens = append(pageTokens, &pools[idx].Token0, &pools[idx].Token1)

				lastID = pool.ID
			}

			// batch fill token metadata (best-effort)
			if len(pageTokens) > 0 {
				if err := p.EnsureTokensMetadataBatch(ctx, pageTokens); err != nil {
					fmt.Printf("warning: batch token metadata fetch failed: %v\n", err)
				}
			}

			if !fetchAll {
				break
			}
		}

		if lastErr == nil {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("failed to fetch V4 pools for query %s after %d attempts: %w", queryConfig.Name, p.retries+1, lastErr)
	}

	return pools, nil
}

// shouldIncludeV4Pool determines if a V4 pool should be included based on smart-order-router V4 logic
func (p *GraphQLSubgraphProvider) shouldIncludeV4Pool(pool types.V4Pool, threshold float64) bool {
	return p.queryBuilder.ShouldIncludeV4Pool(pool, threshold)
}

// GetV4Pools fetches V4 pools using GraphQL with query like smart-order-router
func (p *GraphQLSubgraphProvider) GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error) {
	pageSize := 1000
	threshold := 0.01
	fetchAll := false

	// Use larger page size for V4 on BASE chain like smart-order-router
	if p.chainID == types.Base {
		pageSize = 3500
	}

	if config != nil {
		if config.PageSize > 0 {
			pageSize = config.PageSize
		}
		if config.MinTVLETH > 0 {
			threshold = config.MinTVLETH
		}
		fetchAll = config.FetchAll
	}

	// Define queries similar to smart-order-router V4 logic
	queries := []V4QueryConfig{
		// 1. High tracked ETH pools
		{
			Name:  "High tracked ETH pools",
			Where: fmt.Sprintf("totalValueLockedETH_gt: \"%f\"", threshold),
		},
		// 2. V4 high liquidity pools (special V4 condition)
		{
			Name:  "V4 high liquidity pools",
			Where: "liquidity_gt: \"0\"",
		},
	}

	var (
		allPools []types.V4Pool
		poolMap  = make(map[string]types.V4Pool) // For deduplication
		mu       sync.Mutex
		wg       sync.WaitGroup
		resCh    = make(chan struct {
			name  string
			pools []types.V4Pool
			err   error
		}, len(queries))
	)

	// fetch queries in parallel
	for _, qc := range queries {
		wg.Add(1)
		qc := qc
		go func() {
			defer wg.Done()
			pools, err := p.fetchV4PoolsForQuery(ctx, qc, pageSize, fetchAll)
			resCh <- struct {
				name  string
				pools []types.V4Pool
				err   error
			}{qc.Name, pools, err}
		}()
	}

	// wait and close
	go func() {
		wg.Wait()
		close(resCh)
	}()

	// process results
	for r := range resCh {
		if r.err != nil {
			return nil, fmt.Errorf("failed to fetch V4 pools for query %s: %w", r.name, r.err)
		}
		sanitized := sanitizeFilename(r.name)
		qFilename := fmt.Sprintf("v4_query_%s.json", sanitized)
		if err := p.savePoolsToFile(qFilename, r.pools); err != nil {
			fmt.Printf("warning: failed to save V4 query '%s' to file: %v\n", r.name, err)
		}
		mu.Lock()
		for _, pool := range r.pools {
			poolMap[pool.ID] = pool
		}
		mu.Unlock()
	}

	for _, pool := range poolMap {
		allPools = append(allPools, pool)
	}

	var filteredPools []types.V4Pool
	for _, pool := range allPools {
		if p.shouldIncludeV4Pool(pool, threshold) {
			filteredPools = append(filteredPools, pool)
		}
	}

	aggFilename := fmt.Sprintf("v4_chain_%d.json", p.chainID)
	if err := p.savePoolsToFile(aggFilename, filteredPools); err != nil {
		fmt.Printf("warning: failed to save aggregated V4 pools to file: %v\n", err)
	}

	return filteredPools, nil
}

// tokenListCache caches token metadata loaded from a local tokenlist JSON
var (
	tokenListOnce sync.Once
	tokenListMap  map[string]tokenMeta
)

type tokenMeta struct {
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

// loadTokenList attempts to load a token list JSON from env TOKEN_LIST_FILE or ./tokenlist.json
func loadTokenList() {
	tokenListOnce.Do(func() {
		tokenListMap = make(map[string]tokenMeta)
		path := os.Getenv("TOKEN_LIST_FILE")
		if path == "" {
			path = "./tokenlist.json"
		}
		b, err := os.ReadFile(path)
		if err != nil {
			// no tokenlist available
			return
		}
		// Expect file to be either an object map or a standard token list (Uniswap tokenlist) with "tokens" array
		var asMap map[string]tokenMeta
		if err := json.Unmarshal(b, &asMap); err == nil && len(asMap) > 0 {
			for k, v := range asMap {
				tokenListMap[strings.ToLower(k)] = v
			}
			return
		}

		// Try Uniswap tokenlist format
		var uni struct {
			Tokens []struct {
				Address  string `json:"address"`
				Name     string `json:"name"`
				Symbol   string `json:"symbol"`
				Decimals int    `json:"decimals"`
			} `json:"tokens"`
		}
		if err := json.Unmarshal(b, &uni); err == nil && len(uni.Tokens) > 0 {
			for _, t := range uni.Tokens {
				tokenListMap[strings.ToLower(t.Address)] = tokenMeta{Name: t.Name, Symbol: t.Symbol, Decimals: t.Decimals}
			}
		}
	})
}

// fetchTokenOnChain attempts to read ERC20 name() and decimals() from chain using ETH_RPC_URL via raw JSON-RPC eth_call
func fetchTokenOnChain(ctx context.Context, tokenAddr string) (tokenMeta, error) {
	rpc := os.Getenv("ETH_RPC_URL")
	if rpc == "" {
		return tokenMeta{}, fmt.Errorf("no ETH_RPC_URL provided")
	}

	// helper to perform eth_call with given data
	call := func(data string) (string, error) {
		payload := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "eth_call",
			"params": []interface{}{
				map[string]string{"to": strings.ToLower(tokenAddr), "data": data},
				"latest",
			},
		}
		b, _ := json.Marshal(payload)
		req, err := http.NewRequestWithContext(ctx, "POST", rpc, strings.NewReader(string(b)))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		rb, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		var r struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(rb, &r); err != nil {
			return "", err
		}
		return r.Result, nil
	}

	// selectors
	nameSel := "0x06fdde03"
	decimalsSel := "0x313ce567"

	var meta tokenMeta

	// 1) try name()
	res, err := call(nameSel)
	if err == nil && res != "" && res != "0x" {
		// res is hex-encoded ABI return; try decode as dynamic string then as bytes32
		s := decodeABIString(res)
		if s != "" {
			meta.Name = s
		} else {
			// try bytes32
			if bs, err := hex.DecodeString(strings.TrimPrefix(res, "0x")); err == nil && len(bs) >= 32 {
				// take first 32 bytes and trim zeros
				n := bytesTrimZero(bs[:32])
				meta.Name = string(n)
			}
		}
	}

	// 2) decimals()
	res2, err2 := call(decimalsSel)
	if err2 == nil && res2 != "" && res2 != "0x" {
		if v, err := decodeUint256(res2); err == nil {
			meta.Decimals = v
		}
	}

	if meta.Name == "" && meta.Decimals == 0 {
		return tokenMeta{}, fmt.Errorf("on-chain metadata not available")
	}
	return meta, nil
}

// decodeABIString attempts to decode a dynamic ABI-encoded string return (0x + offset + length + data)
func decodeABIString(hexRes string) string {
	s := strings.TrimPrefix(hexRes, "0x")
	if len(s) < 64*3 { // at least offset+length+one word
		return ""
	}
	// length is at position 64..128
	lengthHex := s[64:128]
	length, err := strconv.ParseInt(lengthHex, 16, 64)
	if err != nil || length <= 0 {
		return ""
	}
	dataStart := 128
	dataEnd := dataStart + int(length)*2
	if dataEnd > len(s) {
		dataEnd = len(s)
	}
	dataHex := s[dataStart:dataEnd]
	if b, err := hex.DecodeString(dataHex); err == nil {
		return string(b)
	}
	return ""
}

// decodeUint256 parses a hex-encoded 32-byte uint256 return into int
func decodeUint256(hexRes string) (int, error) {
	s := strings.TrimPrefix(hexRes, "0x")
	if s == "" {
		return 0, fmt.Errorf("empty result")
	}
	// take last 64 chars
	if len(s) > 64 {
		s = s[len(s)-64:]
	}
	v, err := strconv.ParseInt(strings.TrimLeft(s, "0"), 16, 64)
	if err != nil {
		// fallback: parse full hex
		u, err2 := strconv.ParseInt(s, 16, 64)
		if err2 != nil {
			return 0, err2
		}
		return int(u), nil
	}
	return int(v), nil
}

// bytesTrimZero trims trailing zero bytes
func bytesTrimZero(b []byte) []byte {
	end := len(b)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0 {
			end = i + 1
			break
		}
	}
	return b[:end]
}

// multicall ABI (only tryAggregate needed)
const multicallABI = `[{"inputs":[{"internalType":"bool","name":"requireSuccess","type":"bool"},{"components":[{"internalType":"address","name":"target","type":"address"},{"internalType":"bytes","name":"callData","type":"bytes"}],"internalType":"struct Multicall2.Call[]","name":"calls","type":"tuple[]"}],"name":"tryAggregate","outputs":[{"components":[{"internalType":"bool","name":"success","type":"bool"},{"internalType":"bytes","name":"returnData","type":"bytes"}],"name":"returnData","type":"tuple[]"}],"stateMutability":"nonpayable","type":"function"}]`

// getMulticallAddress returns a multicall contract address for common chains (lowercased).
// It first checks an environment variable MULTICALL_ADDRESS_<CHAINID> to allow overrides.
func getMulticallAddress(chainID types.ChainID) string {
	// allow per-chain override via env e.g. MULTICALL_ADDRESS_56
	if v := os.Getenv(fmt.Sprintf("MULTICALL_ADDRESS_%d", int(chainID))); v != "" {
		return strings.ToLower(v)
	}

	switch int(chainID) {
	case 1: // Ethereum Mainnet (Multicall2)
		return strings.ToLower("0x5BA1e12693Dc8F9c48aAD8770482f4739bEeD696")
	case 56: // BSC common multicall address (use this default)
		return strings.ToLower("0x1Ee38d535d541c55C9dae27B12edf090C608E6Fb")
	case 42161: // Arbitrum (user provided default earlier)
		if v := os.Getenv("MULTICALL_ADDRESS_ARBITRUM"); v != "" {
			return strings.ToLower(v)
		}
		return strings.ToLower("0xBF69a56D35B8d6f5A8e0e96B245a72F735751e54")
	default:
		return ""
	}
}

// getRPCForChain returns an RPC URL for a chain. It prefers explicit env vars RPC_MAINNET, RPC_BSC, RPC_ARBITRUM
// otherwise falls back to ETH_RPC_URL.
func getRPCForChain(chainID types.ChainID) string {
	switch int(chainID) {
	case 1:
		if v := os.Getenv("RPC_MAINNET"); v != "" {
			return v
		}
	case 56:
		if v := os.Getenv("RPC_BSC"); v != "" {
			return v
		}
	case 42161:
		if v := os.Getenv("RPC_ARBITRUM"); v != "" {
			return v
		}
	}
	// fallback
	return os.Getenv("ETH_RPC_URL")
}

// getRPCRetryParams reads RPC_RETRIES and RPC_BACKOFF_MS env vars (defaults: retries=3, baseBackoff=200ms)
func getRPCRetryParams() (int, int) {
	retries := 3
	if v := os.Getenv("RPC_RETRIES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			retries = n
		}
	}
	baseMs := 200
	if v := os.Getenv("RPC_BACKOFF_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			baseMs = n
		}
	}
	return retries, baseMs
}

// fetchTokensOnChainMulticall uses an on-chain Multicall contract (tryAggregate) to fetch name() and decimals()
// for many tokens in batches. It returns a map[addressLower]tokenMeta.
func (p *GraphQLSubgraphProvider) fetchTokensOnChainMulticall(ctx context.Context, tokenAddrs []string) (map[string]tokenMeta, error) {
	if len(tokenAddrs) == 0 {
		return nil, nil
	}

	mcAddr := getMulticallAddress(p.chainID)
	if mcAddr == "" {
		return nil, fmt.Errorf("no multicall address for chain %d", p.chainID)
	}

	rpc := getRPCForChain(p.chainID)
	if rpc == "" {
		return nil, fmt.Errorf("no RPC URL for chain %d", p.chainID)
	}

	// decode selectors
	nameSel, _ := hex.DecodeString("06fdde03")
	decSel, _ := hex.DecodeString("313ce567")

	// parsed ABI
	parsed, err := abi.JSON(strings.NewReader(multicallABI))
	if err != nil {
		return nil, err
	}

	// prepare result map
	outMap := make(map[string]tokenMeta)

	// chunk tokens to avoid too-large payloads; default 100 tokens per multicall
	batchSize := 100
	if v := os.Getenv("MULTICALL_BATCH_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			batchSize = n
		}
	}

	for i := 0; i < len(tokenAddrs); i += batchSize {
		end := i + batchSize
		if end > len(tokenAddrs) {
			end = len(tokenAddrs)
		}
		batch := tokenAddrs[i:end]

		// build calls: for each token add name() and decimals()
		type mcCall struct {
			Target   common.Address
			CallData []byte
		}
		calls := make([]mcCall, 0, len(batch)*2)
		addrOrder := make([]string, 0, len(batch)*2) // matches calls index -> addrLower

		for _, a := range batch {
			addrLower := strings.ToLower(a)
			t := common.HexToAddress(addrLower)
			calls = append(calls, mcCall{Target: t, CallData: nameSel})
			addrOrder = append(addrOrder, addrLower)
			calls = append(calls, mcCall{Target: t, CallData: decSel})
			addrOrder = append(addrOrder, addrLower)
		}

		// Pack data for tryAggregate(requireSuccess=false, calls)
		// Note: abi.Pack accepts the function name and arguments
		packed, err := parsed.Pack("tryAggregate", false, calls)
		if err != nil {
			return nil, err
		}

		payloadHex := "0x" + hex.EncodeToString(packed)

		// eth_call
		reqBody := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "eth_call",
			"params": []interface{}{
				map[string]string{"to": mcAddr, "data": payloadHex},
				"latest",
			},
		}
		b, _ := json.Marshal(reqBody)

		// Create multicall result for tracking
		mcResult := MulticallResult{
			Success:     false,
			ReturnData:  nil,
			GasUsed:     0,
			GasLimit:    0,
			PoolAddress: mcAddr,
			RouteID:     fmt.Sprintf("multicall_batch_%d", i/batchSize),
			AmountIn:    nil,
			BlockNumber: 0,
		}

		// Retryable HTTP POST with exponential backoff + jitter
		rand.Seed(time.Now().UnixNano())
		retries, baseMs := getRPCRetryParams()
		var rb []byte
		var lastErr error
		for attempt := 0; attempt <= retries; attempt++ {
			req, err := http.NewRequestWithContext(ctx, "POST", rpc, bytes.NewReader(b))
			if err != nil {
				lastErr = err
				mcResult.ErrorMessage = err.Error()
			} else {
				req.Header.Set("Content-Type", "application/json")
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					lastErr = err
					mcResult.ErrorMessage = err.Error()
				} else {
					func() {
						defer resp.Body.Close()
						if resp.StatusCode >= 500 {
							lastErr = fmt.Errorf("rpc http status %d", resp.StatusCode)
							mcResult.ErrorMessage = lastErr.Error()
							return
						}
						body, err := io.ReadAll(resp.Body)
						if err != nil {
							lastErr = err
							mcResult.ErrorMessage = err.Error()
							return
						}
						rb = body
						lastErr = nil
						mcResult.Success = true
					}()
				}
			}

			if lastErr == nil {
				break
			}
			if attempt < retries {
				backoff := time.Duration(baseMs*(1<<attempt)) * time.Millisecond
				jitter := time.Duration(rand.Intn(baseMs)) * time.Millisecond
				time.Sleep(backoff + jitter)
			}
		}

		// Classify the multicall result
		if lastErr != nil {
			mcResult.ErrorMessage = lastErr.Error()
			classified := ClassifyMulticallResults([]MulticallResult{mcResult}, 0.90)
			fmt.Printf("Multicall batch %d failed with status: %s, reason: %s\n",
				i/batchSize, classified[0].Status.String(), classified[0].Reason)
			return nil, lastErr
		}

		var r struct {
			Result string `json:"result"`
		}
		if err := json.Unmarshal(rb, &r); err != nil {
			return nil, err
		}
		if r.Result == "" || r.Result == "0x" {
			continue
		}

		decoded, err := hex.DecodeString(strings.TrimPrefix(r.Result, "0x"))
		if err != nil {
			return nil, err
		}

		// Unpack outputs for tryAggregate
		vals, err := parsed.Unpack("tryAggregate", decoded)
		if err != nil {
			return nil, err
		}
		if len(vals) == 0 {
			continue
		}

		// vals[0] should be a slice of tuples (success, returnData)
		retSlice, ok := vals[0].([]interface{})
		if !ok {
			continue
		}

		// Iterate return tuples in order and map back to addresses
		for idx, item := range retSlice {
			if idx >= len(addrOrder) {
				break
			}
			addr := addrOrder[idx]
			// item is a tuple -> []interface{}{success(bool), returnData([]byte)}
			tuple, ok := item.([]interface{})
			if !ok || len(tuple) < 2 {
				continue
			}
			success, _ := tuple[0].(bool)
			if !success {
				continue
			}
			var retBytes []byte
			// returnData may be []byte or string; support both
			switch v := tuple[1].(type) {
			case []byte:
				retBytes = v
			case string:
				retBytes, _ = hex.DecodeString(strings.TrimPrefix(v, "0x"))
			default:
				continue
			}

			// Determine if this item was name or decimals by looking at callData order: even indices are name, odd are decimals
			isName := (idx%2 == 0)
			m := outMap[addr]
			if isName {
				// try dynamic string decode
				s := decodeABIString("0x" + hex.EncodeToString(retBytes))
				if s != "" {
					m.Name = s
				} else if len(retBytes) >= 32 {
					n := bytesTrimZero(retBytes[:32])
					m.Name = string(n)
				}
			} else {
				if v, err := decodeUint256("0x" + hex.EncodeToString(retBytes)); err == nil {
					m.Decimals = v
				}
			}
			outMap[addr] = m
		}
	}

	// prune empty
	for k, v := range outMap {
		if v.Name == "" && v.Decimals == 0 {
			delete(outMap, k)
		}
	}

	if len(outMap) == 0 {
		return nil, fmt.Errorf("no on-chain metadata available via multicall")
	}
	return outMap, nil
}

// EnsureTokensMetadataBatch fills missing Name/Decimals for multiple tokens using local tokenlist and
// a batched on-chain JSON-RPC eth_call (acts like a multicall for name() and decimals()).
// It does best-effort: first uses local tokenlist, then batch on-chain, then per-token on-chain as fallback.
func (p *GraphQLSubgraphProvider) EnsureTokensMetadataBatch(ctx context.Context, tokens []*types.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	// load local token list once
	loadTokenList()

	// group tokens by lowercased address for dedup
	addrToTokens := make(map[string][]*types.Token)
	for _, t := range tokens {
		if t == nil {
			continue
		}
		if t.Name != "" && t.Decimals != 0 {
			continue
		}
		addr := strings.ToLower(t.ID)

		// try tokenlist first
		if tokenListMap != nil {
			if m, ok := tokenListMap[addr]; ok {
				if t.Name == "" {
					t.Name = m.Name
				}
				if t.Decimals == 0 {
					t.Decimals = m.Decimals
				}
				if t.Name != "" && t.Decimals != 0 {
					continue
				}
			}
		}

		addrToTokens[addr] = append(addrToTokens[addr], t)
	}

	if len(addrToTokens) == 0 {
		return nil
	}

	// prepare address list
	addrs := make([]string, 0, len(addrToTokens))
	for a := range addrToTokens {
		addrs = append(addrs, a)
	}

	// 1) try multicall on-chain (contract-based)
	if multicallRes, err := p.fetchTokensOnChainMulticall(ctx, addrs); err == nil {
		for addr, meta := range multicallRes {
			if toks, ok := addrToTokens[addr]; ok {
				for _, tt := range toks {
					if tt.Name == "" {
						tt.Name = meta.Name
					}
					if tt.Decimals == 0 {
						tt.Decimals = meta.Decimals
					}
				}
				delete(addrToTokens, addr)
			}
		}
	} else {
		// fall back to JSON-RPC batch (use per-chain RPC)
		batchRes, err := fetchTokensOnChainBatch(ctx, p.chainID, addrs)
		if err == nil {
			for addr, meta := range batchRes {
				if toks, ok := addrToTokens[addr]; ok {
					for _, tt := range toks {
						if tt.Name == "" {
							tt.Name = meta.Name
						}
						if tt.Decimals == 0 {
							tt.Decimals = meta.Decimals
						}
					}
					delete(addrToTokens, addr)
				}
			}
		} else {
			// per-token fallback
			for addr, toks := range addrToTokens {
				meta, err2 := fetchTokenOnChain(ctx, addr)
				if err2 != nil {
					continue
				}
				for _, tt := range toks {
					if tt.Name == "" {
						tt.Name = meta.Name
					}
					if tt.Decimals == 0 {
						tt.Decimals = meta.Decimals
					}
				}
				delete(addrToTokens, addr)
			}
		}
	}

	// If any addresses remain unresolved, return an error indicating partial failure
	if len(addrToTokens) > 0 {
		return fmt.Errorf("some token metadata could not be resolved on-chain: %d tokens", len(addrToTokens))
	}
	return nil
}

// fetchTokensOnChainBatch performs a batched JSON-RPC eth_call for multiple tokens' name() and decimals()
// It returns a map keyed by lowercased token address to tokenMeta. Uses per-chain RPC via getRPCForChain.
func fetchTokensOnChainBatch(ctx context.Context, chainID types.ChainID, tokenAddrs []string) (map[string]tokenMeta, error) {
	rpc := getRPCForChain(chainID)
	if rpc == "" {
		return nil, fmt.Errorf("no RPC URL for chain %d", chainID)
	}

	// Prepare batch requests: for each token we push two eth_call requests (name, decimals)
	reqs := make([]map[string]interface{}, 0, len(tokenAddrs)*2)
	idMap := make(map[int]struct {
		addrLower string
		typ       string // "name" or "decimals"
	})
	idCounter := 1
	for _, a := range tokenAddrs {
		addrLower := strings.ToLower(a)
		// name()
		reqs = append(reqs, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      idCounter,
			"method":  "eth_call",
			"params": []interface{}{
				map[string]string{"to": addrLower, "data": "0x06fdde03"},
				"latest",
			},
		})
		idMap[idCounter] = struct {
			addrLower string
			typ       string
		}{addrLower, "name"}
		idCounter++

		// decimals()
		reqs = append(reqs, map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      idCounter,
			"method":  "eth_call",
			"params": []interface{}{
				map[string]string{"to": addrLower, "data": "0x313ce567"},
				"latest",
			},
		})
		idMap[idCounter] = struct {
			addrLower string
			typ       string
		}{addrLower, "decimals"}
		idCounter++
	}

	b, err := json.Marshal(reqs)
	if err != nil {
		return nil, err
	}

	// Retryable HTTP POST with exponential backoff + jitter
	rand.Seed(time.Now().UnixNano())
	retries, baseMs := getRPCRetryParams()
	var rb []byte
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, "POST", rpc, strings.NewReader(string(b)))
		if err != nil {
			lastErr = err
		} else {
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				lastErr = err
			} else {
				func() {
					defer resp.Body.Close()
					if resp.StatusCode >= 500 {
						lastErr = fmt.Errorf("rpc http status %d", resp.StatusCode)
						return
					}
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						lastErr = err
						return
					}
					rb = body
					lastErr = nil
				}()
			}
		}

		if lastErr == nil {
			break
		}
		if attempt < retries {
			backoff := time.Duration(baseMs*(1<<attempt)) * time.Millisecond
			jitter := time.Duration(rand.Intn(baseMs)) * time.Millisecond
			time.Sleep(backoff + jitter)
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}

	raw := make([]map[string]interface{}, 0)
	if err := json.Unmarshal(rb, &raw); err != nil {
		return nil, err
	}

	result := make(map[string]tokenMeta)
	for _, item := range raw {
		// id may be float64
		idv, ok := item["id"]
		if !ok {
			continue
		}
		var idInt int
		switch v := idv.(type) {
		case float64:
			idInt = int(v)
		case int:
			idInt = v
		default:
			continue
		}

		info, exists := idMap[idInt]
		if !exists {
			continue
		}

		resHex := ""
		if r, ok := item["result"].(string); ok {
			resHex = r
		}

		m := result[info.addrLower]
		if info.typ == "name" {
			if resHex != "" && resHex != "0x" {
				if s := decodeABIString(resHex); s != "" {
					m.Name = s
				} else if bs, err := hex.DecodeString(strings.TrimPrefix(resHex, "0x")); err == nil && len(bs) >= 32 {
					n := bytesTrimZero(bs[:32])
					m.Name = string(n)
				}
			}
		} else if info.typ == "decimals" {
			if resHex != "" && resHex != "0x" {
				if v, err := decodeUint256(resHex); err == nil {
					m.Decimals = v
				}
			}
		}
		result[info.addrLower] = m
	}

	// Remove empty entries
	for k, v := range result {
		if v.Name == "" && v.Decimals == 0 {
			delete(result, k)
		}
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no on-chain metadata available via batch")
	}
	return result, nil
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

// savePoolsToFile writes a slice to a file under data/chain_<chainID>/filename.
// If the value is a slice, it writes NDJSON (one JSON object per line) to stream and avoid building a huge JSON string.
func (p *GraphQLSubgraphProvider) savePoolsToFile(filename string, v interface{}) error {
	dir := filepath.Join("data", fmt.Sprintf("chain_%d", p.chainID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Slice {
		enc := json.NewEncoder(f)
		// write as NDJSON: one object per line
		for i := 0; i < rv.Len(); i++ {
			if err := enc.Encode(rv.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	}

	// fallback: write full JSON
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

// sanitizeFilename converts a string into a filesystem-safe lowercase name
func sanitizeFilename(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, ",", "")

	// remove any remaining characters that are not alphanumeric or underscore or dash
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			out = append(out, r)
		}
	}
	return string(out)
}
