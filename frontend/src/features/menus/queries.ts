import { useApiQuery, useApiMutation } from '@/lib/swr'
import { menusApi } from './api'
import type { CreateMenuRequest, CreateMenuCategoryRequest, CreateMenuItemRequest, UpdateMenuItemRequest } from './types'

export function useMenusQuery() {
  return useApiQuery('menus', menusApi.list)
}

export function useMenuQuery(id: number) {
  return useApiQuery(id ? `menus-${id}` : null, () => menusApi.getById(id))
}

export function useCreateMenuMutation() {
  return useApiMutation(
    'menus/create',
    (data: CreateMenuRequest) => menusApi.create(data),
    ['menus'],
  )
}

export function useDeleteMenuMutation() {
  return useApiMutation(
    'menus/delete',
    (id: number) => menusApi.remove(id),
    ['menus'],
  )
}

export function useAddCategoryMutation(menuId: number) {
  return useApiMutation(
    `menus/${menuId}/add-category`,
    (data: CreateMenuCategoryRequest) => menusApi.addCategory(menuId, data),
    [`menus-${menuId}`, 'menus'],
  )
}

export function useAddItemMutation(menuId: number, categoryId: number) {
  return useApiMutation(
    `menus/${menuId}/categories/${categoryId}/add-item`,
    (data: CreateMenuItemRequest) => menusApi.addItem(menuId, categoryId, data),
    [`menus-${menuId}`],
  )
}

export function useUpdateItemMutation() {
  return useApiMutation(
    'menus/update-item',
    ({ itemId, data }: { itemId: number; data: UpdateMenuItemRequest }) => menusApi.updateItem(itemId, data),
    ['menus'],
  )
}

export function useClientOrderStatsQuery(clientId: number) {
  return useApiQuery(
    clientId ? `client-order-stats-${clientId}` : null,
    () => menusApi.getClientOrderStats(clientId),
  )
}
