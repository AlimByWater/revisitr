import { api } from '@/lib/api'
import type { POSLocation, CreatePOSRequest } from './types'

export async function list(): Promise<POSLocation[]> {
  const { data } = await api.get<POSLocation[]>('/pos')
  return data
}

export async function getById(id: number): Promise<POSLocation> {
  const { data } = await api.get<POSLocation>(`/pos/${id}`)
  return data
}

export async function create(body: CreatePOSRequest): Promise<POSLocation> {
  const { data } = await api.post<POSLocation>('/pos', body)
  return data
}

export async function update(
  id: number,
  body: Partial<POSLocation>,
): Promise<POSLocation> {
  const { data } = await api.patch<POSLocation>(`/pos/${id}`, body)
  return data
}

export async function remove(id: number): Promise<void> {
  await api.delete(`/pos/${id}`)
}
