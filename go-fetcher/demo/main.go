package main

import (
	"context"
	"flag"
	"fmt"
	"go-fetcher/config"
	"go-fetcher/providers"
	"go-fetcher/types"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env if present (silently ignore errors)
	_ = godotenv.Load()

	fmt.Println("🚀 Testing Smart-Order-Router Compatible Pool Fetcher with Authentication")

	// Show available subgraphs
	config.PrintAvailableSubgraphs()

	// read chain id from CLI flag or env
	chainID := flag.Int("chain", 1, "chain id to fetch (default 1 = mainnet)")
	multicallTest := flag.Bool("multicall", false, "run a multicall token metadata test for known stablecoins on the selected chain")
	flag.Parse()
	if env := os.Getenv("CHAIN_ID"); env != "" {
		if v, err := strconv.Atoi(env); err == nil {
			*chainID = v
		}
	}

	cid := types.ChainID(*chainID)

	// If user requested multicall test, run it and exit
	if *multicallTest {
		testMulticallForChain(cid)
		return
	}

	// Test V3 pools từ chain input với API key
	testV3PoolsWithAuth(cid)

	// Test V2 pools từ chain input với API key
	testV2PoolsWithAuth(cid)

	// Demonstrate QueryBuilder features
	demonstrateQueryBuilder()
}

// testV3PoolsWithAuth tests V3 pools with authentication for the given chain ID
func testV3PoolsWithAuth(chainID types.ChainID) {
	fmt.Printf("\n📊 Testing V3 Pools for chain %d with API Key Authentication\n", chainID)

	// Get best subgraph configuration
	subgraphConfig := config.GetBestSubgraphConfig(chainID, types.V3)
	if subgraphConfig == nil {
		log.Printf("❌ No V3 subgraph configuration for chain %d", chainID)
		return
	}

	fmt.Printf("🔗 Using: %s\n", subgraphConfig.URL)
	fmt.Printf("🔐 Authenticated: %v\n", subgraphConfig.RequiresAuth)

	// Tạo GraphQL provider với hoặc không auth
	var provider *providers.GraphQLSubgraphProvider
	if subgraphConfig.RequiresAuth {
		provider = providers.NewGraphQLSubgraphProviderWithAuth(
			chainID,
			types.V3,
			subgraphConfig.URL,
			subgraphConfig.APIKey,
			30*time.Second,
			2,
		)
		fmt.Printf("✅ Created authenticated provider with API key: %s...\n", subgraphConfig.APIKey[:8])
	} else {
		provider = providers.NewGraphQLSubgraphProvider(
			chainID,
			types.V3,
			subgraphConfig.URL,
			30*time.Second,
			2,
		)
		fmt.Printf("✅ Created public provider (no auth)\n")
	}

	// Fetch pools với smart-order-router logic
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pools, err := provider.GetV3Pools(ctx, &providers.ProviderConfig{
		FetchAll:  false, // Chỉ lấy first page để test
		MinTVLETH: 0.01,  // Default threshold từ smart-order-router
	})

	if err != nil {
		log.Printf("❌ Error fetching V3 pools: %v", err)
		return
	}

	fmt.Printf("✅ Successfully fetched %d V3 pools from chain %d\n", len(pools), chainID)

	// Show sample pools
	if len(pools) > 0 {
		fmt.Printf("📝 Sample V3 pools:\n")
		for i, pool := range pools[:min(3, len(pools))] {
			fmt.Printf("  %d. Pool %s: %s/%s (Fee: %s, TVL ETH: %.4f)\n",
				i+1, pool.ID[:8], pool.Token0.Symbol, pool.Token1.Symbol, pool.FeeTier, pool.TotalValueLockedETH)
		}
	}

	// Analyze filtering results
	highTVLCount := 0
	zeroETHCount := 0
	for _, pool := range pools {
		if pool.TotalValueLockedETH > 0.01 {
			highTVLCount++
		} else if pool.TotalValueLockedETH == 0 {
			zeroETHCount++
		}
	}

	fmt.Printf("📈 Analysis: %d high TVL pools, %d zero ETH pools\n", highTVLCount, zeroETHCount)
}

// testV2PoolsWithAuth tests V2 pools with authentication for the given chain ID
func testV2PoolsWithAuth(chainID types.ChainID) {
	fmt.Printf("\n📊 Testing V2 Pools for chain %d with API Key Authentication\n", chainID)

	subgraphConfig := config.GetBestSubgraphConfig(chainID, types.V2)
	if subgraphConfig == nil {
		log.Printf("❌ No V2 subgraph configuration for chain %d", chainID)
		return
	}

	var provider *providers.GraphQLSubgraphProvider
	if subgraphConfig.RequiresAuth {
		provider = providers.NewGraphQLSubgraphProviderWithAuth(
			chainID,
			types.V2,
			subgraphConfig.URL,
			subgraphConfig.APIKey,
			30*time.Second,
			2,
		)
	} else {
		provider = providers.NewGraphQLSubgraphProvider(
			chainID,
			types.V2,
			subgraphConfig.URL,
			30*time.Second,
			2,
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pools, err := provider.GetV2Pools(ctx, &providers.ProviderConfig{
		FetchAll:          false,
		MinTrackedReserve: 0.025, // V2 threshold từ smart-order-router
	})

	if err != nil {
		log.Printf("❌ Error fetching V2 pools: %v", err)
		return
	}

	fmt.Printf("✅ Successfully fetched %d V2 pools from chain %d\n", len(pools), chainID)

	if len(pools) > 0 {
		fmt.Printf("📝 Sample V2 pools:\n")
		for i, pool := range pools[:min(3, len(pools))] {
			fmt.Printf("  %d. Pool %s: %s/%s (Reserve: %.4f ETH, USD: %.2f)\n",
				i+1, pool.ID[:8], pool.Token0.Symbol, pool.Token1.Symbol, pool.Reserve, pool.ReserveUSD)
		}
	}

	// Check for FEI token pools
	feiToken := "0x956f47f50a910163d8bf957cf5846d573e7f87ca"
	feiCount := 0
	for _, pool := range pools {
		if pool.Token0.ID == feiToken || pool.Token1.ID == feiToken {
			feiCount++
		}
	}

	fmt.Printf("🪙 Found %d FEI token pools (special handling)\n", feiCount)
}

// testMulticallForChain runs EnsureTokensMetadataBatch on a set of known stablecoin addresses for the given chain
func testMulticallForChain(chainID types.ChainID) {
	fmt.Printf("\n🧪 Running multicall token metadata test on chain %d\n", chainID)

	// create provider (subgraph URL optional; EnsureTokensMetadataBatch doesn't use GraphQL client)
	subgraphConfig := config.GetBestSubgraphConfig(chainID, types.V2)
	var subgraphURL string
	if subgraphConfig != nil {
		subgraphURL = subgraphConfig.URL
	}
	provider := providers.NewGraphQLSubgraphProvider(chainID, types.V2, subgraphURL, 30*time.Second, 2)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var addrs []string
	switch int(chainID) {
	case 56: // BSC
		addrs = []string{
			"0xe9e7cea3dedca5984780bafc599bd69add087d56", // BUSD
			"0x55d398326f99059ff775485246999027b3197955", // USDT
			"0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d", // USDC
			"0x1af3f329e8be154074d8769d1ffa4ee058b1dbc3", // DAI
		}
	case 1: // Ethereum Mainnet
		addrs = []string{
			"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC
			"0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT
			"0x6B175474E89094C44Da98b954EedeAC495271d0F", // DAI
			"0x4fabb145d64652a948d72533023f6e7a623c7c53", // BUSD (Binance USD on mainnet)
		}
	case 42161: // Arbitrum One
		addrs = []string{
			"0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8", // USDC
			"0xFd086bC7CD5C481DCC9C85eBe478A1C0b69FCbb9", // USDT
			"0xDA10009cBd5D07dd0CeCc66161FC93D7c9000da1", // DAI
		}
	default:
		fmt.Printf("⚠️  No predefined token list for chain %d\n", chainID)
		return
	}

	tokens := make([]*types.Token, 0, len(addrs))
	for _, a := range addrs {
		tokens = append(tokens, &types.Token{ID: strings.ToLower(a)})
	}

	// Call EnsureTokensMetadataBatch which will try tokenlist -> multicall -> batch -> per-token
	if err := provider.EnsureTokensMetadataBatch(ctx, tokens); err != nil {
		fmt.Printf("⚠️  EnsureTokensMetadataBatch returned error: %v\n", err)
	}

	fmt.Println("✅ Resolved token metadata (if available):")
	for _, t := range tokens {
		fmt.Printf(" - %s -> Name: %q, Decimals: %d\n", t.ID, t.Name, t.Decimals)
	}
}

func demonstrateQueryBuilder() {
	fmt.Println("\n🔧 Demonstrating QueryBuilder Features")

	// V3 Query Builder
	qb := providers.NewQueryBuilder(types.V3, types.Mainnet)

	fmt.Printf("📄 V3 Queries Count: %d\n", len(qb.BuildV3Queries()))
	fmt.Printf("📏 V3 Page Size: %d\n", qb.GetPageSize())

	// V4 Base Query Builder (different page size)
	qbV4 := providers.NewQueryBuilder(types.V4, types.Base)
	fmt.Printf("📄 V4 Base Queries Count: %d\n", len(qbV4.BuildV4Queries()))
	fmt.Printf("📏 V4 Base Page Size: %d\n", qbV4.GetPageSize())

	// V2 with virtual pairs
	qbV2 := providers.NewQueryBuilder(types.V2, types.Base)
	fmt.Printf("📄 V2 Base Queries Count: %d (includes virtual pairs)\n", len(qbV2.BuildV2Queries()))

	// V2 without virtual pairs (other chains)
	qbV2Mainnet := providers.NewQueryBuilder(types.V2, types.Mainnet)
	fmt.Printf("📄 V2 Mainnet Queries Count: %d (no virtual pairs)\n", len(qbV2Mainnet.BuildV2Queries()))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
