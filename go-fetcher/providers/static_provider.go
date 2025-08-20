package providers

import (
	"context"
	"fmt"
	"go-fetcher/types"
)

// StaticSubgraphProvider provides static pool data for testing
type StaticSubgraphProvider struct {
	BaseProvider
}

// NewStaticSubgraphProvider creates a new static provider
func NewStaticSubgraphProvider(chainID types.ChainID, protocol types.Protocol) *StaticSubgraphProvider {
	return &StaticSubgraphProvider{
		BaseProvider: NewBaseProvider(chainID, protocol),
	}
}

// GetPools returns static pool data
func (p *StaticSubgraphProvider) GetPools(ctx context.Context, config *ProviderConfig) ([]types.Pool, error) {
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

// GetV2Pools returns static V2 pool data
func (p *StaticSubgraphProvider) GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error) {
	// Return some sample V2 pools
	return []types.V2Pool{
		{
			ID: "0x88e6a0c2ddd26feeb64f039a2c41296fcb3f5640",
			Token0: types.Token{
				ID:     "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
				Symbol: "USDC",
			},
			Token1: types.Token{
				ID:     "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Symbol: "WETH",
			},
			Reserve:     1000000.0,
			ReserveUSD:  2000000.0,
			TotalSupply: "1000000000000000000",
		},
		{
			ID: "0xbb2b8038a1640196fbe3e38816f3e67cba72d940",
			Token0: types.Token{
				ID:     "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
				Symbol: "WBTC",
			},
			Token1: types.Token{
				ID:     "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Symbol: "WETH",
			},
			Reserve:     500.0,
			ReserveUSD:  15000000.0,
			TotalSupply: "500000000000000000",
		},
	}, nil
}

// GetV3Pools returns static V3 pool data
func (p *StaticSubgraphProvider) GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error) {
	// Return some sample V3 pools
	return []types.V3Pool{
		{
			ID: "0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8",
			Token0: types.Token{
				ID:       "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 6,
			},
			Token1: types.Token{
				ID:       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Symbol:   "WETH",
				Name:     "Wrapped Ether",
				Decimals: 18,
			},
			FeeTier:             "3000",
			Liquidity:           "1000000000000000000",
			TotalValueLockedUSD: 5000000.0,
			TotalValueLockedETH: 2500.0,
			TickSpacing:         "60",
		},
		{
			ID: "0x4e68ccd3e89f51c3074ca5072bbac773960dfa36",
			Token0: types.Token{
				ID:       "0xdac17f958d2ee523a2206206994597c13d831ec7",
				Symbol:   "USDT",
				Name:     "Tether USD",
				Decimals: 6,
			},
			Token1: types.Token{
				ID:       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Symbol:   "WETH",
				Name:     "Wrapped Ether",
				Decimals: 18,
			},
			FeeTier:             "3000",
			Liquidity:           "2000000000000000000",
			TotalValueLockedUSD: 8000000.0,
			TotalValueLockedETH: 4000.0,
			TickSpacing:         "60",
		},
		{
			ID: "0x7d7e436f0b2aafde74ddee4e66141d556b9a8f5e",
			Token0: types.Token{
				ID:       "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599",
				Symbol:   "WBTC",
				Name:     "Wrapped Bitcoin",
				Decimals: 8,
			},
			Token1: types.Token{
				ID:       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Symbol:   "WETH",
				Name:     "Wrapped Ether",
				Decimals: 18,
			},
			FeeTier:             "500",
			Liquidity:           "500000000000000000",
			TotalValueLockedUSD: 25000000.0,
			TotalValueLockedETH: 12500.0,
			TickSpacing:         "10",
		},
	}, nil
}

// GetV4Pools returns static V4 pool data
func (p *StaticSubgraphProvider) GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error) {
	// Return some sample V4 pools
	return []types.V4Pool{
		{
			ID: "0x1234567890123456789012345678901234567890",
			Token0: types.Token{
				ID:       "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
				Symbol:   "USDC",
				Name:     "USD Coin",
				Decimals: 6,
			},
			Token1: types.Token{
				ID:       "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2",
				Symbol:   "WETH",
				Name:     "Wrapped Ether",
				Decimals: 18,
			},
			FeeTier:             "3000",
			TickSpacing:         "60",
			Hooks:               "0x0000000000000000000000000000000000000000",
			Liquidity:           "1000000000000000000",
			TotalValueLockedUSD: 3000000.0,
			TotalValueLockedETH: 1500.0,
		},
	}, nil
}
