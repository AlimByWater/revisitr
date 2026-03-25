import { useApiQuery, useApiMutation } from '../../lib/swr'
import { walletApi } from './api'
import type { SaveWalletConfigRequest, IssueWalletPassRequest } from './types'

export function useWalletConfigsQuery() {
  return useApiQuery('wallet-configs', walletApi.getConfigs)
}

export function useWalletConfigQuery(platform: string) {
  return useApiQuery(
    platform ? `wallet-config-${platform}` : null,
    () => walletApi.getConfig(platform),
  )
}

export function useSaveWalletConfigMutation() {
  return useApiMutation(
    'wallet/save-config',
    ({ platform, data }: { platform: string; data: SaveWalletConfigRequest }) =>
      walletApi.saveConfig(platform, data),
    ['wallet-configs'],
  )
}

export function useDeleteWalletConfigMutation() {
  return useApiMutation(
    'wallet/delete-config',
    (platform: string) => walletApi.deleteConfig(platform),
    ['wallet-configs'],
  )
}

export function useWalletPassesQuery() {
  return useApiQuery('wallet-passes', walletApi.getPasses)
}

export function useIssueWalletPassMutation() {
  return useApiMutation(
    'wallet/issue-pass',
    (data: IssueWalletPassRequest) => walletApi.issuePass(data),
    ['wallet-passes', 'wallet-stats'],
  )
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
