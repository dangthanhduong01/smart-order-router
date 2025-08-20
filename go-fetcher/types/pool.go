package types

import (
	"math/big"
	"strings"
)

// ChainID represents different blockchain networks
type ChainID int64

const (
	Mainnet   ChainID = 1
	Optimism  ChainID = 10
	Arbitrum  ChainID = 42161
	Polygon   ChainID = 137
	Base      ChainID = 8453
	BSC       ChainID = 56
	Avalanche ChainID = 43114
	Celo      ChainID = 42220
	Sepolia   ChainID = 11155111
	Goerli    ChainID = 5
)

// Protocol represents different Uniswap versions
type Protocol string

const (
	V2 Protocol = "V2"
	V3 Protocol = "V3"
	V4 Protocol = "V4"
)

// Token represents a token in a pool
type Token struct {
	ID       string `json:"id"`
	Symbol   string `json:"symbol"`
	Name     string `json:"name"`
	Decimals int    `json:"decimals"`
}

// V2Pool represents a V2 pool from subgraph
type V2Pool struct {
	ID          string  `json:"id"`
	Token0      Token   `json:"token0"`
	Token1      Token   `json:"token1"`
	Reserve     float64 `json:"reserve"`
	ReserveUSD  float64 `json:"reserveUSD"`
	TotalSupply string  `json:"totalSupply"`
}

// V3Pool represents a V3 pool from subgraph
type V3Pool struct {
	ID                  string  `json:"id"`
	Token0              Token   `json:"token0"`
	Token1              Token   `json:"token1"`
	FeeTier             string  `json:"feeTier"`
	Liquidity           string  `json:"liquidity"`
	TotalValueLockedUSD float64 `json:"totalValueLockedUSD"`
	TotalValueLockedETH float64 `json:"totalValueLockedETH"`
	TickSpacing         string  `json:"tickSpacing"`
}

// V4Pool represents a V4 pool from subgraph
type V4Pool struct {
	ID                  string  `json:"id"`
	Token0              Token   `json:"token0"`
	Token1              Token   `json:"token1"`
	FeeTier             string  `json:"feeTier"`
	TickSpacing         string  `json:"tickSpacing"`
	Hooks               string  `json:"hooks"`
	Liquidity           string  `json:"liquidity"`
	TotalValueLockedUSD float64 `json:"totalValueLockedUSD"`
	TotalValueLockedETH float64 `json:"totalValueLockedETH"`
}

// Pool represents a generic pool interface
type Pool interface {
	GetID() string
	GetToken0() Token
	GetToken1() Token
	GetTVLUSD() float64
	GetProtocol() Protocol
}

// Implement Pool interface for V2Pool
func (p V2Pool) GetID() string         { return p.ID }
func (p V2Pool) GetToken0() Token      { return p.Token0 }
func (p V2Pool) GetToken1() Token      { return p.Token1 }
func (p V2Pool) GetTVLUSD() float64    { return p.ReserveUSD }
func (p V2Pool) GetProtocol() Protocol { return V2 }

// Implement Pool interface for V3Pool
func (p V3Pool) GetID() string         { return p.ID }
func (p V3Pool) GetToken0() Token      { return p.Token0 }
func (p V3Pool) GetToken1() Token      { return p.Token1 }
func (p V3Pool) GetTVLUSD() float64    { return p.TotalValueLockedUSD }
func (p V3Pool) GetProtocol() Protocol { return V3 }

// Implement Pool interface for V4Pool
func (p V4Pool) GetID() string         { return p.ID }
func (p V4Pool) GetToken0() Token      { return p.Token0 }
func (p V4Pool) GetToken1() Token      { return p.Token1 }
func (p V4Pool) GetTVLUSD() float64    { return p.TotalValueLockedUSD }
func (p V4Pool) GetProtocol() Protocol { return V4 }

// NormalizeAddress converts address to lowercase for consistent comparison
func NormalizeAddress(address string) string {
	return strings.ToLower(address)
}

// BigInt represents a big integer
type BigInt struct {
	*big.Int
}

// UnmarshalJSON implements json.Unmarshaler
func (b *BigInt) UnmarshalJSON(data []byte) error {
	// Remove quotes
	str := string(data)
	str = strings.Trim(str, "\"")

	// Parse as big.Int
	bi := new(big.Int)
	bi.SetString(str, 10)
	b.Int = bi
	return nil
}

// MarshalJSON implements json.Marshaler
func (b BigInt) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}
