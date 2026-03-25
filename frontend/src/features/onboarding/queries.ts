import { useApiQuery, useApiMutation } from '@/lib/swr'
import { onboardingApi } from './api'
import type { UpdateOnboardingRequest } from './types'

export function useOnboardingQuery() {
  return useApiQuery('onboarding', onboardingApi.getState)
}

export function useUpdateOnboardingMutation() {
  return useApiMutation(
    'onboarding/update',
    (data: UpdateOnboardingRequest) => onboardingApi.updateStep(data),
    ['onboarding'],
  )
}

export function useCompleteOnboardingMutation() {
  return useApiMutation(
    'onboarding/complete',
    () => onboardingApi.complete(),
    ['onboarding'],
  )
}

export function useResetOnboardingMutation() {
  return useApiMutation(
    'onboarding/reset',
    () => onboardingApi.reset(),
    ['onboarding'],
  )
}
