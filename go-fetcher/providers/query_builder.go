package providers

import (
	"fmt"
	"go-fetcher/types"
	"strings"
)

// QueryBuilder helps build dynamic GraphQL queries similar to smart-order-router
type QueryBuilder struct {
	protocol types.Protocol
	chainID  types.ChainID
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(protocol types.Protocol, chainID types.ChainID) *QueryBuilder {
	return &QueryBuilder{
		protocol: protocol,
		chainID:  chainID,
	}
}

// BuildV2Queries builds V2 queries similar to smart-order-router
func (qb *QueryBuilder) BuildV2Queries() []string {
	feiToken := "0x956f47f50a910163d8bf957cf5846d573e7f87ca"
	virtualToken := "0x0b3e328455c4059eeb9e3f84b5543f74e24e7e1b"

	queries := []string{
		// 1. FEI token pools (token0)
		qb.buildV2Query(fmt.Sprintf(`token0: "%s"`, feiToken)),
		// 2. FEI token pools (token1)
		qb.buildV2Query(fmt.Sprintf(`token1: "%s"`, feiToken)),
		// 3. High tracked reserve ETH pools
		qb.buildV2Query(`trackedReserveETH_gt: "0.025"`),
		// 4. High untracked USD pools (essentially disabled with max value)
		qb.buildV2Query(`reserveUSD_gt: "1.7976931348623157e+308"`),
	}

	// Add virtual pair pools for BASE chain
	if qb.chainID == types.Base {
		queries = append(queries,
			qb.buildV2Query(fmt.Sprintf(`token0: "%s"`, virtualToken)),
			qb.buildV2Query(fmt.Sprintf(`token1: "%s"`, virtualToken)),
		)
	}

	return queries
}

// BuildV3Queries builds V3 queries similar to smart-order-router
func (qb *QueryBuilder) BuildV3Queries() []string {
	return []string{
		// 1. High tracked ETH pools
		qb.buildV3Query(`totalValueLockedETH_gt: "0.01"`),
		// 2. V3 zero ETH pools (special V3 condition)
		qb.buildV3Query(`liquidity_gt: "0", totalValueLockedETH: "0"`),
	}
}

// BuildV4Queries builds V4 queries similar to smart-order-router
func (qb *QueryBuilder) BuildV4Queries() []string {
	return []string{
		// 1. High tracked ETH pools
		qb.buildV4Query(`totalValueLockedETH_gt: "0.01"`),
		// 2. V4 high liquidity pools
		qb.buildV4Query(`liquidity_gt: "0"`),
	}
}

// buildV2Query builds a V2 GraphQL query with given conditions
func (qb *QueryBuilder) buildV2Query(conditions string) string {
	return fmt.Sprintf(`
		query getPools($pageSize: Int!, $lastID: String!) {
			pairs(
				first: $pageSize
				where: { id_gt: $lastID, %s }
				orderBy: id
				orderDirection: asc
			) {
				id
				token0 {
					id
					symbol
				}
				token1 {
					id
					symbol
				}
				totalSupply
				trackedReserveETH
				reserveETH
				reserveUSD
			}
		}
	`, conditions)
}

// buildV3Query builds a V3 GraphQL query with given conditions
func (qb *QueryBuilder) buildV3Query(conditions string) string {
	return fmt.Sprintf(`
		query getPools($pageSize: Int!, $lastID: String!) {
			pools(
				first: $pageSize
				where: { id_gt: $lastID, %s }
				orderBy: id
				orderDirection: asc
			) {
				id
				token0 {
					id
					symbol
					name
					decimals
				}
				token1 {
					id
					symbol
					name
					decimals
				}
				feeTier
				liquidity
				totalValueLockedUSD
				totalValueLockedETH
				tickSpacing
			}
		}
	`, conditions)
}

// buildV4Query builds a V4 GraphQL query with given conditions
func (qb *QueryBuilder) buildV4Query(conditions string) string {
	return fmt.Sprintf(`
		query getPools($pageSize: Int!, $lastID: String!) {
			pools(
				first: $pageSize
				where: { id_gt: $lastID, %s }
				orderBy: id
				orderDirection: asc
			) {
				id
				token0 {
					id
					symbol
					name
					decimals
				}
				token1 {
					id
					symbol
					name
					decimals
				}
				feeTier
				tickSpacing
				hooks
				liquidity
				totalValueLockedUSD
				totalValueLockedETH
			}
		}
	`, conditions)
}

// GetPageSize returns the appropriate page size based on protocol and chain
func (qb *QueryBuilder) GetPageSize() int {
	// Use larger page size for V4 on BASE chain like smart-order-router
	if qb.protocol == types.V4 && qb.chainID == types.Base {
		return 3500 // BASE_V4_PAGE_SIZE from smart-order-router
	}
	return 1000 // Default PAGE_SIZE
}

// ShouldIncludeV2Pool determines if a V2 pool should be included based on smart-order-router logic
func (qb *QueryBuilder) ShouldIncludeV2Pool(pool types.V2Pool, threshold float64) bool {
	feiToken := strings.ToLower("0x956f47f50a910163d8bf957cf5846d573e7f87ca")
	virtualToken := strings.ToLower("0x0b3e328455c4059eeb9e3f84b5543f74e24e7e1b")

	// FEI token pools
	if pool.Token0.ID == feiToken || pool.Token1.ID == feiToken {
		return true
	}

	// Virtual pair pools (BASE chain only)
	if qb.chainID == types.Base {
		if pool.Token0.ID == virtualToken || pool.Token1.ID == virtualToken {
			return true
		}
	}

	// High tracked reserve ETH pools
	if pool.Reserve > threshold {
		return true
	}

	// High untracked USD pools (essentially disabled with max threshold)
	if pool.ReserveUSD > 1.7976931348623157e+308 {
		return true
	}

	return false
}

// ShouldIncludeV3Pool determines if a V3 pool should be included based on smart-order-router V3 logic
func (qb *QueryBuilder) ShouldIncludeV3Pool(pool types.V3Pool, threshold float64) bool {
	// Parse liquidity as float for comparison
	liquidity := parseFloatSafe(pool.Liquidity)

	// V3 specific logic: Include pools with liquidity > 0 AND totalValueLockedETH = 0
	// OR pools with totalValueLockedETH > threshold
	if (liquidity > 0 && pool.TotalValueLockedETH == 0) || pool.TotalValueLockedETH > threshold {
		return true
	}

	return false
}

// ShouldIncludeV4Pool determines if a V4 pool should be included based on smart-order-router V4 logic
func (qb *QueryBuilder) ShouldIncludeV4Pool(pool types.V4Pool, threshold float64) bool {
	// Parse liquidity as float for comparison
	liquidity := parseFloatSafe(pool.Liquidity)

	// V4 logic: Include pools with liquidity > 0 OR totalValueLockedETH > threshold
	if liquidity > 0 || pool.TotalValueLockedETH > threshold {
		return true
	}

	return false
}

// parseFloatSafe safely parses a string to float64, returns 0 if error
func parseFloatSafe(s string) float64 {
	val, err := parseFloat(s)
	if err != nil {
		return 0
	}
	return val
}
