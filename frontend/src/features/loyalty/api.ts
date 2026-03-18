import { api } from '@/lib/api'
import type {
  LoyaltyProgram,
  LoyaltyLevel,
  CreateProgramRequest,
} from './types'

export async function listPrograms(): Promise<LoyaltyProgram[]> {
  const { data } = await api.get<LoyaltyProgram[]>('/loyalty/programs')
  return data
}

export async function getProgram(id: number): Promise<LoyaltyProgram> {
  const { data } = await api.get<LoyaltyProgram>(`/loyalty/programs/${id}`)
  return data
}

export async function createProgram(
  body: CreateProgramRequest,
): Promise<LoyaltyProgram> {
  const { data } = await api.post<LoyaltyProgram>('/loyalty/programs', body)
  return data
}

export async function updateProgram(
  id: number,
  body: Partial<Pick<LoyaltyProgram, 'name' | 'is_active' | 'config'>>,
): Promise<LoyaltyProgram> {
  const { data } = await api.patch<LoyaltyProgram>(
    `/loyalty/programs/${id}`,
    body,
  )
  return data
}

export async function createLevel(
  programId: number,
  body: Omit<LoyaltyLevel, 'id' | 'program_id'>,
): Promise<LoyaltyLevel> {
  const { data } = await api.post<LoyaltyLevel>(
    `/loyalty/programs/${programId}/levels`,
    body,
  )
  return data
}

export async function updateLevels(
  programId: number,
  levels: Omit<LoyaltyLevel, 'program_id'>[],
): Promise<LoyaltyLevel[]> {
  const { data } = await api.put<LoyaltyLevel[]>(
    `/loyalty/programs/${programId}/levels`,
    { levels },
  )
  return data
}

export async function deleteLevel(
  programId: number,
  levelId: number,
): Promise<void> {
  await api.delete(`/loyalty/programs/${programId}/levels/${levelId}`)
}
