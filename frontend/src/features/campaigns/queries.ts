import { useApiQuery, useApiMutation } from '../../lib/swr'
import { campaignsApi } from './api'
import type {
  CreateCampaignRequest,
  UpdateCampaignRequest,
  CreateScenarioRequest,
  UpdateScenarioRequest,
  AudienceFilter,
} from './types'

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
    (filter: AudienceFilter) => campaignsApi.previewAudience(filter),
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

export function useScheduleCampaignMutation() {
  return useApiMutation(
    'campaigns/schedule',
    ({ id, scheduledAt }: { id: number; scheduledAt: string }) =>
      campaignsApi.schedule(id, scheduledAt),
    ['campaigns'],
  )
}

export function useCancelScheduleMutation() {
  return useApiMutation(
    'campaigns/cancel-schedule',
    (id: number) => campaignsApi.cancelSchedule(id),
    ['campaigns'],
  )
}

export function useCampaignAnalyticsQuery(id: number) {
  return useApiQuery(id ? `campaigns-${id}-analytics` : null, () =>
    campaignsApi.getAnalytics(id),
  )
}

export function useTemplatesQuery() {
  return useApiQuery('scenario-templates', campaignsApi.getTemplates)
}

export function useCloneTemplateMutation() {
  return useApiMutation(
    'scenarios/clone-template',
    ({ key, botId }: { key: string; botId: number }) =>
      campaignsApi.cloneTemplate(key, botId),
    ['scenarios'],
  )
}

export function useActionLogQuery(scenarioId: number, limit = 20, offset = 0) {
  return useApiQuery(
    scenarioId ? `scenarios-${scenarioId}-log-${limit}-${offset}` : null,
    () => campaignsApi.getActionLog(scenarioId, limit, offset),
  )
}
