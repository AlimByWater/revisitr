import { api } from '@/lib/api'
import type {
  Promotion,
  PromoCode,
  PromoCodeValidation,
  PromoResult,
  PromoChannelAnalytics,
  CreatePromotionRequest,
  UpdatePromotionRequest,
  CreatePromoCodeRequest,
} from './types'

export const promotionsApi = {
  // Promotions
  list: async (): Promise<Promotion[]> => {
    const response = await api.get<Promotion[]>('/promotions')
    return response.data
  },

  getById: async (id: number): Promise<Promotion> => {
    const response = await api.get<Promotion>(`/promotions/${id}`)
    return response.data
  },

  create: async (data: CreatePromotionRequest): Promise<Promotion> => {
    const response = await api.post<Promotion>('/promotions', data)
    return response.data
  },

  update: async (
    id: number,
    data: UpdatePromotionRequest,
  ): Promise<Promotion> => {
    const response = await api.patch<Promotion>(`/promotions/${id}`, data)
    return response.data
  },

  remove: async (id: number): Promise<void> => {
    await api.delete(`/promotions/${id}`)
  },

  getPromotionCodes: async (promotionId: number): Promise<PromoCode[]> => {
    const response = await api.get<PromoCode[]>(
      `/promotions/${promotionId}/codes`,
    )
    return response.data
  },

  // Promo Codes
  listCodes: async (): Promise<PromoCode[]> => {
    const response = await api.get<PromoCode[]>('/promotions/promo-codes')
    return response.data
  },

  createCode: async (data: CreatePromoCodeRequest): Promise<PromoCode> => {
    const response = await api.post<PromoCode>('/promotions/promo-codes', data)
    return response.data
  },

  deactivateCode: async (id: number): Promise<void> => {
    await api.delete(`/promotions/promo-codes/${id}`)
  },

  validateCode: async (
    code: string,
    clientId: number,
    orderAmount?: number,
  ): Promise<PromoCodeValidation> => {
    const response = await api.post<PromoCodeValidation>(
      '/promotions/promo-codes/validate',
      { code, client_id: clientId, order_amount: orderAmount },
    )
    return response.data
  },

  activateCode: async (
    code: string,
    clientId: number,
  ): Promise<PromoResult> => {
    const response = await api.post<PromoResult>(
      '/promotions/promo-codes/activate',
      { code, client_id: clientId },
    )
    return response.data
  },

  generateCode: async (): Promise<string> => {
    const response = await api.get<{ code: string }>(
      '/promotions/promo-codes/generate',
    )
    return response.data.code
  },

  getChannelAnalytics: async (): Promise<PromoChannelAnalytics[]> => {
    const response = await api.get<PromoChannelAnalytics[]>(
      '/promotions/promo-codes/analytics',
    )
    return response.data
  },
}
