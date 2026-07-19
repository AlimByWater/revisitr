export interface PluginKey {
  id: number
  org_id: number
  integration_id: number
  label: string
  last_used_at?: string | null
  revoked_at?: string | null
  created_at: string
}

// The raw key is returned only once, on creation.
export interface CreateKeyResponse {
  id: number
  key: string
  label: string
}
