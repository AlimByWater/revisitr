import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useAuthStore } from './auth'

// Mock the auth API module
vi.mock('@/features/auth/api', () => ({
  authApi: {
    login: vi.fn(),
    register: vi.fn(),
    logout: vi.fn(),
    refresh: vi.fn(),
  },
}))

import { authApi } from '@/features/auth/api'

const mockUser = {
  id: 1,
  email: 'test@example.com',
  name: 'Test User',
  role: 'admin',
  org_id: 10,
}

const mockTokens = {
  access_token: 'access-tok',
  refresh_token: 'refresh-tok',
  expires_in: 3600,
}

describe('useAuthStore', () => {
  beforeEach(() => {
    localStorage.clear()
    useAuthStore.setState({ user: null, accessToken: null, isAuthenticated: false })
    vi.clearAllMocks()
  })

  it('initial state: unauthenticated when no token in localStorage', () => {
    const { isAuthenticated, user, accessToken } = useAuthStore.getState()
    expect(isAuthenticated).toBe(false)
    expect(user).toBeNull()
    expect(accessToken).toBeNull()
  })

  it('setAuth stores tokens and user', () => {
    useAuthStore.getState().setAuth({ user: mockUser, tokens: mockTokens })

    const { user, accessToken, isAuthenticated } = useAuthStore.getState()
    expect(user).toEqual(mockUser)
    expect(accessToken).toBe('access-tok')
    expect(isAuthenticated).toBe(true)
    expect(localStorage.getItem('token')).toBe('access-tok')
    expect(localStorage.getItem('refresh_token')).toBe('refresh-tok')
  })

  it('clearAuth removes tokens and user', () => {
    useAuthStore.getState().setAuth({ user: mockUser, tokens: mockTokens })
    useAuthStore.getState().clearAuth()

    const { user, accessToken, isAuthenticated } = useAuthStore.getState()
    expect(user).toBeNull()
    expect(accessToken).toBeNull()
    expect(isAuthenticated).toBe(false)
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('login calls API and sets auth state', async () => {
    vi.mocked(authApi.login).mockResolvedValue({ user: mockUser, tokens: mockTokens })

    await useAuthStore.getState().login({ email: 'test@example.com', password: 'pw' })

    expect(authApi.login).toHaveBeenCalledWith({ email: 'test@example.com', password: 'pw' })
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
  })

  it('register calls API and sets auth state', async () => {
    vi.mocked(authApi.register).mockResolvedValue({ user: mockUser, tokens: mockTokens })

    await useAuthStore.getState().register({
      email: 'test@example.com',
      password: 'pw',
      name: 'Test',
      organization: 'Org',
    })

    expect(authApi.register).toHaveBeenCalled()
    expect(useAuthStore.getState().isAuthenticated).toBe(true)
  })

  it('logout calls API with refresh token and clears state', async () => {
    useAuthStore.getState().setAuth({ user: mockUser, tokens: mockTokens })
    vi.mocked(authApi.logout).mockResolvedValue(undefined)

    await useAuthStore.getState().logout()

    expect(authApi.logout).toHaveBeenCalledWith('refresh-tok')
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('logout clears state even if API throws', async () => {
    useAuthStore.getState().setAuth({ user: mockUser, tokens: mockTokens })
    vi.mocked(authApi.logout).mockRejectedValue(new Error('network'))

    await useAuthStore.getState().logout()

    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })

  it('refreshToken returns false when no token stored', async () => {
    const result = await useAuthStore.getState().refreshToken()
    expect(result).toBe(false)
    expect(authApi.refresh).not.toHaveBeenCalled()
  })

  it('refreshToken updates tokens on success', async () => {
    localStorage.setItem('refresh_token', 'old-refresh')
    vi.mocked(authApi.refresh).mockResolvedValue({
      access_token: 'new-access',
      refresh_token: 'new-refresh',
      expires_in: 3600,
    })

    const result = await useAuthStore.getState().refreshToken()

    expect(result).toBe(true)
    expect(localStorage.getItem('token')).toBe('new-access')
  })

  it('refreshToken clears auth on failure', async () => {
    localStorage.setItem('refresh_token', 'bad-token')
    useAuthStore.setState({ isAuthenticated: true })
    vi.mocked(authApi.refresh).mockRejectedValue(new Error('expired'))

    const result = await useAuthStore.getState().refreshToken()

    expect(result).toBe(false)
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
  })
})
