import { api } from '@/lib/api'
import type { AuthResponse, LoginRequest, RegisterRequest, TokenPair } from './types'

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await api.post<AuthResponse>('/auth/login', data)
    return response.data
  },

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await api.post<AuthResponse>('/auth/register', data)
    return response.data
  },

  refresh: async (refreshToken: string): Promise<TokenPair> => {
    const response = await api.post<TokenPair>('/auth/refresh', {
      refresh_token: refreshToken,
    })
    return response.data
  },

  logout: async (refreshToken: string): Promise<void> => {
    await api.post('/auth/logout', { refresh_token: refreshToken })
  },
}
