import { create } from 'zustand'
import { authApi } from '@/features/auth/api'
import type { AuthResponse, LoginRequest, RegisterRequest, User } from '@/features/auth/types'

interface AuthState {
  user: User | null
  accessToken: string | null
  isAuthenticated: boolean

  login: (data: LoginRequest) => Promise<void>
  register: (data: RegisterRequest) => Promise<void>
  logout: () => Promise<void>
  setAuth: (response: AuthResponse) => void
  clearAuth: () => void
  refreshToken: () => Promise<boolean>
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: localStorage.getItem('token'),
  isAuthenticated: localStorage.getItem('token') !== null,

  setAuth: (response: AuthResponse) => {
    localStorage.setItem('token', response.tokens.access_token)
    localStorage.setItem('refresh_token', response.tokens.refresh_token)
    set({
      user: response.user,
      accessToken: response.tokens.access_token,
      isAuthenticated: true,
    })
  },

  clearAuth: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('refresh_token')
    set({
      user: null,
      accessToken: null,
      isAuthenticated: false,
    })
  },

  login: async (data: LoginRequest) => {
    const response = await authApi.login(data)
    get().setAuth(response)
  },

  register: async (data: RegisterRequest) => {
    const response = await authApi.register(data)
    get().setAuth(response)
  },

  logout: async () => {
    const refreshToken = localStorage.getItem('refresh_token')
    if (refreshToken) {
      try {
        await authApi.logout(refreshToken)
      } catch {
        // Ignore logout API errors, clear local state anyway
      }
    }
    get().clearAuth()
  },

  refreshToken: async () => {
    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) return false

    try {
      const tokens = await authApi.refresh(refreshToken)
      localStorage.setItem('token', tokens.access_token)
      localStorage.setItem('refresh_token', tokens.refresh_token)
      set({
        accessToken: tokens.access_token,
        isAuthenticated: true,
      })
      return true
    } catch {
      get().clearAuth()
      return false
    }
  },
}))
