package providers

import (
	"context"
	"go-fetcher/types"
)

// ProviderConfig contains configuration for providers
type ProviderConfig struct {
	BlockNumber *int64
	Timeout     int // in seconds
	Retries     int
	// Pagination and filtering
	FetchAll           bool
	PageSize           int
	MinTVLETH          float64
	MinTrackedReserve  float64
}

// SubgraphProvider defines the interface for getting pools from subgraph
type SubgraphProvider interface {
	GetPools(ctx context.Context, config *ProviderConfig) ([]types.Pool, error)
	GetProtocol() types.Protocol
	GetChainID() types.ChainID
}

// V2SubgraphProvider interface for V2 pools
type V2SubgraphProvider interface {
	SubgraphProvider
	GetV2Pools(ctx context.Context, config *ProviderConfig) ([]types.V2Pool, error)
}

// V3SubgraphProvider interface for V3 pools
type V3SubgraphProvider interface {
	SubgraphProvider
	GetV3Pools(ctx context.Context, config *ProviderConfig) ([]types.V3Pool, error)
}

// V4SubgraphProvider interface for V4 pools
type V4SubgraphProvider interface {
	SubgraphProvider
	GetV4Pools(ctx context.Context, config *ProviderConfig) ([]types.V4Pool, error)
}

// BaseProvider provides common functionality for all providers
type BaseProvider struct {
	chainID  types.ChainID
	protocol types.Protocol
}

// NewBaseProvider creates a new base provider
func NewBaseProvider(chainID types.ChainID, protocol types.Protocol) BaseProvider {
	return BaseProvider{
		chainID:  chainID,
		protocol: protocol,
	}
}

// GetProtocol returns the protocol
func (p BaseProvider) GetProtocol() types.Protocol {
	return p.protocol
}

// GetChainID returns the chain ID
func (p BaseProvider) GetChainID() types.ChainID {
	return p.chainID
}
