package providers

import (
	"fmt"
	"math/big"
	"strings"
	"sync"
)

// CallResultStatus represents the classification of a multicall result
type CallResultStatus int

const (
	CallSuccess CallResultStatus = iota
	CallOutOfGas
	CallRevert
	CallDecodeError
	CallInsufficientLiquidity
	CallUnknownError
)

func (s CallResultStatus) String() string {
	switch s {
	case CallSuccess:
		return "SUCCESS"
	case CallOutOfGas:
		return "OUT_OF_GAS"
	case CallRevert:
		return "REVERT"
	case CallDecodeError:
		return "DECODE_ERROR"
	case CallInsufficientLiquidity:
		return "INSUFFICIENT_LIQUIDITY"
	case CallUnknownError:
		return "UNKNOWN_ERROR"
	default:
		return "UNDEFINED"
	}
}

// MulticallResult represents a single multicall response
type MulticallResult struct {
	Success      bool
	ReturnData   []byte
	GasUsed      uint64
	GasLimit     uint64
	PoolAddress  string
	RouteID      string
	AmountIn     *big.Int
	BlockNumber  uint64
	ErrorMessage string
}

// ClassifiedResult contains the original result plus classification
type ClassifiedResult struct {
	Original  MulticallResult
	Status    CallResultStatus
	AmountOut *big.Int // decoded if successful
	Reason    string   // detailed reason for classification
}

// PoolBlacklist manages per-request pool blacklisting
type PoolBlacklist struct {
	mu            sync.RWMutex
	blacklisted   map[string]CallResultStatus // poolAddress -> reason
	failureCounts map[string]int              // poolAddress -> failure count
	maxFailures   int                         // threshold before blacklisting
}

// NewPoolBlacklist creates a new pool blacklist manager
func NewPoolBlacklist(maxFailures int) *PoolBlacklist {
	return &PoolBlacklist{
		blacklisted:   make(map[string]CallResultStatus),
		failureCounts: make(map[string]int),
		maxFailures:   maxFailures,
	}
}

// IsBlacklisted checks if a pool is blacklisted
func (pb *PoolBlacklist) IsBlacklisted(poolAddress string) (bool, CallResultStatus) {
	pb.mu.RLock()
	defer pb.mu.RUnlock()
	status, exists := pb.blacklisted[poolAddress]
	return exists, status
}

// RecordFailure records a failure for a pool and potentially blacklists it
func (pb *PoolBlacklist) RecordFailure(poolAddress string, status CallResultStatus) bool {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// Don't count successes as failures
	if status == CallSuccess {
		return false
	}

	pb.failureCounts[poolAddress]++

	// Blacklist immediately for certain critical failures
	if status == CallRevert || status == CallInsufficientLiquidity {
		pb.blacklisted[poolAddress] = status
		return true
	}

	// Blacklist after reaching failure threshold for other failures
	if pb.failureCounts[poolAddress] >= pb.maxFailures {
		pb.blacklisted[poolAddress] = status
		return true
	}

	return false
}

// GetBlacklistedPools returns all currently blacklisted pools
func (pb *PoolBlacklist) GetBlacklistedPools() map[string]CallResultStatus {
	pb.mu.RLock()
	defer pb.mu.RUnlock()

	result := make(map[string]CallResultStatus)
	for addr, status := range pb.blacklisted {
		result[addr] = status
	}
	return result
}

// ClassifyMulticallResults classifies a batch of multicall results
func ClassifyMulticallResults(results []MulticallResult, gasLimitThreshold float64) []ClassifiedResult {
	classified := make([]ClassifiedResult, len(results))

	for i, result := range results {
		classified[i] = ClassifiedResult{
			Original: result,
			Status:   classifyResult(result, gasLimitThreshold),
		}

		// Try to decode amountOut if successful
		if classified[i].Status == CallSuccess && len(result.ReturnData) >= 32 {
			// Assume first 32 bytes is amountOut (adjust based on your quoter ABI)
			classified[i].AmountOut = new(big.Int).SetBytes(result.ReturnData[:32])

			// Additional check: if amountOut is 0, classify as insufficient liquidity
			if classified[i].AmountOut.Cmp(big.NewInt(0)) == 0 {
				classified[i].Status = CallInsufficientLiquidity
				classified[i].Reason = "amountOut is zero"
			}
		}
	}

	return classified
}

// classifyResult classifies a single multicall result
func classifyResult(result MulticallResult, gasLimitThreshold float64) CallResultStatus {
	// Basic failure case
	if !result.Success {
		// Check for common revert patterns
		if strings.Contains(strings.ToLower(result.ErrorMessage), "out of gas") {
			return CallOutOfGas
		}
		if strings.Contains(strings.ToLower(result.ErrorMessage), "revert") {
			return CallRevert
		}
		if len(result.ReturnData) == 0 {
			return CallRevert
		}
		return CallUnknownError
	}

	// Success case - check for out of gas risk
	if result.GasLimit > 0 {
		gasUsageRatio := float64(result.GasUsed) / float64(result.GasLimit)
		if gasUsageRatio >= gasLimitThreshold {
			return CallOutOfGas
		}
	}

	// Check for decode issues
	if len(result.ReturnData) < 32 {
		return CallDecodeError
	}

	return CallSuccess
}

// UpdateBlacklistFromResults updates the blacklist based on classified results
func UpdateBlacklistFromResults(blacklist *PoolBlacklist, results []ClassifiedResult) {
	for _, result := range results {
		wasBlacklisted := blacklist.RecordFailure(result.Original.PoolAddress, result.Status)
		if wasBlacklisted {
			fmt.Printf("Pool %s blacklisted due to %s\n",
				result.Original.PoolAddress, result.Status.String())
		}
	}
}

// FilterResultsByBlacklist removes results for blacklisted pools
func FilterResultsByBlacklist(results []ClassifiedResult, blacklist *PoolBlacklist) []ClassifiedResult {
	filtered := make([]ClassifiedResult, 0, len(results))

	for _, result := range results {
		if isBlacklisted, _ := blacklist.IsBlacklisted(result.Original.PoolAddress); !isBlacklisted {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// GetSuccessfulResults returns only successful results with valid amountOut
func GetSuccessfulResults(results []ClassifiedResult, minAmountOut *big.Int) []ClassifiedResult {
	successful := make([]ClassifiedResult, 0, len(results))

	for _, result := range results {
		if result.Status == CallSuccess &&
			result.AmountOut != nil &&
			result.AmountOut.Cmp(minAmountOut) >= 0 {
			successful = append(successful, result)
		}
	}

	return successful
}

// MulticallStats provides statistics about multicall results
type MulticallStats struct {
	TotalCalls           int
	SuccessfulCalls      int
	OutOfGasCalls        int
	RevertCalls          int
	DecodeErrorCalls     int
	InsufficientLiqCalls int
	UnknownErrorCalls    int
	SuccessRate          float64
	AvgGasUsed           uint64
	BlacklistedPools     int
}

// CalculateStats calculates statistics from classified results
func CalculateStats(results []ClassifiedResult, blacklist *PoolBlacklist) MulticallStats {
	stats := MulticallStats{
		TotalCalls: len(results),
	}

	var totalGasUsed uint64
	statusCounts := make(map[CallResultStatus]int)

	for _, result := range results {
		statusCounts[result.Status]++
		totalGasUsed += result.Original.GasUsed
	}

	stats.SuccessfulCalls = statusCounts[CallSuccess]
	stats.OutOfGasCalls = statusCounts[CallOutOfGas]
	stats.RevertCalls = statusCounts[CallRevert]
	stats.DecodeErrorCalls = statusCounts[CallDecodeError]
	stats.InsufficientLiqCalls = statusCounts[CallInsufficientLiquidity]
	stats.UnknownErrorCalls = statusCounts[CallUnknownError]

	if stats.TotalCalls > 0 {
		stats.SuccessRate = float64(stats.SuccessfulCalls) / float64(stats.TotalCalls)
		stats.AvgGasUsed = totalGasUsed / uint64(stats.TotalCalls)
	}

	stats.BlacklistedPools = len(blacklist.GetBlacklistedPools())

	return stats
}

// Example usage helper
func ProcessMulticallBatch(results []MulticallResult, blacklist *PoolBlacklist) ([]ClassifiedResult, MulticallStats) {
	// Classify results (consider calls using >90% of gas limit as OOG risk)
	classified := ClassifyMulticallResults(results, 0.90)

	// Update blacklist based on failures
	UpdateBlacklistFromResults(blacklist, classified)

	// Filter out blacklisted pools for future batches
	filtered := FilterResultsByBlacklist(classified, blacklist)

	// Calculate statistics
	stats := CalculateStats(classified, blacklist)

	return filtered, stats
}
