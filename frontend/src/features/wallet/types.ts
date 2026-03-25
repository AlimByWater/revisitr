export interface WalletConfig {
  id: number
  org_id: number
  platform: 'apple' | 'google'
  is_enabled: boolean
  design: WalletDesign
  created_at: string
  updated_at: string
}

export interface WalletDesign {
  logo_url?: string
  background_color?: string
  foreground_color?: string
  label_color?: string
  description?: string
}

export interface WalletCredentials {
  // Apple
  pass_type_id?: string
  team_id?: string
  certificate?: string
  // Google
  issuer_id?: string
  service_account_key?: string
}

export interface SaveWalletConfigRequest {
  platform: 'apple' | 'google'
  is_enabled: boolean
  credentials: WalletCredentials
  design: WalletDesign
}

export interface WalletPass {
  id: number
  org_id: number
  client_id: number
  platform: 'apple' | 'google'
  serial_number: string
  last_balance: number
  last_level: string
  status: 'active' | 'suspended' | 'revoked'
  created_at: string
  updated_at: string
}

export interface IssueWalletPassRequest {
  client_id: number
  platform: 'apple' | 'google'
}

export interface WalletStats {
  total_passes: number
  apple_passes: number
  google_passes: number
  active_passes: number
}
