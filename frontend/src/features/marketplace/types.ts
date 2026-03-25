export interface MarketplaceProduct {
  id: number
  org_id: number
  name: string
  description: string
  image_url: string
  price_points: number
  stock: number | null
  is_active: boolean
  sort_order: number
  created_at: string
  updated_at: string
}

export interface CreateProductRequest {
  name: string
  description?: string
  image_url?: string
  price_points: number
  stock?: number | null
  sort_order?: number
}

export interface UpdateProductRequest {
  name?: string
  description?: string
  image_url?: string
  price_points?: number
  stock?: number | null
  is_active?: boolean
  sort_order?: number
}

export interface MarketplaceOrderItem {
  product_id: number
  product_name: string
  quantity: number
  points: number
}

export interface MarketplaceOrder {
  id: number
  org_id: number
  client_id: number
  status: 'pending' | 'confirmed' | 'completed' | 'cancelled'
  total_points: number
  items: MarketplaceOrderItem[]
  note: string
  created_at: string
  updated_at: string
}

export interface PlaceOrderRequest {
  client_id: number
  items: { product_id: number; quantity: number }[]
  note?: string
}

export interface MarketplaceStats {
  total_products: number
  active_products: number
  total_orders: number
  total_spent_points: number
}
