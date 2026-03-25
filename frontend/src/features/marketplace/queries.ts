import { useApiQuery, useApiMutation } from '../../lib/swr'
import { marketplaceApi } from './api'
import type { CreateProductRequest, UpdateProductRequest, PlaceOrderRequest } from './types'

export function useMarketplaceProductsQuery() {
  return useApiQuery('marketplace-products', marketplaceApi.getProducts)
}

export function useMarketplaceProductQuery(id: number) {
  return useApiQuery(
    id ? `marketplace-product-${id}` : null,
    () => marketplaceApi.getProduct(id),
  )
}

export function useCreateProductMutation() {
  return useApiMutation(
    'marketplace/create-product',
    (data: CreateProductRequest) => marketplaceApi.createProduct(data),
    ['marketplace-products', 'marketplace-stats'],
  )
}

export function useUpdateProductMutation() {
  return useApiMutation(
    'marketplace/update-product',
    ({ id, data }: { id: number; data: UpdateProductRequest }) =>
      marketplaceApi.updateProduct(id, data),
    ['marketplace-products'],
  )
}

export function useDeleteProductMutation() {
  return useApiMutation(
    'marketplace/delete-product',
    (id: number) => marketplaceApi.deleteProduct(id),
    ['marketplace-products', 'marketplace-stats'],
  )
}

export function useMarketplaceOrdersQuery() {
  return useApiQuery('marketplace-orders', marketplaceApi.getOrders)
}

export function usePlaceOrderMutation() {
  return useApiMutation(
    'marketplace/place-order',
    (data: PlaceOrderRequest) => marketplaceApi.placeOrder(data),
    ['marketplace-orders', 'marketplace-stats'],
  )
}

export function useUpdateOrderStatusMutation() {
  return useApiMutation(
    'marketplace/update-order-status',
    ({ id, status }: { id: number; status: string }) =>
      marketplaceApi.updateOrderStatus(id, status),
    ['marketplace-orders'],
  )
}

export function useMarketplaceStatsQuery() {
  return useApiQuery('marketplace-stats', marketplaceApi.getStats)
}
