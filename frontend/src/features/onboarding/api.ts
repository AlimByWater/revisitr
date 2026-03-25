import { api } from '@/lib/api'
import type { OnboardingResponse, OnboardingState, UpdateOnboardingRequest } from './types'

export const onboardingApi = {
  getState: async (): Promise<OnboardingResponse> => {
    const response = await api.get<OnboardingResponse>('/onboarding')
    return response.data
  },

  updateStep: async (data: UpdateOnboardingRequest): Promise<OnboardingState> => {
    const response = await api.patch<OnboardingState>('/onboarding', data)
    return response.data
  },

  complete: async (): Promise<void> => {
    await api.post('/onboarding/complete')
  },

  reset: async (): Promise<void> => {
    await api.post('/onboarding/reset')
  },
}
