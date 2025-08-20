package main

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"time"
// 	"go-fetcher/providers"
// 	"go-fetcher/types"

// 	"github.com/sirupsen/logrus"
// )

// // Import the main package to access NewPoolFetcher
// // Note: This is just for demonstration, in real usage you would import the package properly

// func main() {
// 	// Example 1: Basic usage
// 	basicExample()

// 	// Example 2: Fetch specific protocol
// 	specificProtocolExample()

// 	// Example 3: With configuration
// 	configuredExample()

// 	// Example 4: Error handling
// 	errorHandlingExample()

// 	// Example 5: Performance monitoring
// 	performanceExample()
// }

// func basicExample() {
// 	fmt.Println("\n=== Basic Example ===")

// 	// Create pool fetcher for Mainnet
// 	fetcher, err := NewPoolFetcher(types.Mainnet)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Fetch all pools
// 	allPools, err := fetcher.GetAllPools(nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Print summary
// 	for protocol, pools := range allPools {
// 		fmt.Printf("%s: %d pools\n", protocol, len(pools))
// 	}
// }

// func specificProtocolExample() {
// 	fmt.Println("\n=== Specific Protocol Example ===")

// 	// Create pool fetcher for Optimism
// 	fetcher, err := NewPoolFetcher(types.Optimism)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Fetch only V3 pools
// 	v3Pools, err := fetcher.GetPools(types.V3, nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("Found %d V3 pools on Optimism\n", len(v3Pools))

// 	// Show top 5 pools by TVL
// 	if len(v3Pools) > 0 {
// 		fmt.Println("Top 5 V3 pools by TVL:")
// 		for i := 0; i < 5 && i < len(v3Pools); i++ {
// 			pool := v3Pools[i]
// 			fmt.Printf("  %d. %s/%s - TVL: $%.2f\n",
// 				i+1,
// 				pool.GetToken0().Symbol,
// 				pool.GetToken1().Symbol,
// 				pool.GetTVLUSD(),
// 			)
// 		}
// 	}
// }

// func configuredExample() {
// 	fmt.Println("\n=== Configured Example ===")

// 	// Create pool fetcher for Arbitrum
// 	fetcher, err := NewPoolFetcher(types.Arbitrum)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Configure provider
// 	config := &providers.ProviderConfig{
// 		Timeout: 60, // 60 seconds timeout
// 		Retries: 3,  // 3 retries
// 	}

// 	// Fetch V2 pools with configuration
// 	v2Pools, err := fetcher.GetV2Pools(config)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("Found %d V2 pools on Arbitrum\n", len(v2Pools))

// 	// Show some pool details
// 	if len(v2Pools) > 0 {
// 		fmt.Println("Sample V2 pool:")
// 		pool := v2Pools[0]
// 		fmt.Printf("  ID: %s\n", pool.ID)
// 		fmt.Printf("  Tokens: %s (%s) / %s (%s)\n",
// 			pool.Token0.Symbol, pool.Token0.ID,
// 			pool.Token1.Symbol, pool.Token1.ID,
// 		)
// 		fmt.Printf("  Reserve: %.2f\n", pool.Reserve)
// 		fmt.Printf("  Reserve USD: %.2f\n", pool.ReserveUSD)
// 	}
// }

// func errorHandlingExample() {
// 	fmt.Println("\n=== Error Handling Example ===")

// 	// Try to create fetcher for unsupported chain
// 	_, err := NewPoolFetcher(999999) // Invalid chain ID
// 	if err != nil {
// 		fmt.Printf("Expected error for invalid chain: %v\n", err)
// 	}

// 	// Create fetcher for supported chain
// 	fetcher, err := NewPoolFetcher(types.Polygon)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Try to fetch with invalid protocol
// 	_, err = fetcher.GetPools("INVALID", nil)
// 	if err != nil {
// 		fmt.Printf("Expected error for invalid protocol: %v\n", err)
// 	}

// 	// This should timeout
// 	config := &providers.ProviderConfig{
// 		Timeout: 1, // 1 second timeout
// 		Retries: 0, // No retries
// 	}

// 	_, err = fetcher.GetPools(types.V3, config)
// 	if err != nil {
// 		fmt.Printf("Expected timeout error: %v\n", err)
// 	}
// }

// func performanceExample() {
// 	fmt.Println("\n=== Performance Example ===")

// 	// Create pool fetcher for Base
// 	fetcher, err := NewPoolFetcher(types.Base)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Set debug logging
// 	fetcher.SetLogLevel(logrus.DebugLevel)

// 	// Measure performance
// 	start := time.Now()

// 	// Fetch V4 pools
// 	v4Pools, err := fetcher.GetV4Pools(nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	duration := time.Since(start)
// 	fmt.Printf("Fetched %d V4 pools in %v\n", len(v4Pools), duration)

// 	// Show performance metrics
// 	if len(v4Pools) > 0 {
// 		fmt.Printf("Average time per pool: %v\n", duration/time.Duration(len(v4Pools)))
// 	}

// 	// Test caching performance
// 	fmt.Println("\nTesting cache performance...")

// 	// First request (cache miss)
// 	start = time.Now()
// 	_, err = fetcher.GetV4Pools(nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	firstDuration := time.Since(start)

// 	// Second request (cache hit)
// 	start = time.Now()
// 	_, err = fetcher.GetV4Pools(nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	secondDuration := time.Since(start)

// 	fmt.Printf("First request (cache miss): %v\n", firstDuration)
// 	fmt.Printf("Second request (cache hit): %v\n", secondDuration)
// 	fmt.Printf("Cache speedup: %.2fx\n", float64(firstDuration)/float64(secondDuration))
// }

// // Example of using the fetcher as a library
// func libraryExample() {
// 	fmt.Println("\n=== Library Usage Example ===")

// 	// Create fetcher
// 	fetcher, err := NewPoolFetcher(types.Mainnet)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Fetch pools and process them
// 	v3Pools, err := fetcher.GetV3Pools(nil)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// Process pools
// 	var highTVLPools []types.V3Pool
// 	var totalTVL float64

// 	for _, pool := range v3Pools {
// 		totalTVL += pool.TotalValueLockedUSD
// 		if pool.TotalValueLockedUSD > 1000000 { // $1M threshold
// 			highTVLPools = append(highTVLPools, pool)
// 		}
// 	}

// 	fmt.Printf("Total V3 pools: %d\n", len(v3Pools))
// 	fmt.Printf("High TVL pools (>$1M): %d\n", len(highTVLPools))
// 	fmt.Printf("Total TVL: $%.2f\n", totalTVL)

// 	// Export to JSON
// 	if len(highTVLPools) > 0 {
// 		jsonData, err := json.MarshalIndent(highTVLPools, "", "  ")
// 		if err != nil {
// 			log.Printf("Failed to marshal to JSON: %v", err)
// 		} else {
// 			fmt.Printf("High TVL pools JSON (first 500 chars): %s...\n", string(jsonData[:500]))
// 		}
// 	}
// }
