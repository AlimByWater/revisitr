import { api } from '@/lib/api'
import type {
  Menu, CreateMenuRequest, MenuCategory, CreateMenuCategoryRequest,
  MenuItem, CreateMenuItemRequest, UpdateMenuCategoryRequest, UpdateMenuItemRequest,
  UpdateMenuRequest,
  ClientOrderStats,
} from './types'

export const menusApi = {
  list: async (): Promise<Menu[]> => {
    const response = await api.get<Menu[]>('/menus')
    return response.data
  },

  getById: async (id: number): Promise<Menu> => {
    const response = await api.get<Menu>(`/menus/${id}`)
    return response.data
  },

  create: async (data: CreateMenuRequest): Promise<Menu> => {
    const response = await api.post<Menu>('/menus', data)
    return response.data
  },

  update: async (id: number, data: UpdateMenuRequest): Promise<void> => {
    await api.patch(`/menus/${id}`, data)
  },

  remove: async (id: number): Promise<void> => {
    await api.delete(`/menus/${id}`)
  },

  addCategory: async (menuId: number, data: CreateMenuCategoryRequest): Promise<MenuCategory> => {
    const response = await api.post<MenuCategory>(`/menus/${menuId}/categories`, data)
    return response.data
  },

  updateCategory: async (categoryId: number, data: UpdateMenuCategoryRequest): Promise<MenuCategory> => {
    const response = await api.patch<MenuCategory>(`/menus/categories/${categoryId}`, data)
    return response.data
  },

  addItem: async (menuId: number, categoryId: number, data: CreateMenuItemRequest): Promise<MenuItem> => {
    const response = await api.post<MenuItem>(`/menus/${menuId}/categories/${categoryId}/items`, data)
    return response.data
  },

  updateItem: async (itemId: number, data: UpdateMenuItemRequest): Promise<MenuItem> => {
    const response = await api.patch<MenuItem>(`/menus/items/${itemId}`, data)
    return response.data
  },

  getClientOrderStats: async (clientId: number): Promise<ClientOrderStats> => {
    const response = await api.get<ClientOrderStats>(`/clients/${clientId}/order-stats`)
    return response.data
  },

  getBotPOSLocations: async (botId: number): Promise<{ pos_ids: number[] }> => {
    const response = await api.get<{ pos_ids: number[] }>(`/bots/${botId}/pos-locations`)
    return response.data
  },

  setBotPOSLocations: async (botId: number, posIds: number[]): Promise<void> => {
    await api.put(`/bots/${botId}/pos-locations`, { pos_ids: posIds })
  },
}
