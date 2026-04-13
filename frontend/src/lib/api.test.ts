import axios from 'axios'
import type { AxiosRequestConfig, InternalAxiosRequestConfig } from 'axios'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { waitFor } from '@testing-library/react'

interface Deferred<T> {
  promise: Promise<T>
  resolve: (value: T) => void
  reject: (reason?: unknown) => void
}

function createDeferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void
  let reject!: (reason?: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

const originalLocation = window.location

function unauthorizedError(config: AxiosRequestConfig) {
  return Promise.reject({
    config,
    response: {
      status: 401,
      data: { error: 'unauthorized' },
      headers: {},
      statusText: 'Unauthorized',
      config,
    },
  })
}

function mockLocation(href = 'http://localhost/revisitr/dashboard') {
  Object.defineProperty(window, 'location', {
    configurable: true,
    value: {
      href,
    },
  })
}

describe('api interceptors', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    vi.resetModules()
    vi.stubEnv('VITE_MOCK_API', 'false')
    vi.stubEnv('BASE_URL', '/revisitr/')
    localStorage.clear()
    mockLocation()
  })

  afterEach(() => {
    vi.unstubAllEnvs()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: originalLocation,
    })
  })

  it('adds Authorization header from localStorage token', async () => {
    localStorage.setItem('token', 'access-123')
    const { api } = await import('./api')

    let seenAuthHeader: string | undefined
    api.defaults.adapter = async (config) => {
      seenAuthHeader = config.headers?.Authorization as string | undefined
      return {
        data: { ok: true },
        status: 200,
        statusText: 'OK',
        headers: {},
        config,
      }
    }

    const response = await api.get('/protected')

    expect(response.data).toEqual({ ok: true })
    expect(seenAuthHeader).toBe('Bearer access-123')
  })

  it('refreshes token after 401 and retries original request', async () => {
    localStorage.setItem('token', 'expired-token')
    localStorage.setItem('refresh_token', 'refresh-123')

    const refreshSpy = vi.spyOn(axios, 'post').mockResolvedValue({
      data: {
        access_token: 'new-access',
        refresh_token: 'new-refresh',
      },
    })

    const { api } = await import('./api')

    let requestAttempts = 0
    const authHeaders: string[] = []
    api.defaults.adapter = async (config) => {
      requestAttempts += 1
      authHeaders.push((config.headers?.Authorization as string | undefined) ?? '')

      if (requestAttempts === 1) {
        return unauthorizedError(config)
      }

      return {
        data: { ok: true },
        status: 200,
        statusText: 'OK',
        headers: {},
        config,
      }
    }

    const response = await api.get('/protected')

    expect(response.data).toEqual({ ok: true })
    expect(requestAttempts).toBe(2)
    expect(refreshSpy).toHaveBeenCalledWith('/api/v1/auth/refresh', {
      refresh_token: 'refresh-123',
    })
    expect(authHeaders).toEqual(['Bearer expired-token', 'Bearer new-access'])
    expect(localStorage.getItem('token')).toBe('new-access')
    expect(localStorage.getItem('refresh_token')).toBe('new-refresh')
  })

  it('queues concurrent 401 requests behind a single refresh', async () => {
    localStorage.setItem('token', 'expired-token')
    localStorage.setItem('refresh_token', 'refresh-xyz')

    const refreshDeferred = createDeferred<{ data: { access_token: string; refresh_token: string } }>()
    const refreshSpy = vi.spyOn(axios, 'post').mockReturnValue(refreshDeferred.promise as ReturnType<typeof axios.post>)

    const { api } = await import('./api')

    let protectedCalls = 0
    api.defaults.adapter = async (config) => {
      protectedCalls += 1
      const authHeader = config.headers?.Authorization as string | undefined

      if (authHeader === 'Bearer new-access') {
        return {
          data: { ok: true, url: config.url },
          status: 200,
          statusText: 'OK',
          headers: {},
          config,
        }
      }

      return unauthorizedError(config)
    }

    const first = api.get('/protected-a')
    const second = api.get('/protected-b')

    await waitFor(() => {
      expect(refreshSpy).toHaveBeenCalledTimes(1)
    })

    refreshDeferred.resolve({
      data: {
        access_token: 'new-access',
        refresh_token: 'new-refresh',
      },
    })

    const [firstResponse, secondResponse] = await Promise.all([first, second])

    expect(firstResponse.data).toEqual({ ok: true, url: '/protected-a' })
    expect(secondResponse.data).toEqual({ ok: true, url: '/protected-b' })
    expect(refreshSpy).toHaveBeenCalledTimes(1)
    expect(protectedCalls).toBe(4)
  })

  it('does not try to refresh auth endpoints', async () => {
    localStorage.setItem('token', 'expired-token')
    localStorage.setItem('refresh_token', 'refresh-123')

    const refreshSpy = vi.spyOn(axios, 'post')
    const { api } = await import('./api')

    api.defaults.adapter = async (config) => unauthorizedError(config)

    await expect(api.post('/auth/login', { email: 'a', password: 'b' })).rejects.toMatchObject({
      response: { status: 401 },
    })
    expect(refreshSpy).not.toHaveBeenCalled()
  })

  it('clears tokens and redirects when refresh token is missing', async () => {
    localStorage.setItem('token', 'expired-token')
    const { api } = await import('./api')

    api.defaults.adapter = async (config) => unauthorizedError(config)

    await expect(api.get('/protected')).rejects.toMatchObject({
      response: { status: 401 },
    })

    expect(localStorage.getItem('token')).toBeNull()
    expect(window.location.href).toBe('/revisitr/auth/login')
  })

  it('clears tokens and redirects when refresh request fails', async () => {
    localStorage.setItem('token', 'expired-token')
    localStorage.setItem('refresh_token', 'refresh-123')

    vi.spyOn(axios, 'post').mockRejectedValue(new Error('refresh failed'))
    const { api } = await import('./api')

    api.defaults.adapter = async (config) => unauthorizedError(config)

    await expect(api.get('/protected')).rejects.toThrow('refresh failed')

    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('refresh_token')).toBeNull()
    expect(window.location.href).toBe('/revisitr/auth/login')
  })
})
