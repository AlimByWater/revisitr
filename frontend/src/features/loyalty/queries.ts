import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  listPrograms,
  getProgram,
  createProgram,
  updateProgram,
  updateLevels,
} from './api'
import type { CreateProgramRequest, LoyaltyLevel, LoyaltyProgram } from './types'

export function useProgramsQuery() {
  return useQuery({
    queryKey: ['loyalty-programs'],
    queryFn: listPrograms,
  })
}

export function useProgramQuery(id: number) {
  return useQuery({
    queryKey: ['loyalty-programs', id],
    queryFn: () => getProgram(id),
    enabled: !!id,
  })
}

export function useCreateProgramMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: CreateProgramRequest) => createProgram(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['loyalty-programs'] })
    },
  })
}

export function useUpdateProgramMutation() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: number
      data: Partial<Pick<LoyaltyProgram, 'name' | 'is_active' | 'config'>>
    }) => updateProgram(id, data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['loyalty-programs'] })
      queryClient.invalidateQueries({
        queryKey: ['loyalty-programs', variables.id],
      })
    },
  })
}

export function useUpdateLevelsMutation(programId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (levels: Omit<LoyaltyLevel, 'program_id'>[]) =>
      updateLevels(programId, levels),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['loyalty-programs', programId],
      })
    },
  })
}
