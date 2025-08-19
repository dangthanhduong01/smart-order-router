# Uniswap Pool Fetcher

Một project Golang để lấy danh sách pools từ Uniswap subgraphs, tương tự như smart-order-router của Uniswap. Project này implement luồng fallback providers và caching để đảm bảo tính ổn định và hiệu suất cao.

## Tính Năng

- **Multi-Protocol Support**: Hỗ trợ V2, V3, và V4 pools
- **Multi-Chain Support**: Hỗ trợ nhiều blockchain networks (Mainnet, Optimism, Arbitrum, Polygon, Base, BSC, Avalanche, Celo, etc.)
- **Fallback Providers**: Tự động thử các providers khác nhau khi một provider thất bại
- **Caching**: Cache kết quả để tăng hiệu suất
- **IPFS Cache**: Sử dụng IPFS cache để lấy dữ liệu nhanh nhất
- **GraphQL Fallback**: Fallback về GraphQL subgraph khi IPFS cache không khả dụng
- **CLI Interface**: Giao diện command line dễ sử dụng

## Cấu Trúc Project

```
uniswap-pool-fetcher/
├── types/
│   └── pool.go              # Định nghĩa các types cho pools
├── config/
│   └── chains.go            # Cấu hình URLs cho các chains
├── providers/
│   ├── provider.go          # Interface cho providers
│   ├── uri_provider.go      # Provider lấy từ IPFS cache
│   ├── graphql_provider.go  # Provider query GraphQL subgraph
│   ├── fallback_provider.go # Provider với fallback chain
│   └── caching_provider.go  # Provider với caching
├── pool_fetcher.go          # Main pool fetcher class
├── main.go                  # CLI interface
├── go.mod                   # Go modules
└── README.md               # Documentation
```

## Luồng Hoạt Động

1. **Khởi tạo Providers**: Tạo fallback chain cho mỗi protocol (V2, V3, V4)
2. **IPFS Cache First**: Thử lấy từ IPFS cache trước (nhanh nhất)
3. **GraphQL Fallback**: Nếu IPFS cache thất bại, thử query GraphQL subgraph
4. **Caching**: Cache kết quả để tái sử dụng
5. **Error Handling**: Xử lý lỗi và retry logic

## Cài Đặt

```bash
# Clone repository
git clone <repository-url>
cd uniswap-pool-fetcher

# Install dependencies
go mod tidy

# Build project
go build -o pool-fetcher .
```

## Sử Dụng

### Command Line Interface

```bash
# Fetch tất cả pools từ Mainnet
./pool-fetcher -chain 1 -protocol all

# Fetch chỉ V3 pools từ Optimism
./pool-fetcher -chain 10 -protocol v3

# Fetch V2 pools từ Arbitrum và lưu vào file
./pool-fetcher -chain 42161 -protocol v2 -output arbitrum_v2_pools.json

# Hiển thị top 20 V4 pools từ Base với verbose logging
./pool-fetcher -chain 8453 -protocol v4 -top 20 -verbose
```

### Programmatic Usage

```go
package main

import (
    "log"
    "uniswap-pool-fetcher/types"
    "uniswap-pool-fetcher/providers"
)

func main() {
    // Tạo pool fetcher cho Mainnet
    fetcher, err := NewPoolFetcher(types.Mainnet)
    if err != nil {
        log.Fatal(err)
    }

    // Fetch V3 pools
    pools, err := fetcher.GetPools(types.V3, nil)
    if err != nil {
        log.Fatal(err)
    }

    // In ra số lượng pools
    log.Printf("Found %d V3 pools", len(pools))

    // Fetch tất cả protocols
    allPools, err := fetcher.GetAllPools(nil)
    if err != nil {
        log.Fatal(err)
    }

    for protocol, pools := range allPools {
        log.Printf("%s: %d pools", protocol, len(pools))
    }
}
```

## Các Chains Hỗ Trợ

| Chain ID | Network | V2 | V3 | V4 |
|----------|---------|----|----|----|
| 1 | Mainnet | ✅ | ✅ | ✅ |
| 10 | Optimism | ✅ | ✅ | ✅ |
| 42161 | Arbitrum | ✅ | ✅ | ✅ |
| 137 | Polygon | ✅ | ✅ | ✅ |
| 8453 | Base | ✅ | ✅ | ✅ |
| 56 | BSC | ✅ | ✅ | ✅ |
| 43114 | Avalanche | ✅ | ✅ | ✅ |
| 42220 | Celo | ✅ | ✅ | ✅ |
| 11155111 | Sepolia | ✅ | ✅ | ✅ |
| 5 | Goerli | ✅ | ✅ | ✅ |

## Các Protocols Hỗ Trợ

### V2 Pools
- **Token Pairs**: Standard ERC20 token pairs
- **AMM**: Constant Product AMM
- **Fee**: 0.3% trading fee

### V3 Pools
- **Concentrated Liquidity**: Liquidity providers can concentrate their capital within custom price ranges
- **Multiple Fee Tiers**: 0.01%, 0.05%, 0.3%, 1%
- **NFT Positions**: Liquidity positions are represented as NFTs

### V4 Pools
- **Hooks**: Custom logic that can be executed during swaps
- **Dynamic Fees**: Fees can be adjusted based on market conditions
- **Singleton Architecture**: All pools in a single contract

## Cấu Hình

### Provider Config

```go
config := &providers.ProviderConfig{
    BlockNumber: &blockNumber, // Optional: specific block number
    Timeout:     30,           // Timeout in seconds
    Retries:     2,            // Number of retries
}
```

### Cache Configuration

- **Default TTL**: 5 minutes cho IPFS cache
- **Cleanup Interval**: 1 minute
- **Cache Keys**: Dựa trên chain ID, protocol, và block number

## Error Handling

Project implement comprehensive error handling:

- **Network Errors**: Retry logic với exponential backoff
- **Timeout Handling**: Configurable timeouts cho mỗi request
- **Fallback Chain**: Tự động thử provider khác khi một provider thất bại
- **Graceful Degradation**: Vẫn hoạt động khi một số providers không khả dụng

## Performance

- **IPFS Cache**: ~100-500ms response time
- **GraphQL Subgraph**: ~1-3s response time
- **Caching**: Subsequent requests served from cache (~1ms)
- **Concurrent Requests**: Support multiple concurrent requests

## Monitoring

Project includes logging và metrics:

```go
// Set log level
fetcher.SetLogLevel(logrus.DebugLevel)

// Log messages include:
// - Provider selection
// - Request timing
// - Error details
// - Cache hits/misses
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -v ./providers -run TestURISubgraphProvider
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

MIT License

## Tương Đồng Với Smart-Order-Router

Project này implement tương tự luồng code của Uniswap smart-order-router:

1. **Fallback Chain**: URISubgraphProvider → GraphQLSubgraphProvider
2. **Caching**: CachingSubgraphProvider với NodeCache
3. **Provider Interfaces**: Tương tự ISubgraphProvider
4. **Error Handling**: Retry logic và graceful degradation
5. **Configuration**: Chain-specific URLs và settings

## Roadmap

- [ ] On-chain pool provider (static pools)
- [ ] Pool filtering và sorting options
- [ ] WebSocket support cho real-time updates
- [ ] Metrics và monitoring dashboard
- [ ] Docker containerization
- [ ] API server mode
