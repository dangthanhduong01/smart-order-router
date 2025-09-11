package providers

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"go-fetcher/types"
)

// RouteCandidate is a lightweight representation of a route composed of one or more pools.
// This is intended for cheap off-chain prefiltering and coarse sampling before any
// on-chain multicall is performed.
//
// This implementation follows an "Uniswap-inspired" approach: we use TotalTVLUSD as a
// conservative proxy for route liquidity, aggregate fee tiers as a simple cost, and
// a cheap constant-product based price-impact approximation for coarse ranking.
type RouteCandidate struct {
	ID               string
	Pools            []types.Pool
	TotalTVLUSD      float64 // summed TVL (USD) across pools as a proxy for liquidity
	AggregatedFeeBps int     // sum of fees in basis points across the route
}

// NewRouteCandidate builds a RouteCandidate from given pools. ID can be any stable identifier.
func NewRouteCandidate(id string, pools []types.Pool) RouteCandidate {
	tvl := 0.0
	fee := 0
	for _, p := range pools {
		tvl += p.GetTVLUSD()
		fee += poolFeeBps(p)
	}
	return RouteCandidate{
		ID:               id,
		Pools:            pools,
		TotalTVLUSD:      tvl,
		AggregatedFeeBps: fee,
	}
}

// poolFeeBps returns an approximate fee in basis points for the pool.
// V3/V4 store fee tier as string (e.g. "3000" for 0.3%). V2 doesn't expose fee in our types
// so we default to 30 bps (0.3%).
func poolFeeBps(p types.Pool) int {
	switch v := p.(type) {
	case types.V3Pool:
		if v.FeeTier == "" {
			return 30
		}
		if n, err := strconv.Atoi(strings.TrimSpace(v.FeeTier)); err == nil {
			// In subgraph data FeeTier often comes as string like "3000" (meaning 0.3%).
			// We convert to basis points conservatively.
			if n >= 1000 {
				// Treat "3000" -> 30 bps
				return n / 100
			}
			return n
		}
		return 30
	case types.V4Pool:
		if v.FeeTier == "" {
			return 30
		}
		if n, err := strconv.Atoi(strings.TrimSpace(v.FeeTier)); err == nil {
			if n >= 1000 {
				return n / 100
			}
			return n
		}
		return 30
	case types.V2Pool:
		// V2 default fee 0.3% -> 30 bps
		return 30
	default:
		// Unknown pool type, assume conservative 30 bps
		return 30
	}
}

// PrefilterRouteCandidates performs very cheap, O(1) checks per route to drop obviously bad routes.
// It filters by minimum TVL (USD) and a maximum aggregated fee (bps). Both checks are simple numeric
// comparisons and do not perform any network calls.
func PrefilterRouteCandidates(cands []RouteCandidate, minTVLUSD float64, maxAggregatedFeeBps int) []RouteCandidate {
	out := make([]RouteCandidate, 0, len(cands))
	for _, c := range cands {
		if c.TotalTVLUSD < minTVLUSD {
			// skip low-liquidity routes
			continue
		}
		if c.AggregatedFeeBps > maxAggregatedFeeBps {
			// skip routes with too large fees
			continue
		}
		out = append(out, c)
	}
	return out
}

// effectiveLiquidityForRoute returns a conservative estimate of the route liquidity (in USD).
// Uniswap implementations compute more accurate effective liquidity using on-chain state and ticks.
// Here we use TotalTVLUSD with a floor to avoid divide-by-zero and extremely noisy estimates.
func effectiveLiquidityForRoute(c RouteCandidate) float64 {
	if c.TotalTVLUSD <= 0 {
		return 1e-6
	}
	return c.TotalTVLUSD
}

// estimateOutputFactorForPercent computes a cheap approximation of the expected output factor
// (output / input) for swapping an amount equal to percent% of the route's effective liquidity.
// This uses a conservative constant-product inspired model:
//   - assume route reserve ~= effectiveLiquidity / 2 (approx per-side liquidity)
//   - amountUSD = percent/100 * effectiveLiquidity
//   - priceImpact ~= amountUSD / (reserve + amountUSD)
//   - fee is subtracted (aggregated fee fraction)
//
// The returned value is a fraction in [0,1] representing estimated output per unit input.
func estimateOutputFactorForPercent(c RouteCandidate, percent float64) float64 {
	liq := effectiveLiquidityForRoute(c)
	reserve := liq / 2.0
	amountUSD := (percent / 100.0) * liq
	if reserve <= 0 {
		return 0.0
	}
	// price impact estimate (conservative)
	priceImpact := amountUSD / (reserve + amountUSD)
	// fee fraction from aggregated bps
	feeFrac := float64(c.AggregatedFeeBps) / 10000.0
	estimated := 1.0 - priceImpact - feeFrac
	if estimated < 0 {
		return 0.0
	}
	return estimated
}

// CoarseSampleAndSelect performs a cheap approximate evaluation for each route across a set of
// distributionPercents (e.g. [5,10,15,...]) and returns the top K candidates with the best
// estimated output factor (higher is better). This function uses simple heuristics and a rough
// price impact model derived from route TVL as a proxy for liquidity. It is intentionally
// conservative to avoid over-ranking risky routes.
func CoarseSampleAndSelect(cands []RouteCandidate, distributionPercents []float64, topK int) ([]RouteCandidate, error) {
	if topK <= 0 {
		return nil, fmt.Errorf("topK must be > 0")
	}
	if len(cands) == 0 {
		return []RouteCandidate{}, nil
	}

	type scored struct {
		cand  RouteCandidate
		score float64
	}

	sc := make([]scored, 0, len(cands))

	for _, c := range cands {
		avg := 0.0
		count := 0
		for _, p := range distributionPercents {
			if p <= 0 {
				continue
			}
			est := estimateOutputFactorForPercent(c, p)
			avg += est
			count++
		}
		if count > 0 {
			avg = avg / float64(count)
		}
		// Slight tie-break using TVL (prefer deeper routes)
		score := avg + math.Log1p(effectiveLiquidityForRoute(c))*1e-9
		sc = append(sc, scored{cand: c, score: score})
	}

	// sort by descending score
	sort.Slice(sc, func(i, j int) bool { return sc[i].score > sc[j].score })

	if topK > len(sc) {
		topK = len(sc)
	}
	out := make([]RouteCandidate, 0, topK)
	for i := 0; i < topK; i++ {
		out = append(out, sc[i].cand)
	}
	return out, nil
}

// BuildCandidatesFromPools converts a list of raw routes (each route is a slice of types.Pool)
// to RouteCandidates with stable IDs. This mirrors how an on-chain router would enumerate
// possible routes before ranking them.
func BuildCandidatesFromPools(routes [][]types.Pool) []RouteCandidate {
	out := make([]RouteCandidate, 0, len(routes))
	for i, r := range routes {
		id := fmt.Sprintf("route_%d_%d", len(r), i)
		out = append(out, NewRouteCandidate(id, r))
	}
	return out
}

// SelectTopKRoutes is a convenience pipeline function that performs Prefilter -> CoarseSample -> TopK.
// Useful to plug directly into the quoting pipeline before on-chain multicall.
func SelectTopKRoutes(routes [][]types.Pool, minTVLUSD float64, maxAggregatedFeeBps int, distributionPercents []float64, topK int) ([]RouteCandidate, error) {
	cands := BuildCandidatesFromPools(routes)
	filtered := PrefilterRouteCandidates(cands, minTVLUSD, maxAggregatedFeeBps)
	selected, err := CoarseSampleAndSelect(filtered, distributionPercents, topK)
	if err != nil {
		return nil, err
	}
	return selected, nil
}
