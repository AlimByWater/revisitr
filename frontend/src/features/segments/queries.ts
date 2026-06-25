import { useApiQuery, useApiMutation } from '@/lib/swr'
import { segmentsApi } from './api'
import type { CreateSegmentRequest, UpdateSegmentRequest } from './types'

export function useSegmentsQuery() {
  return useApiQuery('segments', segmentsApi.list)
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
