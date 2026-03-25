import { api } from '@/lib/api'
import type {
  Tariff,
  SubscriptionWithTariff,
  Invoice,
  CreateSubscriptionRequest,
  ChangeSubscriptionRequest,
} from './types'

export const billingApi = {
  getTariffs: async (): Promise<Tariff[]> => {
    const response = await api.get<Tariff[]>('/billing/tariffs')
    return response.data
  },

  getSubscription: async (): Promise<SubscriptionWithTariff> => {
    const response = await api.get<SubscriptionWithTariff>('/billing/subscription')
    return response.data
  },

  subscribe: async (data: CreateSubscriptionRequest): Promise<void> => {
    await api.post('/billing/subscription', data)
  },

  changePlan: async (data: ChangeSubscriptionRequest): Promise<void> => {
    await api.patch('/billing/subscription', data)
  },

  cancelSubscription: async (): Promise<void> => {
    await api.delete('/billing/subscription')
  },

  getInvoices: async (): Promise<Invoice[]> => {
    const response = await api.get<Invoice[]>('/billing/invoices')
    return response.data
  },

  getInvoice: async (id: number): Promise<Invoice> => {
    const response = await api.get<Invoice>(`/billing/invoices/${id}`)
    return response.data
  },
}
