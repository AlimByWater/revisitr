import { useApiQuery, useApiMutation } from '@/lib/swr'
import { rfmApi } from './api'
import type { UpdateRFMConfigRequest, SetTemplateRequest } from './types'

export function useRFMDashboardQuery() {
  return useApiQuery('rfm-dashboard', rfmApi.getDashboard)
}

export function useRFMConfigQuery() {
  return useApiQuery('rfm-config', rfmApi.getConfig)
}

export function useRFMRecalculateMutation() {
  return useApiMutation<void, void>(
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

export function useRFMTemplatesQuery() {
  return useApiQuery('rfm-templates', rfmApi.listTemplates)
}

export function useRFMActiveTemplateQuery() {
  return useApiQuery('rfm-active-template', rfmApi.getActiveTemplate)
}

export function useRFMSetTemplateMutation() {
  return useApiMutation(
    'rfm/set-template',
    (data: SetTemplateRequest) => rfmApi.setTemplate(data),
    ['rfm-active-template', 'rfm-dashboard', 'rfm-config'],
  )
}

export function useRFMOnboardingQuestionsQuery() {
  return useApiQuery('rfm-onboarding-questions', rfmApi.getOnboardingQuestions)
}

export function useRFMRecommendMutation() {
  return useApiMutation(
    'rfm/recommend',
    (answers: number[]) => rfmApi.recommendTemplate(answers),
  )
}

export function useRFMSegmentClientsQuery(
  segment: string | undefined,
  params: { page?: number; per_page?: number; sort?: string; order?: string } = {},
) {
  const key = segment
    ? `rfm-segment-clients-${segment}-${params.page ?? 1}-${params.sort ?? 'monetary_sum'}-${params.order ?? 'desc'}`
    : null
  return useApiQuery(key, () => rfmApi.getSegmentClients(segment!, params))
}
