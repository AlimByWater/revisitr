import { useApiQuery, useApiMutation } from '@/lib/swr'
import { rfmApi } from './api'
import type { UpdateRFMConfigRequest } from './types'

export function useRFMDashboardQuery() {
  return useApiQuery('rfm-dashboard', rfmApi.getDashboard)
}

export function useRFMConfigQuery() {
  return useApiQuery('rfm-config', rfmApi.getConfig)
}

export function useRFMRecalculateMutation() {
  return useApiMutation(
    'rfm/recalculate',
    () => rfmApi.recalculate(),
    ['rfm-dashboard', 'rfm-config'],
  )
}

export function useRFMUpdateConfigMutation() {
  return useApiMutation(
    'rfm/update-config',
    (data: UpdateRFMConfigRequest) => rfmApi.updateConfig(data),
    ['rfm-config', 'rfm-dashboard'],
  )
}
