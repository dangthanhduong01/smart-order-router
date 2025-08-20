package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go-fetcher/providers"
	"go-fetcher/types"
	"os"

	"github.com/sirupsen/logrus"
)

func main() {
	// Parse command line flags
	var (
		chainID           = flag.Int64("chain", 1, "Chain ID (1=Mainnet, 10=Optimism, 42161=Arbitrum, etc.)")
		protocol          = flag.String("protocol", "all", "Protocol to fetch (v2, v3, v4, all)")
		output            = flag.String("output", "", "Output file path (optional)")
		verbose           = flag.Bool("verbose", false, "Enable verbose logging")
		topN              = flag.Int("top", 10, "Show top N pools by TVL")
		fetchAll          = flag.Bool("fetch-all", false, "Fetch all pages from subgraph (may take time)")
		pageSize          = flag.Int("page-size", 1000, "GraphQL page size")
		minTVLETH         = flag.Float64("min-tvl-eth", 0.01, "Min TVL in ETH for V3/V4 filtering")
		minTrackedReserve = flag.Float64("min-tracked-reserve", 0.025, "Min tracked reserve ETH for V2 filtering")
	)
	flag.Parse()

	// Set log level
	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Create pool fetcher
	fetcher, err := NewPoolFetcher(types.ChainID(*chainID))
	if err != nil {
		logrus.Fatalf("Failed to create pool fetcher: %v", err)
	}

	// Set log level for fetcher
	if *verbose {
		fetcher.SetLogLevel(logrus.DebugLevel)
	}

	logrus.Infof("Fetching pools for chain %d", *chainID)

	// Build provider config
	providerConfig := &providers.ProviderConfig{
		FetchAll:          *fetchAll,
		PageSize:          *pageSize,
		MinTVLETH:         *minTVLETH,
		MinTrackedReserve: *minTrackedReserve,
	}

	// Fetch pools based on protocol
	switch *protocol {
	case "v2":
		pools, err := fetcher.GetPools(types.V2, providerConfig)
		if err != nil {
			logrus.Fatalf("Failed to fetch V2 pools: %v", err)
		}
		displayPools("V2", pools, *topN, *output)

	case "v3":
		pools, err := fetcher.GetPools(types.V3, providerConfig)
		if err != nil {
			logrus.Fatalf("Failed to fetch V3 pools: %v", err)
		}
		displayPools("V3", pools, *topN, *output)

	case "v4":
		pools, err := fetcher.GetPools(types.V4, providerConfig)
		if err != nil {
			logrus.Fatalf("Failed to fetch V4 pools: %v", err)
		}
		displayPools("V4", pools, *topN, *output)

	case "all":
		allPools, err := fetcher.GetAllPools(providerConfig)
		if err != nil {
			logrus.Fatalf("Failed to fetch all pools: %v", err)
		}

		for protocol, pools := range allPools {
			displayPools(string(protocol), pools, *topN, "")
		}

		// Save to file if specified
		if *output != "" {
			saveToFile(allPools, *output)
		}

	default:
		logrus.Fatalf("Unsupported protocol: %s. Use v2, v3, v4, or all", *protocol)
	}
}

// displayPools displays pools in a formatted way
func displayPools(protocol string, pools []types.Pool, topN int, outputFile string) {
	if len(pools) == 0 {
		logrus.Warnf("No %s pools found", protocol)
		return
	}

	logrus.Infof("\n=== %s Pools (%d total) ===", protocol, len(pools))

	// Sort pools by TVL (descending)
	sortedPools := make([]types.Pool, len(pools))
	copy(sortedPools, pools)

	// Simple bubble sort for top N pools
	for i := 0; i < topN && i < len(sortedPools); i++ {
		for j := i + 1; j < len(sortedPools); j++ {
			if sortedPools[j].GetTVLUSD() > sortedPools[i].GetTVLUSD() {
				sortedPools[i], sortedPools[j] = sortedPools[j], sortedPools[i]
			}
		}
	}

	// Display top N pools
	for i := 0; i < topN && i < len(sortedPools); i++ {
		pool := sortedPools[i]
		fmt.Printf("%d. %s/%s - TVL: $%.2f - ID: %s\n",
			i+1,
			pool.GetToken0().Symbol,
			pool.GetToken1().Symbol,
			pool.GetTVLUSD(),
			pool.GetID(),
		)
	}

	// Save to file if specified
	if outputFile != "" {
		savePoolsToFile(pools, outputFile)
	}
}

// saveToFile saves all pools to a JSON file
func saveToFile(allPools map[types.Protocol][]types.Pool, filename string) {
	data := make(map[string]interface{})

	for protocol, pools := range allPools {
		protocolStr := string(protocol)
		data[protocolStr] = pools
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal pools to JSON: %v", err)
		return
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		logrus.Errorf("Failed to write to file %s: %v", filename, err)
		return
	}

	logrus.Infof("Saved all pools to %s", filename)
}

// savePoolsToFile saves pools to a JSON file
func savePoolsToFile(pools []types.Pool, filename string) {
	jsonData, err := json.MarshalIndent(pools, "", "  ")
	if err != nil {
		logrus.Errorf("Failed to marshal pools to JSON: %v", err)
		return
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		logrus.Errorf("Failed to write to file %s: %v", filename, err)
		return
	}

	logrus.Infof("Saved pools to %s", filename)
}

// Example usage function
func exampleUsage() {
	fmt.Print(`
Uniswap Pool Fetcher - Fetch pools from Uniswap subgraphs

Usage:
  go run . [flags]

Flags:
  -chain int
        Chain ID (default 1)
        Supported chains: 1 (Mainnet), 10 (Optimism), 42161 (Arbitrum), 137 (Polygon), 8453 (Base), 56 (BSC), 43114 (Avalanche), 42220 (Celo)
  
  -protocol string
        Protocol to fetch: v2, v3, v4, or all (default "all")
  
  -output string
        Output file path for JSON export (optional)
  
  -verbose
        Enable verbose logging
  
  -top int
        Show top N pools by TVL (default 10)

  -fetch-all
        Fetch all pages from subgraph (may take time)

  -page-size
        GraphQL page size (default 1000)

  -min-tvl-eth
        Min TVL in ETH for V3/V4 filtering (default 0.01)

  -min-tracked-reserve
        Min tracked reserve ETH for V2 filtering (default 0.025)

Examples:
  # Fetch all pools from Mainnet
  go run . -chain 1 -protocol all -fetch-all

  # Fetch only V3 pools from Optimism, 500/page, TVL threshold 0
  go run . -chain 10 -protocol v3 -fetch-all -page-size 500 -min-tvl-eth 0

  # Fetch V2 pools from Arbitrum with low reserve threshold
  go run . -chain 42161 -protocol v2 -min-tracked-reserve 0

  # Show top 20 V4 pools from Base with verbose logging
  go run . -chain 8453 -protocol v4 -top 20 -verbose
`)
}
