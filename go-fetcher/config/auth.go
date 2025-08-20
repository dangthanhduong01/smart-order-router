package config

import (
	"fmt"
	"go-fetcher/types"
	"os"
)

// API configuration constants
const (
	// The Graph API key for authenticated access
	DEFAULT_API_KEY = "27b52db08151aa3014307993833f704c"

	// Environment variable name for API key
	API_KEY_ENV = "GRAPH_API_KEY"
)

// GetAPIKey returns the API key from environment or default
func GetAPIKey() string {
	if key := os.Getenv(API_KEY_ENV); key != "" {
		return key
	}
	return DEFAULT_API_KEY
}

// GetAuthenticatedSubgraphURL returns the authenticated subgraph URL for The Graph hosted service
func GetAuthenticatedSubgraphURL(chainID types.ChainID, protocol types.Protocol) (string, bool) {
	// Prefer the canonical mapping defined in chains.go (SubgraphURLs)
	if url, ok := GetSubgraphURL(chainID, protocol); ok {
		return url, true
	}
	return "", false
}

// GetFallbackSubgraphURL returns the public (free) subgraph URLs as fallback
func GetFallbackSubgraphURL(chainID types.ChainID, protocol types.Protocol) (string, bool) {
	// Public URLs (no API key required, but may have rate limits)
	publicURLs := map[types.ChainID]map[types.Protocol]string{
		types.Mainnet: {
			types.V2: "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v2-dev",
			types.V3: "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3",
		},
		types.Optimism: {
			types.V3: "https://api.thegraph.com/subgraphs/name/ianlapham/optimism-post-regenesis",
		},
		types.Arbitrum: {
			types.V3: "https://api.thegraph.com/subgraphs/name/ianlapham/arbitrum-minimal",
		},
		types.Polygon: {
			types.V3: "https://api.thegraph.com/subgraphs/name/ianlapham/uniswap-v3-polygon",
		},
	}

	if chainURLs, exists := publicURLs[chainID]; exists {
		if url, exists := chainURLs[protocol]; exists {
			return url, true
		}
	}

	return "", false
}

// SubgraphConfig represents configuration for subgraph access
type SubgraphConfig struct {
	URL          string
	RequiresAuth bool
	APIKey       string
	IsProduction bool
}

// GetBestSubgraphConfig returns the best available subgraph configuration
func GetBestSubgraphConfig(chainID types.ChainID, protocol types.Protocol) *SubgraphConfig {
	// Try authenticated URL first (better performance, higher rate limits)
	if url, exists := GetAuthenticatedSubgraphURL(chainID, protocol); exists {
		return &SubgraphConfig{
			URL:          url,
			RequiresAuth: true,
			APIKey:       GetAPIKey(),
			IsProduction: true,
		}
	}

	// Fallback to public URL
	if url, exists := GetFallbackSubgraphURL(chainID, protocol); exists {
		return &SubgraphConfig{
			URL:          url,
			RequiresAuth: false,
			APIKey:       "",
			IsProduction: false,
		}
	}

	return nil
}

// PrintAvailableSubgraphs prints all available subgraph configurations
func PrintAvailableSubgraphs() {
	fmt.Println("🔗 Available Authenticated Subgraphs (Production):")
	chains := []types.ChainID{types.Mainnet, types.Optimism, types.Arbitrum, types.Polygon, types.Base, types.BSC, types.Avalanche, types.Celo}
	protocols := []types.Protocol{types.V2, types.V3, types.V4}

	for _, chain := range chains {
		fmt.Printf("\n📍 %s (Chain ID: %d):\n", getChainName(chain), int64(chain))
		for _, protocol := range protocols {
			if _, exists := GetAuthenticatedSubgraphURL(chain, protocol); exists {
				fmt.Printf("  ✅ %s: Available\n", protocol)
			} else if _, exists := GetFallbackSubgraphURL(chain, protocol); exists {
				fmt.Printf("  🔄 %s: Fallback only\n", protocol)
			} else {
				fmt.Printf("  ❌ %s: Not available\n", protocol)
			}
		}
	}
}

// getChainName returns human-readable chain name
func getChainName(chainID types.ChainID) string {
	names := map[types.ChainID]string{
		types.Mainnet:   "Ethereum Mainnet",
		types.Optimism:  "Optimism",
		types.Arbitrum:  "Arbitrum One",
		types.Polygon:   "Polygon",
		types.Base:      "Base",
		types.BSC:       "BNB Smart Chain",
		types.Avalanche: "Avalanche",
		types.Celo:      "Celo",
		types.Sepolia:   "Sepolia Testnet",
		types.Goerli:    "Goerli Testnet",
	}

	if name, exists := names[chainID]; exists {
		return name
	}
	return fmt.Sprintf("Chain %d", int64(chainID))
}
