import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { campaignsApi } from './api'
import type {
  CreateCampaignRequest,
  UpdateCampaignRequest,
  CreateScenarioRequest,
  UpdateScenarioRequest,
  AudienceFilter,
} from './types'

export function useCampaignsQuery(limit = 20, offset = 0) {
  return useQuery({
    queryKey: ['campaigns', limit, offset],
    queryFn: () => campaignsApi.list(limit, offset),
  })
}

export function useCampaignQuery(id: number) {
  return useQuery({
    queryKey: ['campaigns', id],
    queryFn: () => campaignsApi.getById(id),
    enabled: !!id,
  })
}

export function useScenariosQuery() {
  return useQuery({
    queryKey: ['scenarios'],
    queryFn: campaignsApi.listScenarios,
  })
}

export function useCreateCampaignMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateCampaignRequest) => campaignsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
    },
  })
}

export function useUpdateCampaignMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateCampaignRequest }) =>
      campaignsApi.update(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
    },
  })
}

export function useDeleteCampaignMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => campaignsApi.remove(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
    },
  })
}

export function useSendCampaignMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => campaignsApi.send(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['campaigns'] })
    },
  })
}

export function usePreviewAudienceMutation() {
  return useMutation({
    mutationFn: (filter: AudienceFilter) =>
      campaignsApi.previewAudience(filter),
  })
}

export function useCreateScenarioMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: CreateScenarioRequest) =>
      campaignsApi.createScenario(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scenarios'] })
    },
  })
}

export function useUpdateScenarioMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number
      data: UpdateScenarioRequest
    }) => campaignsApi.updateScenario(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scenarios'] })
    },
  })
}

export function useDeleteScenarioMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => campaignsApi.deleteScenario(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['scenarios'] })
    },
  })
}
