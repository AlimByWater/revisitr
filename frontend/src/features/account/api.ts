import { api } from '@/lib/api'
import type { User } from '@/features/auth/types'
import type {
  UpdateProfileRequest,
  ChangeEmailRequest,
  ChangePhoneRequest,
  ChangePasswordRequest,
  BillingDetails,
} from './types'

export const accountApi = {
  getProfile: async (): Promise<User> => {
    const response = await api.get<User>('/account/profile')
    return response.data
  },

  updateProfile: async (data: UpdateProfileRequest): Promise<User> => {
    const response = await api.put<User>('/account/profile', data)
    return response.data
  },

  changeEmail: async (data: ChangeEmailRequest): Promise<void> => {
    await api.post('/account/change-email', data)
  },

  changePhone: async (data: ChangePhoneRequest): Promise<void> => {
    await api.post('/account/change-phone', data)
  },

  changePassword: async (data: ChangePasswordRequest): Promise<void> => {
    await api.post('/account/change-password', data)
  },

  getBillingDetails: async (): Promise<BillingDetails> => {
    const response = await api.get<BillingDetails>('/account/billing-details')
    return response.data
  },

  updateBillingDetails: async (data: BillingDetails): Promise<BillingDetails> => {
    const response = await api.put<BillingDetails>('/account/billing-details', data)
    return response.data
  },
}
