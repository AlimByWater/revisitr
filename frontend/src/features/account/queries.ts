import { useApiQuery, useApiMutation } from '@/lib/swr'
import { accountApi } from './api'
import type {
  UpdateProfileRequest,
  ChangeEmailRequest,
  ChangePhoneRequest,
  ChangePasswordRequest,
  BillingDetails,
} from './types'
import type { User } from '@/features/auth/types'

export function useProfileQuery() {
  return useApiQuery<User>('account/profile', accountApi.getProfile)
}

export function useUpdateProfileMutation() {
  return useApiMutation<User, UpdateProfileRequest>(
    'account/profile/update',
    accountApi.updateProfile,
    ['account/profile'],
  )
}

export function useChangeEmailMutation() {
  return useApiMutation<void, ChangeEmailRequest>(
    'account/change-email',
    accountApi.changeEmail,
  )
}

export function useChangePhoneMutation() {
  return useApiMutation<void, ChangePhoneRequest>(
    'account/change-phone',
    accountApi.changePhone,
  )
}

export function useChangePasswordMutation() {
  return useApiMutation<void, ChangePasswordRequest>(
    'account/change-password',
    accountApi.changePassword,
  )
}

export function useBillingDetailsQuery() {
  return useApiQuery<BillingDetails>('account/billing-details', accountApi.getBillingDetails)
}

export function useUpdateBillingDetailsMutation() {
  return useApiMutation<BillingDetails, BillingDetails>(
    'account/billing-details/update',
    accountApi.updateBillingDetails,
    ['account/billing-details'],
  )
}
