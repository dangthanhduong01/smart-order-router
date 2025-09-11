// TypeScript/JavaScript version for smart-order-router integration

export enum CallResultStatus {
  SUCCESS = 'SUCCESS',
  OUT_OF_GAS = 'OUT_OF_GAS',
  REVERT = 'REVERT',
  DECODE_ERROR = 'DECODE_ERROR',
  INSUFFICIENT_LIQUIDITY = 'INSUFFICIENT_LIQUIDITY',
  UNKNOWN_ERROR = 'UNKNOWN_ERROR'
}

export interface MulticallResult {
  success: boolean;
  returnData: string;
  gasUsed?: number;
  gasLimit?: number;
  poolAddress: string;
  routeId: string;
  amountIn: string;
  blockNumber: number;
  errorMessage?: string;
}

export interface ClassifiedResult {
  original: MulticallResult;
  status: CallResultStatus;
  amountOut?: string;
  reason?: string;
}

export class PoolBlacklist {
  private blacklisted: Map<string, CallResultStatus> = new Map();
  private failureCounts: Map<string, number> = new Map();
  private maxFailures: number;

  constructor(maxFailures: number = 3) {
    this.maxFailures = maxFailures;
  }

  isBlacklisted(poolAddress: string): { isBlacklisted: boolean; reason?: CallResultStatus } {
    const reason = this.blacklisted.get(poolAddress);
    return {
      isBlacklisted: reason !== undefined,
      reason
    };
  }

  recordFailure(poolAddress: string, status: CallResultStatus): boolean {
    if (status === CallResultStatus.SUCCESS) {
      return false;
    }

    const currentCount = this.failureCounts.get(poolAddress) || 0;
    this.failureCounts.set(poolAddress, currentCount + 1);

    // Immediate blacklist for critical failures
    if (status === CallResultStatus.REVERT || status === CallResultStatus.INSUFFICIENT_LIQUIDITY) {
      this.blacklisted.set(poolAddress, status);
      return true;
    }

    // Threshold-based blacklist for other failures
    if (currentCount + 1 >= this.maxFailures) {
      this.blacklisted.set(poolAddress, status);
      return true;
    }

    return false;
  }

  getBlacklistedPools(): Map<string, CallResultStatus> {
    return new Map(this.blacklisted);
  }

  reset(): void {
    this.blacklisted.clear();
    this.failureCounts.clear();
  }
}

export function classifyMulticallResults(
  results: MulticallResult[],
  gasLimitThreshold: number = 0.90
): ClassifiedResult[] {
  return results.map(result => {
    const status = classifyResult(result, gasLimitThreshold);
    const classified: ClassifiedResult = {
      original: result,
      status
    };

    // Try to decode amountOut if successful
    if (status === CallResultStatus.SUCCESS && result.returnData) {
      try {
        // Assume returnData is hex string like "0x000...amount..."
        // Adjust decoding based on your quoter contract ABI
        const cleanHex = result.returnData.startsWith('0x') 
          ? result.returnData.slice(2) 
          : result.returnData;
        
        if (cleanHex.length >= 64) { // at least 32 bytes for amountOut
          const amountOutHex = cleanHex.slice(0, 64);
          classified.amountOut = BigInt('0x' + amountOutHex).toString();
          
          // Check for zero output
          if (classified.amountOut === '0') {
            classified.status = CallResultStatus.INSUFFICIENT_LIQUIDITY;
            classified.reason = 'amountOut is zero';
          }
        }
      } catch (error) {
        classified.status = CallResultStatus.DECODE_ERROR;
        classified.reason = `Failed to decode amountOut: ${error.message}`;
      }
    }

    return classified;
  });
}

function classifyResult(result: MulticallResult, gasLimitThreshold: number): CallResultStatus {
  // Basic failure case
  if (!result.success) {
    const errorMsg = (result.errorMessage || '').toLowerCase();
    
    if (errorMsg.includes('out of gas') || errorMsg.includes('gas')) {
      return CallResultStatus.OUT_OF_GAS;
    }
    if (errorMsg.includes('revert')) {
      return CallResultStatus.REVERT;
    }
    if (!result.returnData || result.returnData === '0x') {
      return CallResultStatus.REVERT;
    }
    return CallResultStatus.UNKNOWN_ERROR;
  }

  // Success case - check for out of gas risk
  if (result.gasUsed && result.gasLimit) {
    const gasUsageRatio = result.gasUsed / result.gasLimit;
    if (gasUsageRatio >= gasLimitThreshold) {
      return CallResultStatus.OUT_OF_GAS;
    }
  }

  // Check for decode issues
  const cleanHex = result.returnData?.startsWith('0x') 
    ? result.returnData.slice(2) 
    : result.returnData || '';
  
  if (cleanHex.length < 64) { // less than 32 bytes
    return CallResultStatus.DECODE_ERROR;
  }

  return CallResultStatus.SUCCESS;
}

export function updateBlacklistFromResults(
  blacklist: PoolBlacklist,
  results: ClassifiedResult[]
): void {
  results.forEach(result => {
    const wasBlacklisted = blacklist.recordFailure(
      result.original.poolAddress,
      result.status
    );
    
    if (wasBlacklisted) {
      console.warn(`Pool ${result.original.poolAddress} blacklisted due to ${result.status}`);
    }
  });
}

export function filterResultsByBlacklist(
  results: ClassifiedResult[],
  blacklist: PoolBlacklist
): ClassifiedResult[] {
  return results.filter(result => {
    const { isBlacklisted } = blacklist.isBlacklisted(result.original.poolAddress);
    return !isBlacklisted;
  });
}

export function getSuccessfulResults(
  results: ClassifiedResult[],
  minAmountOut: string = '0'
): ClassifiedResult[] {
  const minAmount = BigInt(minAmountOut);
  
  return results.filter(result => 
    result.status === CallResultStatus.SUCCESS &&
    result.amountOut &&
    BigInt(result.amountOut) >= minAmount
  );
}

export interface MulticallStats {
  totalCalls: number;
  successfulCalls: number;
  outOfGasCalls: number;
  revertCalls: number;
  decodeErrorCalls: number;
  insufficientLiqCalls: number;
  unknownErrorCalls: number;
  successRate: number;
  avgGasUsed: number;
  blacklistedPools: number;
}

export function calculateStats(
  results: ClassifiedResult[],
  blacklist: PoolBlacklist
): MulticallStats {
  const statusCounts = new Map<CallResultStatus, number>();
  let totalGasUsed = 0;

  results.forEach(result => {
    const count = statusCounts.get(result.status) || 0;
    statusCounts.set(result.status, count + 1);
    totalGasUsed += result.original.gasUsed || 0;
  });

  const totalCalls = results.length;
  const successfulCalls = statusCounts.get(CallResultStatus.SUCCESS) || 0;

  return {
    totalCalls,
    successfulCalls,
    outOfGasCalls: statusCounts.get(CallResultStatus.OUT_OF_GAS) || 0,
    revertCalls: statusCounts.get(CallResultStatus.REVERT) || 0,
    decodeErrorCalls: statusCounts.get(CallResultStatus.DECODE_ERROR) || 0,
    insufficientLiqCalls: statusCounts.get(CallResultStatus.INSUFFICIENT_LIQUIDITY) || 0,
    unknownErrorCalls: statusCounts.get(CallResultStatus.UNKNOWN_ERROR) || 0,
    successRate: totalCalls > 0 ? successfulCalls / totalCalls : 0,
    avgGasUsed: totalCalls > 0 ? totalGasUsed / totalCalls : 0,
    blacklistedPools: blacklist.getBlacklistedPools().size
  };
}

// Example usage for smart-order-router integration
export function processMulticallBatch(
  results: MulticallResult[],
  blacklist: PoolBlacklist
): { filtered: ClassifiedResult[]; stats: MulticallStats } {
  // Classify results (consider calls using >90% of gas limit as OOG risk)
  const classified = classifyMulticallResults(results, 0.90);
  
  // Update blacklist based on failures
  updateBlacklistFromResults(blacklist, classified);
  
  // Filter out blacklisted pools for future batches
  const filtered = filterResultsByBlacklist(classified, blacklist);
  
  // Calculate statistics
  const stats = calculateStats(classified, blacklist);
  
  return { filtered, stats };
}

// Utility for integration with existing OnChainQuoteProvider
export function adaptToExistingResult(
  classified: ClassifiedResult[]
): Array<{
  success: boolean;
  result?: any;
  gasUsed?: number;
}> {
  return classified.map(c => ({
    success: c.status === CallResultStatus.SUCCESS,
    result: c.status === CallResultStatus.SUCCESS && c.amountOut ? 
      [BigInt(c.amountOut), [], [], BigInt(c.original.gasUsed || 0)] : 
      undefined,
    gasUsed: c.original.gasUsed
  }));
}
