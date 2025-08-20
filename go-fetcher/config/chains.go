package config

import "go-fetcher/types"

const API_KEY = "27b52db08151aa3014307993833f704c"

// SubgraphURLs maps chain IDs to their subgraph URLs with updated API key
var SubgraphURLs = map[types.ChainID]map[types.Protocol]string{
	types.Mainnet: {
		types.V2: "https://gateway-arbitrum.network.thegraph.com/api/" + API_KEY + "/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
		types.V3: "https://gateway-arbitrum.network.thegraph.com/api/" + API_KEY + "/subgraphs/id/5zvR82QoaXYFyDEKLZ9t6v9adgnptxYpKpSbxtgVENFV",
		types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	},
	// types.Optimism: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cghf4LfVqPiFw6fp6Y5X5Ubc8UpmUhSfJL82zpMFEjsR",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	// },
	types.Arbitrum: {
		types.V2: "https://gateway.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/CStW6CSQbHoXsgKuVCrk3uShGA4JX3CAzzv2x9zaGf8w",
		types.V3: "https://gateway.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/3V7ZY6muhxaQL5qvntX1CFXJ32W7BxXZTGTwmpH5J4t3",
		types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	},
	// types.Polygon: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/3hCPRGf4z88VC5rsBKU5AA9FBBq5nF3jbKJG7VZCbhjm",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	// },
	// types.Base: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/43Hwfi3dJSoGpyas9VwNoDAV7mmhZVMkCFNcJoFMorxe",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/9Bz5KDpGTnqWPH6sxCCwFaYsVN8WSwATp9VaLm8WdB8U",
	// },
	types.BSC: {
		types.V2: "https://gateway.thegraph.com/api/" + API_KEY + "/subgraphs/id/8EjCaWZumyAfN3wyB4QnibeeXaYS8i4sp1PiWT91AGrt",
		types.V3: "https://gateway.thegraph.com/api/" + API_KEY + "/subgraphs/id/F85MNzUGYqgSHSHRGgeVMNsdnW1KtZSVgFULumXRZTw2",
		types.V4: "https://gateway.thegraph.com/api/" + API_KEY + "/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	},
	// types.Avalanche: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/ELnTJ1JYVkJnSuBc3iCQMNaJMDYxWqEcdYZpydVkQD1t",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	// },
	// types.Celo: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/ESdrTJ14n8wxu2ARvmau6LbAo5TYjGPLFV11Uz2HK5z1",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	// },
	// types.Sepolia: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/EYCKATKGBKLWvSfwvBjzfCBmGwYNdVkduYXVivCsLRFu",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/5zvR82QoaXuFYDEKLZ9t6v9adgnptxYpKpSbxtgVENFV",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/27b52db08151aa3014307993833f704c/subgraphs/id/Cd2gEDVeqnjBn1hSeqFMitw8Q1iiyXHRKWOTLxHdHfX7",
	// },
	// types.Goerli: {
	// 	types.V2: "https://gateway-arbitrum.network.thegraph.com/api/0ae45f0bf40ae2e73119b44ccd755967/subgraphs/id/ELUcwgpm14LKPLrBRuVvPvNKHQ9HvwmtKgKSH6123cr7",
	// 	types.V3: "https://gateway-arbitrum.network.thegraph.com/api/0ae45f0bf40ae2e73119b44ccd755967/subgraphs/id/ELUcwgpm14LKPLrBRuVvPvNKHQ9HvwmtKgKSH6123cr7",
	// 	types.V4: "https://gateway-arbitrum.network.thegraph.com/api/0ae45f0bf40ae2e73119b44ccd755967/subgraphs/id/ELUcwgpm14LKPLrBRuVvPvNKHQ9HvwmtKgKSH6123cr7",
	// },
}

// IPFSCacheURLs maps chain IDs to their IPFS cache URLs
var IPFSCacheURLs = map[types.ChainID]map[types.Protocol]string{
	types.Mainnet: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/mainnet.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/mainnet.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/mainnet.json",
	},
	types.Optimism: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/optimism.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/optimism.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/optimism.json",
	},
	types.Arbitrum: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/arbitrum.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/arbitrum.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/arbitrum.json",
	},
	types.Polygon: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/polygon.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/polygon.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/polygon.json",
	},
	types.Base: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/base.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/base.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/base.json",
	},
	types.BSC: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/bsc.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/bsc.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/bsc.json",
	},
	types.Avalanche: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/avalanche.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/avalanche.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/avalanche.json",
	},
	types.Celo: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/celo.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/celo.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/celo.json",
	},
	types.Sepolia: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/sepolia.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/sepolia.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/sepolia.json",
	},
	types.Goerli: {
		types.V2: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v2/goerli.json",
		types.V3: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v3/goerli.json",
		types.V4: "https://cloudflare-ipfs.com/ipns/api.uniswap.org/v1/pools/v4/goerli.json",
	},
}

// GetSubgraphURL returns the subgraph URL for a given chain and protocol
func GetSubgraphURL(chainID types.ChainID, protocol types.Protocol) (string, bool) {
	if urls, exists := SubgraphURLs[chainID]; exists {
		if url, exists := urls[protocol]; exists {
			return url, true
		}
	}
	return "", false
}

// GetIPFSCacheURL returns the IPFS cache URL for a given chain and protocol
func GetIPFSCacheURL(chainID types.ChainID, protocol types.Protocol) (string, bool) {
	if urls, exists := IPFSCacheURLs[chainID]; exists {
		if url, exists := urls[protocol]; exists {
			return url, true
		}
	}
	return "", false
}

// SupportedChains returns a list of supported chain IDs
func SupportedChains() []types.ChainID {
	chains := make([]types.ChainID, 0, len(SubgraphURLs))
	for chainID := range SubgraphURLs {
		chains = append(chains, chainID)
	}
	return chains
}

// SupportedProtocols returns a list of supported protocols
func SupportedProtocols() []types.Protocol {
	return []types.Protocol{types.V2, types.V3, types.V4}
}
