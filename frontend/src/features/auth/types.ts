export interface User {
  id: number
  email: string
  name: string
  role: string
  org_id: number
  phone?: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface AuthResponse {
  user: User
  tokens: TokenPair
}

export interface LoginRequest {
  email: string
  password: string
}

export interface RegisterRequest {
  email: string
  password: string
  name: string
  organization: string
  phone?: string
}
