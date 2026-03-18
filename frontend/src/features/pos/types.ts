export interface DaySchedule {
  open: string
  close: string
  closed?: boolean
}

export type Schedule = Record<string, DaySchedule>

export interface POSLocation {
  id: number
  org_id: number
  name: string
  address: string
  phone: string
  schedule: Schedule
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface CreatePOSRequest {
  name: string
  address: string
  phone: string
  schedule: Schedule
}
