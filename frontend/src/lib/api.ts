import axios from 'axios'

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || '/api/v1',
  timeout: 10000,
})

let isRefreshing = false
let failedQueue: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

function processQueue(error: unknown, token: string | null) {
  failedQueue.forEach((promise) => {
    if (token) {
      promise.resolve(token)
    } else {
      promise.reject(error)
    }
  })
  failedQueue = []
}

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status !== 401 || originalRequest._retry) {
      return Promise.reject(error)
    }

    // Don't try to refresh if the failing request is the refresh endpoint itself
    if (originalRequest.url === '/auth/refresh') {
      localStorage.removeItem('token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/auth/login'
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise<string>((resolve, reject) => {
        failedQueue.push({ resolve, reject })
      }).then((token) => {
        originalRequest.headers.Authorization = `Bearer ${token}`
        return api(originalRequest)
      })
    }

    originalRequest._retry = true
    isRefreshing = true

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      localStorage.removeItem('token')
      window.location.href = '/auth/login'
      isRefreshing = false
      return Promise.reject(error)
    }

    try {
      const response = await axios.post(
        `${api.defaults.baseURL}/auth/refresh`,
        { refresh_token: refreshToken },
      )
      const { access_token, refresh_token } = response.data
      localStorage.setItem('token', access_token)
      localStorage.setItem('refresh_token', refresh_token)
      processQueue(null, access_token)
      originalRequest.headers.Authorization = `Bearer ${access_token}`
      return api(originalRequest)
    } catch (refreshError) {
      processQueue(refreshError, null)
      localStorage.removeItem('token')
      localStorage.removeItem('refresh_token')
      window.location.href = '/auth/login'
      return Promise.reject(refreshError)
    } finally {
      isRefreshing = false
    }
  },
)
