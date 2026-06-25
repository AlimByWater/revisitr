import { useApiQuery, useApiMutation } from '../../lib/swr'
import { walletApi } from './api'
import type { SaveWalletConfigRequest } from './types'

export function useWalletConfigsQuery() {
  return useApiQuery('wallet-configs', walletApi.getConfigs)
}

export function useSaveWalletConfigMutation() {
  return useApiMutation(
    'wallet/save-config',
    ({ platform, data }: { platform: string; data: SaveWalletConfigRequest }) =>
      walletApi.saveConfig(platform, data),
    ['wallet-configs'],
  )
}

export function useWalletPassesQuery() {
  return useApiQuery('wallet-passes', walletApi.getPasses)
}

export function useRevokeWalletPassMutation() {
  return useApiMutation(
    'wallet/revoke-pass',
    (id: number) => walletApi.revokePass(id),
    ['wallet-passes', 'wallet-stats'],
  )
}

export function useWalletStatsQuery() {
  return useApiQuery('wallet-stats', walletApi.getStats)
}
