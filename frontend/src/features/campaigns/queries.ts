import { useApiQuery, useApiMutation } from '../../lib/swr'
import { campaignsApi } from './api'
import type {
  CreateCampaignRequest,
  UpdateCampaignRequest,
  CreateScenarioRequest,
  UpdateScenarioRequest,
} from './types'
import type { SegmentFilter } from '@/features/segments/types'
import { segmentsApi } from '@/features/segments/api'

export function useCampaignsQuery(limit = 20, offset = 0) {
  return useApiQuery(`campaigns-${limit}-${offset}`, () =>
    campaignsApi.list(limit, offset),
  )
}

export function useCampaignQuery(id: number) {
  return useApiQuery(id ? `campaigns-${id}` : null, () =>
    campaignsApi.getById(id),
  )
}

export function useScenariosQuery() {
  return useApiQuery('scenarios', campaignsApi.listScenarios)
}

export function useCreateCampaignMutation() {
  return useApiMutation(
    'campaigns/create',
    (data: CreateCampaignRequest) => campaignsApi.create(data),
    ['campaigns'],
  )
}

export function useUpdateCampaignMutation() {
  return useApiMutation(
    'campaigns/update',
    ({ id, data }: { id: number; data: UpdateCampaignRequest }) =>
      campaignsApi.update(id, data),
    ['campaigns'],
  )
}

export function useDeleteCampaignMutation() {
  return useApiMutation(
    'campaigns/delete',
    (id: number) => campaignsApi.remove(id),
    ['campaigns'],
  )
}

export function useSendCampaignMutation() {
  return useApiMutation(
    'campaigns/send',
    (id: number) => campaignsApi.send(id),
    ['campaigns'],
  )
}

export function usePreviewAudienceMutation() {
  return useApiMutation(
    'campaigns/preview-audience',
    (filter: SegmentFilter) =>
      segmentsApi.previewCount(filter).then((r) => r.count),
  )
}

export function useCreateScenarioMutation() {
  return useApiMutation(
    'scenarios/create',
    (data: CreateScenarioRequest) => campaignsApi.createScenario(data),
    ['scenarios'],
  )
}

export function useUpdateScenarioMutation() {
  return useApiMutation(
    'scenarios/update',
    ({ id, data }: { id: number; data: UpdateScenarioRequest }) =>
      campaignsApi.updateScenario(id, data),
    ['scenarios'],
  )
}

export function useDeleteScenarioMutation() {
  return useApiMutation(
    'scenarios/delete',
    (id: number) => campaignsApi.deleteScenario(id),
    ['scenarios'],
  )
}

export function useActionLogQuery(scenarioId: number, limit = 20, offset = 0) {
  return useApiQuery(
    scenarioId ? `scenarios-${scenarioId}-log-${limit}-${offset}` : null,
    () => campaignsApi.getActionLog(scenarioId, limit, offset),
  )
}

// ── Scenario by ID (client-side filter) ─────────────────────────────────────

export function useScenarioQuery(id: number) {
  const { data, ...rest } = useScenariosQuery()
  const scenario = data?.find((s) => s.id === id)
  return { data: scenario, ...rest }
}

// ── File Upload ─────────────────────────────────────────────────────────────

export function useUploadFileMutation() {
  return useApiMutation(
    'files/upload',
    (file: File) => campaignsApi.uploadFile(file),
  )
}
