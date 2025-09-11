package providers

import (
	"math/big"
	"testing"
)

func TestCallResultClassification(t *testing.T) {
	tests := []struct {
		name     string
		result   MulticallResult
		expected CallResultStatus
	}{
		{
			name: "successful call with valid return data",
			result: MulticallResult{
				Success:     true,
				ReturnData:  make([]byte, 64), // 64 bytes of zeros
				GasUsed:     50000,
				GasLimit:    100000,
				PoolAddress: "0x1234567890123456789012345678901234567890",
			},
			expected: CallSuccess,
		},
		{
			name: "out of gas - high gas usage",
			result: MulticallResult{
				Success:     true,
				ReturnData:  make([]byte, 64),
				GasUsed:     95000,
				GasLimit:    100000, // 95% usage
				PoolAddress: "0x1234567890123456789012345678901234567890",
			},
			expected: CallOutOfGas,
		},
		{
			name: "failed call with out of gas message",
			result: MulticallResult{
				Success:      false,
				ErrorMessage: "execution reverted: out of gas",
				PoolAddress:  "0x1234567890123456789012345678901234567890",
			},
			expected: CallOutOfGas,
		},
		{
			name: "failed call with revert message",
			result: MulticallResult{
				Success:      false,
				ErrorMessage: "execution reverted: insufficient liquidity",
				PoolAddress:  "0x1234567890123456789012345678901234567890",
			},
			expected: CallRevert,
		},
		{
			name: "successful call with zero amount out",
			result: MulticallResult{
				Success:     true,
				ReturnData:  make([]byte, 32), // 32 bytes of zeros
				GasUsed:     50000,
				GasLimit:    100000,
				PoolAddress: "0x1234567890123456789012345678901234567890",
			},
			expected: CallSuccess, // Will be reclassified to CallInsufficientLiquidity by ClassifyMulticallResults
		},
		{
			name: "decode error - insufficient return data",
			result: MulticallResult{
				Success:     true,
				ReturnData:  []byte{0x12, 0x34}, // Less than 32 bytes
				GasUsed:     50000,
				GasLimit:    100000,
				PoolAddress: "0x1234567890123456789012345678901234567890",
			},
			expected: CallDecodeError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyResult(tt.result, 0.90)
			if result != tt.expected {
				t.Errorf("classifyResult() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClassifyMulticallResults(t *testing.T) {
	results := []MulticallResult{
		{
			Success:     true,
			ReturnData:  make([]byte, 32), // Zero amount out
			GasUsed:     50000,
			GasLimit:    100000,
			PoolAddress: "0x1111111111111111111111111111111111111111",
		},
		{
			Success:     true,
			ReturnData:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b}, // 123 in hex
			GasUsed:     50000,
			GasLimit:    100000,
			PoolAddress: "0x2222222222222222222222222222222222222222",
		},
	}

	classified := ClassifyMulticallResults(results, 0.90)

	if len(classified) != 2 {
		t.Fatalf("Expected 2 classified results, got %d", len(classified))
	}

	// First result should be classified as insufficient liquidity due to zero amountOut
	if classified[0].Status != CallInsufficientLiquidity {
		t.Errorf("Expected first result to be %v, got %v", CallInsufficientLiquidity, classified[0].Status)
	}

	// Second result should be successful
	if classified[1].Status != CallSuccess {
		t.Errorf("Expected second result to be %v, got %v", CallSuccess, classified[1].Status)
	}

	// Check amountOut decoding
	if classified[1].AmountOut == nil {
		t.Error("Expected amountOut to be decoded for successful result")
	} else if classified[1].AmountOut.Cmp(big.NewInt(123)) != 0 {
		t.Errorf("Expected amountOut to be 123, got %s", classified[1].AmountOut.String())
	}
}

func TestPoolBlacklist(t *testing.T) {
	blacklist := NewPoolBlacklist(2) // Blacklist after 2 failures
	poolAddr := "0x1234567890123456789012345678901234567890"

	// Initially not blacklisted
	isBlacklisted, _ := blacklist.IsBlacklisted(poolAddr)
	if isBlacklisted {
		t.Error("Pool should not be blacklisted initially")
	}

	// Record first failure - should not blacklist yet
	wasBlacklisted := blacklist.RecordFailure(poolAddr, CallOutOfGas)
	if wasBlacklisted {
		t.Error("Pool should not be blacklisted after first failure")
	}

	// Record second failure - should blacklist now
	wasBlacklisted = blacklist.RecordFailure(poolAddr, CallOutOfGas)
	if !wasBlacklisted {
		t.Error("Pool should be blacklisted after second failure")
	}

	// Check blacklist status
	isBlacklisted, reason := blacklist.IsBlacklisted(poolAddr)
	if !isBlacklisted {
		t.Error("Pool should be blacklisted")
	}
	if reason != CallOutOfGas {
		t.Errorf("Expected blacklist reason to be %v, got %v", CallOutOfGas, reason)
	}
}

func TestPoolBlacklistImmediateBlacklist(t *testing.T) {
	blacklist := NewPoolBlacklist(5) // High threshold
	poolAddr := "0x1234567890123456789012345678901234567890"

	// Record a revert - should blacklist immediately
	wasBlacklisted := blacklist.RecordFailure(poolAddr, CallRevert)
	if !wasBlacklisted {
		t.Error("Pool should be blacklisted immediately for revert")
	}

	isBlacklisted, reason := blacklist.IsBlacklisted(poolAddr)
	if !isBlacklisted {
		t.Error("Pool should be blacklisted")
	}
	if reason != CallRevert {
		t.Errorf("Expected blacklist reason to be %v, got %v", CallRevert, reason)
	}
}

func TestFilterResultsByBlacklist(t *testing.T) {
	blacklist := NewPoolBlacklist(1)

	// Blacklist one pool
	pool1 := "0x1111111111111111111111111111111111111111"
	pool2 := "0x2222222222222222222222222222222222222222"
	blacklist.RecordFailure(pool1, CallRevert)

	results := []ClassifiedResult{
		{
			Original: MulticallResult{PoolAddress: pool1},
			Status:   CallSuccess,
		},
		{
			Original: MulticallResult{PoolAddress: pool2},
			Status:   CallSuccess,
		},
	}

	filtered := FilterResultsByBlacklist(results, blacklist)

	if len(filtered) != 1 {
		t.Fatalf("Expected 1 filtered result, got %d", len(filtered))
	}

	if filtered[0].Original.PoolAddress != pool2 {
		t.Errorf("Expected filtered result to be for pool %s, got %s", pool2, filtered[0].Original.PoolAddress)
	}
}

func TestGetSuccessfulResults(t *testing.T) {
	results := []ClassifiedResult{
		{
			Status:    CallSuccess,
			AmountOut: big.NewInt(100),
		},
		{
			Status:    CallOutOfGas,
			AmountOut: big.NewInt(200),
		},
		{
			Status:    CallSuccess,
			AmountOut: big.NewInt(50),
		},
		{
			Status:    CallSuccess,
			AmountOut: nil, // No amountOut
		},
	}

	successful := GetSuccessfulResults(results, big.NewInt(75))

	if len(successful) != 1 {
		t.Fatalf("Expected 1 successful result, got %d", len(successful))
	}

	if successful[0].AmountOut.Cmp(big.NewInt(100)) != 0 {
		t.Errorf("Expected amountOut to be 100, got %s", successful[0].AmountOut.String())
	}
}

func TestCalculateStats(t *testing.T) {
	blacklist := NewPoolBlacklist(1)
	blacklist.RecordFailure("0x1111111111111111111111111111111111111111", CallRevert)

	results := []ClassifiedResult{
		{Status: CallSuccess, Original: MulticallResult{GasUsed: 50000}},
		{Status: CallSuccess, Original: MulticallResult{GasUsed: 60000}},
		{Status: CallOutOfGas, Original: MulticallResult{GasUsed: 90000}},
		{Status: CallRevert, Original: MulticallResult{GasUsed: 30000}},
	}

	stats := CalculateStats(results, blacklist)

	expectedStats := MulticallStats{
		TotalCalls:           4,
		SuccessfulCalls:      2,
		OutOfGasCalls:        1,
		RevertCalls:          1,
		DecodeErrorCalls:     0,
		InsufficientLiqCalls: 0,
		UnknownErrorCalls:    0,
		SuccessRate:          0.5,
		AvgGasUsed:           57500, // (50000 + 60000 + 90000 + 30000) / 4
		BlacklistedPools:     1,
	}

	if stats != expectedStats {
		t.Errorf("Stats mismatch.\nExpected: %+v\nGot: %+v", expectedStats, stats)
	}
}

func TestProcessMulticallBatch(t *testing.T) {
	blacklist := NewPoolBlacklist(1)

	results := []MulticallResult{
		{
			Success:     true,
			ReturnData:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x7b}, // 123
			GasUsed:     50000,
			GasLimit:    100000,
			PoolAddress: "0x1111111111111111111111111111111111111111",
		},
		{
			Success:      false,
			ErrorMessage: "execution reverted",
			PoolAddress:  "0x2222222222222222222222222222222222222222",
		},
	}

	filtered, stats := ProcessMulticallBatch(results, blacklist)

	// Should have 1 successful result, 1 blacklisted
	if len(filtered) != 1 {
		t.Fatalf("Expected 1 filtered result, got %d", len(filtered))
	}

	if stats.SuccessfulCalls != 1 {
		t.Errorf("Expected 1 successful call, got %d", stats.SuccessfulCalls)
	}

	if stats.BlacklistedPools != 1 {
		t.Errorf("Expected 1 blacklisted pool, got %d", stats.BlacklistedPools)
	}

	// Check that the failed pool is blacklisted
	isBlacklisted, _ := blacklist.IsBlacklisted("0x2222222222222222222222222222222222222222")
	if !isBlacklisted {
		t.Error("Failed pool should be blacklisted")
	}
}
