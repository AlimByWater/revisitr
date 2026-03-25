import { useApiQuery, useApiMutation } from '@/lib/swr'
import { segmentsApi } from './api'
import type { CreateSegmentRequest, UpdateSegmentRequest, CreateSegmentRuleRequest } from './types'

export function useSegmentsQuery() {
  return useApiQuery('segments', segmentsApi.list)
}

export function useSegmentQuery(id: number) {
  return useApiQuery(
    id ? `segment-${id}` : null,
    () => segmentsApi.getById(id),
  )
}

export function useCreateSegmentMutation() {
  return useApiMutation(
    'segments/create',
    (data: CreateSegmentRequest) => segmentsApi.create(data),
    ['segments'],
  )
}

export function useUpdateSegmentMutation() {
  return useApiMutation(
    'segments/update',
    ({ id, data }: { id: number; data: UpdateSegmentRequest }) =>
      segmentsApi.update(id, data),
    ['segments'],
  )
}

export function useDeleteSegmentMutation() {
  return useApiMutation(
    'segments/delete',
    (id: number) => segmentsApi.delete(id),
    ['segments'],
  )
}

export function useSegmentClientsQuery(segmentId: number, limit = 20, offset = 0) {
  return useApiQuery(
    segmentId ? `segment-${segmentId}-clients-${limit}-${offset}` : null,
    () => segmentsApi.getClients(segmentId, limit, offset),
  )
}

export function useRecalculateSegmentMutation() {
  return useApiMutation(
    'segments/recalculate',
    (segmentId: number) => segmentsApi.recalculate(segmentId),
    ['segments'],
  )
}

// Rules
export function useSegmentRulesQuery(segmentId: number) {
  return useApiQuery(
    segmentId ? `segment-${segmentId}-rules` : null,
    () => segmentsApi.getRules(segmentId),
  )
}

export function useAddSegmentRuleMutation(segmentId: number) {
  return useApiMutation(
    `segments/${segmentId}/add-rule`,
    (data: CreateSegmentRuleRequest) => segmentsApi.addRule(segmentId, data),
    [`segment-${segmentId}-rules`, 'segments'],
  )
}

export function useDeleteSegmentRuleMutation(segmentId: number) {
  return useApiMutation(
    `segments/${segmentId}/delete-rule`,
    (ruleId: number) => segmentsApi.deleteRule(segmentId, ruleId),
    [`segment-${segmentId}-rules`],
  )
}

// Predictions
export function usePredictionsQuery(limit = 20, offset = 0) {
  return useApiQuery(
    `predictions-${limit}-${offset}`,
    () => segmentsApi.getPredictions(limit, offset),
  )
}

export function usePredictionSummaryQuery() {
  return useApiQuery('prediction-summary', segmentsApi.getPredictionSummary)
}

export function useHighChurnQuery(threshold = 0.7) {
  return useApiQuery(
    `high-churn-${threshold}`,
    () => segmentsApi.getHighChurnClients(threshold),
  )
}
