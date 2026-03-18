import { useApiQuery, useApiMutation } from '../../lib/swr'
import {
  listPrograms,
  getProgram,
  createProgram,
  updateProgram,
  updateLevels,
} from './api'
import type { CreateProgramRequest, LoyaltyLevel, LoyaltyProgram } from './types'

export function useProgramsQuery() {
  return useApiQuery('loyalty-programs', listPrograms)
}

export function useProgramQuery(id: number) {
  return useApiQuery(id ? `loyalty-programs-${id}` : null, () =>
    getProgram(id),
  )
}

export function useCreateProgramMutation() {
  return useApiMutation(
    'loyalty-programs/create',
    (data: CreateProgramRequest) => createProgram(data),
    ['loyalty-programs'],
  )
}

export function useUpdateProgramMutation() {
  return useApiMutation(
    'loyalty-programs/update',
    ({
      id,
      data,
    }: {
      id: number
      data: Partial<Pick<LoyaltyProgram, 'name' | 'is_active' | 'config'>>
    }) => updateProgram(id, data),
    ['loyalty-programs'],
  )
}

export function useUpdateLevelsMutation(programId: number) {
  return useApiMutation(
    `loyalty-programs/${programId}/levels`,
    (levels: Omit<LoyaltyLevel, 'program_id'>[]) =>
      updateLevels(programId, levels),
    [`loyalty-programs-${programId}`],
  )
}
