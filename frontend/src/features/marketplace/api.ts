import { api } from '@/lib/api'
import type {
  MarketplaceProduct,
  CreateProductRequest,
  UpdateProductRequest,
  MarketplaceOrder,
  PlaceOrderRequest,
  MarketplaceStats,
} from './types'

export const marketplaceApi = {
  getProducts: async (): Promise<MarketplaceProduct[]> => {
    const response = await api.get<MarketplaceProduct[]>('/marketplace/products')
    return response.data
  },

  getProduct: async (id: number): Promise<MarketplaceProduct> => {
    const response = await api.get<MarketplaceProduct>(`/marketplace/products/${id}`)
    return response.data
  },

  createProduct: async (data: CreateProductRequest): Promise<MarketplaceProduct> => {
    const response = await api.post<MarketplaceProduct>('/marketplace/products', data)
    return response.data
  },

  updateProduct: async (id: number, data: UpdateProductRequest): Promise<MarketplaceProduct> => {
    const response = await api.patch<MarketplaceProduct>(`/marketplace/products/${id}`, data)
    return response.data
  },

  deleteProduct: async (id: number): Promise<void> => {
    await api.delete(`/marketplace/products/${id}`)
  },

  getOrders: async (): Promise<MarketplaceOrder[]> => {
    const response = await api.get<MarketplaceOrder[]>('/marketplace/orders')
    return response.data
  },

  getOrder: async (id: number): Promise<MarketplaceOrder> => {
    const response = await api.get<MarketplaceOrder>(`/marketplace/orders/${id}`)
    return response.data
  },

  placeOrder: async (data: PlaceOrderRequest): Promise<MarketplaceOrder> => {
    const response = await api.post<MarketplaceOrder>('/marketplace/orders', data)
    return response.data
  },

  updateOrderStatus: async (id: number, status: string): Promise<void> => {
    await api.patch(`/marketplace/orders/${id}/status`, { status })
  },

  getStats: async (): Promise<MarketplaceStats> => {
    const response = await api.get<MarketplaceStats>('/marketplace/stats')
    return response.data
  },
}
