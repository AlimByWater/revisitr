import { afterEach, describe, expect, it } from 'vitest'

import { authLoader } from './router'

describe('authLoader', () => {
  afterEach(() => {
    localStorage.clear()
  })

  it('redirects guests to login page', () => {
    const result = authLoader()

    expect(result).toBeInstanceOf(Response)
    expect((result as Response).status).toBe(302)
    expect((result as Response).headers.get('Location')).toBe('/auth/login')
  })

  it('allows authenticated users into dashboard routes', () => {
    localStorage.setItem('token', 'access-token')

    expect(authLoader()).toBeNull()
  })
})
