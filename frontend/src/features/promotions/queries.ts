import { useApiQuery, useApiMutation } from '@/lib/swr'
import { promotionsApi } from './api'
import type {
  CreatePromotionRequest,
  UpdatePromotionRequest,
  CreatePromoCodeRequest,
} from './types'

export function usePromotionsQuery() {
  return useApiQuery('promotions', promotionsApi.list)
}

export function usePromotionQuery(id: number) {
  return useApiQuery(id ? `promotions-${id}` : null, () =>
    promotionsApi.getById(id),
  )
}

export function usePromotionCodesQuery(promotionId: number) {
  return useApiQuery(
    promotionId ? `promotions-${promotionId}-codes` : null,
    () => promotionsApi.getPromotionCodes(promotionId),
  )
}

export function usePromoCodesQuery() {
  return useApiQuery('promo-codes', promotionsApi.listCodes)
}

export function useChannelAnalyticsQuery() {
  return useApiQuery('promo-channel-analytics', promotionsApi.getChannelAnalytics)
}

export function useCreatePromotionMutation() {
  return useApiMutation(
    'promotions/create',
    (data: CreatePromotionRequest) => promotionsApi.create(data),
    ['promotions'],
  )
}

export function useUpdatePromotionMutation() {
  return useApiMutation(
    'promotions/update',
    ({ id, data }: { id: number; data: UpdatePromotionRequest }) =>
      promotionsApi.update(id, data),
    ['promotions'],
  )
}

export function useDeletePromotionMutation() {
  return useApiMutation(
    'promotions/delete',
    (id: number) => promotionsApi.remove(id),
    ['promotions'],
  )
}

export function useCreatePromoCodeMutation() {
  return useApiMutation(
    'promo-codes/create',
    (data: CreatePromoCodeRequest) => promotionsApi.createCode(data),
    ['promo-codes'],
  )
}

export function useDeactivatePromoCodeMutation() {
  return useApiMutation(
    'promo-codes/deactivate',
    (id: number) => promotionsApi.deactivateCode(id),
    ['promo-codes'],
  )
}

export function useGenerateCodeMutation() {
  return useApiMutation(
    'promo-codes/generate',
    () => promotionsApi.generateCode(),
  )
}

export function useValidatePromoCodeMutation() {
  return useApiMutation(
    'promo-codes/validate',
    ({
      code,
      clientId,
      orderAmount,
    }: {
      code: string
      clientId: number
      orderAmount?: number
    }) => promotionsApi.validateCode(code, clientId, orderAmount),
  )
}

export function useActivatePromoCodeMutation() {
  return useApiMutation(
    'promo-codes/activate',
    ({ code, clientId }: { code: string; clientId: number }) =>
      promotionsApi.activateCode(code, clientId),
    ['promo-codes'],
  )
}
