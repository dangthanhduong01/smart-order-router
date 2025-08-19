# Go Pool Fetcher

Một project Golang độc lập để lấy danh sách pools từ Uniswap subgraphs, được tách ra từ smart-order-router.

## 🚀 Quick Start

```bash
# Build project
make build

# Fetch V3 pools từ Mainnet
./bin/pool-fetcher -chain 1 -protocol v3 -top 5

# Fetch tất cả protocols từ Optimism
./bin/pool-fetcher -chain 10 -protocol all -top 3

# Export to JSON
./bin/pool-fetcher -chain 137 -protocol v3 -output polygon_v3_pools.json
```

## 📁 Cấu Trúc Project

```
go-fetcher/
├── types/              # Định nghĩa types cho pools
├── config/             # Cấu hình URLs cho các chains
├── providers/          # Các providers (URI, GraphQL, Static, Caching, Fallback)
├── examples/           # Examples sử dụng
├── test/               # Tests
├── pool_fetcher.go     # Main pool fetcher class
├── main.go             # CLI interface
├── go.mod              # Go modules
├── Makefile            # Build và test commands
├── Dockerfile          # Containerization
└── README.md           # Documentation
```

## 🎯 Tính Năng

- **Multi-Protocol Support**: V2, V3, V4 pools
- **Multi-Chain Support**: Mainnet, Optimism, Arbitrum, Polygon, Base, BSC, Avalanche, Celo, etc.
- **Fallback Providers**: Tự động thử các providers khác nhau
- **Caching**: Cache kết quả để tăng hiệu suất
- **CLI Interface**: Giao diện command line dễ sử dụng
- **JSON Export**: Export pools ra file JSON

## 🔧 Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test -v ./providers

# Build for production
make prod-build

# Run with Docker
make docker-build
make docker-run
```

## 📊 Demo

```bash
# Fetch và hiển thị top pools
./bin/pool-fetcher -chain 1 -protocol all -top 3

# Kết quả:
=== V2 Pools (2 total) ===
1. WBTC/WETH - TVL: $15000000.00
2. USDC/WETH - TVL: $2000000.00

=== V3 Pools (3 total) ===
1. WBTC/WETH - TVL: $25000000.00
2. USDT/WETH - TVL: $8000000.00
3. USDC/WETH - TVL: $5000000.00

=== V4 Pools (1 total) ===
1. USDC/WETH - TVL: $3000000.00
```

## 🏗️ Luồng Hoạt Động

1. **IPFS Cache**: Thử lấy từ IPFS cache trước (nhanh nhất)
2. **Static Provider**: Fallback về static provider với dữ liệu mẫu
3. **Caching**: Cache kết quả để tái sử dụng
4. **Error Handling**: Graceful degradation khi providers thất bại

## 📝 License

MIT License
