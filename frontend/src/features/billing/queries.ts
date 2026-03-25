import { useApiQuery, useApiMutation } from '@/lib/swr'
import { billingApi } from './api'
import type { CreateSubscriptionRequest, ChangeSubscriptionRequest } from './types'

export function useTariffsQuery() {
  return useApiQuery('billing-tariffs', billingApi.getTariffs)
}

export function useSubscriptionQuery() {
  return useApiQuery('billing-subscription', billingApi.getSubscription)
}

export function useInvoicesQuery() {
  return useApiQuery('billing-invoices', billingApi.getInvoices)
}

export function useSubscribeMutation() {
  return useApiMutation(
    'billing/subscribe',
    (data: CreateSubscriptionRequest) => billingApi.subscribe(data),
    ['billing-subscription', 'billing-invoices'],
  )
}

export function useChangePlanMutation() {
  return useApiMutation(
    'billing/change-plan',
    (data: ChangeSubscriptionRequest) => billingApi.changePlan(data),
    ['billing-subscription', 'billing-invoices'],
  )
}

export function useCancelSubscriptionMutation() {
  return useApiMutation(
    'billing/cancel',
    () => billingApi.cancelSubscription(),
    ['billing-subscription'],
  )
}
