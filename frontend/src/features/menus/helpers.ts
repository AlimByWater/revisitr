import type { Menu } from './types'

export interface MenuConflict {
  posId: number
  posName: string
  menus: { id: number; name: string; bindingCreatedAt?: string }[]
  fallbackMenuName: string
}

export function findMenuBindingConflicts(menus: Menu[]): MenuConflict[] {
  const byPos = new Map<
    number,
    { posName: string; menuId: number; menuName: string; bindingCreatedAt?: string }[]
  >()

  for (const menu of menus) {
    for (const binding of menu.bindings ?? []) {
      if (!binding.is_active) continue

      const current = byPos.get(binding.pos_id) ?? []
      current.push({
        posName: binding.pos_name ?? `Точка продаж #${binding.pos_id}`,
        menuId: menu.id,
        menuName: menu.name,
        bindingCreatedAt: binding.created_at,
      })
      byPos.set(binding.pos_id, current)
    }
  }

  return Array.from(byPos.entries())
    .filter(([, bindings]) => bindings.length > 1)
    .map(([posId, bindings]) => {
      const sorted = [...bindings].sort((left, right) => {
        const leftCreatedAt = left.bindingCreatedAt ?? ''
        const rightCreatedAt = right.bindingCreatedAt ?? ''

        if (leftCreatedAt !== rightCreatedAt) {
          return leftCreatedAt.localeCompare(rightCreatedAt)
        }

        return left.menuId - right.menuId
      })

      return {
        posId,
        posName: sorted[0]?.posName ?? `Точка продаж #${posId}`,
        menus: sorted.map((binding) => ({
          id: binding.menuId,
          name: binding.menuName,
          bindingCreatedAt: binding.bindingCreatedAt,
        })),
        fallbackMenuName: sorted[0]?.menuName ?? 'Неизвестно',
      }
    })
}
