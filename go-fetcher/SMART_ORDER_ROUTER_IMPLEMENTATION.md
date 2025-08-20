# GraphQL Provider Updates - Smart Order Router Implementation

## Tổng Quan

Project Go đã được cập nhật để query dữ liệu từ subgraph giống như cách **Uniswap smart-order-router** thực hiện. Các thay đổi chính bao gồm:

## 1. Multiple Queries Strategy

### V2 Pools - 4-6 Queries
Thay vì 1 query đơn giản, giờ sử dụng nhiều queries song song:

```go
// 1. FEI token pools (token0)
// 2. FEI token pools (token1)  
// 3. High tracked reserve ETH pools
// 4. High untracked USD pools
// 5. Virtual pair pools (token0) - chỉ BASE chain
// 6. Virtual pair pools (token1) - chỉ BASE chain
```

### V3 Pools - 2 Queries
```go
// 1. High tracked ETH pools (totalValueLockedETH_gt: "0.01")
// 2. V3 zero ETH pools (liquidity_gt: "0", totalValueLockedETH: "0")
```

### V4 Pools - 2 Queries
```go
// 1. High tracked ETH pools (totalValueLockedETH_gt: "0.01")
// 2. V4 high liquidity pools (liquidity_gt: "0")
```

## 2. QueryBuilder Component

Tạo file `query_builder.go` để:
- Build dynamic GraphQL queries
- Xử lý logic filtering cho từng protocol
- Quản lý page size (3500 cho V4 BASE chain, 1000 default)
- Implement smart filtering logic

## 3. Deduplication Logic

```go
poolMap := make(map[string]types.V2Pool) // For deduplication
// Merge results from multiple queries
// Remove duplicates by pool ID
```

## 4. Advanced Filtering

### V2 Filtering
- FEI token pools (đặc biệt include)
- Virtual pair pools (chỉ BASE chain)
- High tracked reserve ETH pools (> 0.025)
- High untracked USD pools (> max value - essentially disabled)

### V3 Filtering (Đặc Biệt)
```go
// Include pools với liquidity > 0 AND totalValueLockedETH = 0
// OR pools với totalValueLockedETH > threshold
if (liquidity > 0 && pool.TotalValueLockedETH == 0) || 
   pool.TotalValueLockedETH > threshold {
    return true
}
```

### V4 Filtering
```go
// Include pools với liquidity > 0 OR totalValueLockedETH > threshold
if liquidity > 0 || pool.TotalValueLockedETH > threshold {
    return true
}
```

## 5. Protocol-Specific Logic

### V2 Đặc Biệt
- **FEI Token**: `0x956f47f50a910163d8bf957cf5846d573e7f87ca`
- **Virtual Token**: `0x0b3e328455c4059eeb9e3f84b5543f74e24e7e1b` (BASE chain)
- **Threshold**: 0.025 ETH tracked reserve

### V3 Đặc Biệt
- **Zero ETH Pools**: Pools có `totalValueLockedETH = 0` nhưng `liquidity > 0`
- **Threshold**: 0.01 ETH
- **Reason**: V3 có vấn đề với derivedETH calculation

### V4 Đặc Biệt
- **Base Chain Page Size**: 3500 (thay vì 1000)
- **Liquidity Focus**: Ưu tiên pools có liquidity > 0
- **Threshold**: 0.01 ETH

## 6. Parallel Execution

```go
// Fetch pools for each query in parallel
poolPromises := queries.map(queryConfig => 
    fetchPoolsForQuery(queryConfig)
)
allPoolsArrays := Promise.all(poolPromises)
```

## 7. Error Handling & Retry

- **Retry Logic**: 3 attempts với exponential backoff
- **Timeout Handling**: Configurable timeout per request
- **Graceful Degradation**: Continue với other queries nếu 1 query fails

## 8. Performance Optimizations

- **Pagination**: 1000 per page (3500 cho V4 BASE)
- **Parallel Queries**: All queries run concurrently
- **Memory Efficient**: Stream processing với pagination
- **Cache Ready**: Structure sẵn sàng cho caching layer

## 9. Smart-Order-Router Compatibility

### Tương Đồng 100%
- ✅ Multiple query strategy
- ✅ Protocol-specific filtering
- ✅ FEI token special handling
- ✅ Virtual pair handling (BASE)
- ✅ V3 zero ETH pool logic
- ✅ V4 liquidity-focused logic
- ✅ Deduplication by pool ID
- ✅ Page size optimization

### Điểm Khác Biệt
- **Language**: Go thay vì TypeScript
- **GraphQL Client**: hasura/go-graphql-client thay vì graphql-request
- **Error Handling**: Go error handling pattern
- **Type Safety**: Go struct-based thay vì interface

## 10. Usage Example

```go
// Tạo provider
provider := NewGraphQLSubgraphProvider(
    types.Mainnet, 
    types.V3, 
    "https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v3",
    30*time.Second,
    2,
)

// Fetch pools với smart-order-router logic
pools, err := provider.GetV3Pools(ctx, &ProviderConfig{
    FetchAll: true,
    MinTVLETH: 0.01,
})
```

## 11. Future Improvements

1. **Dynamic GraphQL Query Building**: Thay vì static queries
2. **Metrics & Monitoring**: Add prometheus metrics
3. **Caching Layer**: Redis/in-memory cache
4. **Circuit Breaker**: Fault tolerance pattern
5. **Rate Limiting**: Respect subgraph rate limits
