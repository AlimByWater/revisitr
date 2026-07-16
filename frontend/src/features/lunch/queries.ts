import { useApiQuery } from '@/lib/swr'
import { lunchApi } from './api'

export function useLunchProgramQuery(botId: number) {
  return useApiQuery(botId ? `lunch-program-${botId}` : null, () => lunchApi.getProgram(botId))
}
