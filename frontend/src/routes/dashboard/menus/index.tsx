import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  useMenusQuery,
  useCreateMenuMutation,
  useDeleteMenuMutation,
  useUpdateMenuMutation,
} from '@/features/menus/queries'
import { findMenuBindingConflicts } from '@/features/menus/helpers'
import { usePOSQuery } from '@/features/pos/queries'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { AlertTriangle, ArrowLeft, ChevronRight, Plus, Trash2, UtensilsCrossed } from 'lucide-react'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export default function MenusPage() {
  const [searchParams] = useSearchParams()
  const botId = searchParams.get('botId')
  const { data: menus, isLoading, isError, mutate } = useMenusQuery()
  const { data: posLocations = [] } = usePOSQuery()
  const createMenu = useCreateMenuMutation()
  const deleteMenu = useDeleteMenuMutation()
  const updateMenu = useUpdateMenuMutation()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')

  const conflicts = useMemo(
    () => findMenuBindingConflicts(menus ?? []),
    [menus],
  )
  const posNamesById = useMemo(
    () => new Map(posLocations.map((location) => [location.id, location.name])),
    [posLocations],
  )

  const handleCreate = async () => {
    if (!newName.trim()) return
    await createMenu.mutate({ name: newName.trim() })
    setNewName('')
    setShowCreate(false)
    mutate()
  }

  const handleDelete = async (id: number) => {
    await deleteMenu.mutate(id)
    mutate()
  }

  const handleToggleBinding = async (menuId: number, posId: number, nextActive: boolean) => {
    const targetMenu = menus?.find((menu) => menu.id === menuId)
    if (!targetMenu) return

    const bindings = (targetMenu.bindings ?? []).map((binding) => {
      if (binding.pos_id !== posId) return binding
      return { pos_id: binding.pos_id, is_active: nextActive }
    })

    await updateMenu.mutate({
      id: menuId,
      data: { bindings },
    })
    mutate()
  }

  const resolveBindingPosName = (posId: number, fallbackName?: string) =>
    fallbackName ?? posNamesById.get(posId) ?? `POS #${posId}`

  if (isLoading) {
    return (
      <div>
        <div className="h-8 w-48 shimmer rounded mb-6" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[0, 1, 2].map((i) => <CardSkeleton key={i} />)}
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <ErrorState
        title="Ошибка загрузки"
        message="Не удалось загрузить список меню."
        onRetry={() => mutate()}
      />
    )
  }

  return (
    <div>
      {botId && (
        <Link
          to={`/dashboard/bots/${botId}?tab=modules`}
          className="mb-4 inline-flex items-center gap-1.5 text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          Назад к модулям бота
        </Link>
      )}

      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Меню</h1>
          <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">Управление меню заведений</p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className={cn(
            'flex items-center gap-1.5 py-2 px-4 rounded text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover',
            'transition-all',
          )}
        >
          <Plus className="w-4 h-4" />
          Создать меню
        </button>
      </div>

      {conflicts.length > 0 && (
        <div className="mb-6 space-y-3 rounded-2xl border border-amber-200 bg-amber-50 p-4">
          {conflicts.map((conflict) => (
            <div key={conflict.posId} className="flex gap-3">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-600" />
              <div className="text-sm text-amber-900">
                <div className="font-medium">
                  Предупреждение: для точки продаж {conflict.posName} выбрано более одного меню.
                </div>
                <div className="mt-1">
                  Пожалуйста, выберите одно. Сейчас показывается: {conflict.fallbackMenuName}.
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="bg-white rounded border border-neutral-900 p-5 mb-6 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900 mb-3">Новое меню</h3>
          <div className="flex gap-3">
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Название меню"
              className={cn(
                'flex-1 px-3 py-2 rounded border border-neutral-200 text-sm',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              )}
              onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            />
            <button
              type="button"
              onClick={handleCreate}
              disabled={createMenu.isPending || !newName.trim()}
              className={cn(
                'py-2 px-4 rounded text-sm font-medium',
                'bg-accent text-white hover:bg-accent-hover',
                'disabled:opacity-50 transition-all',
              )}
            >
              {createMenu.isPending ? 'Создание...' : 'Создать'}
            </button>
            <button
              type="button"
              onClick={() => {
                setShowCreate(false)
                setNewName('')
              }}
              className="py-2 px-3 rounded text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {!menus || menus.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
            <UtensilsCrossed className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">Нет меню</h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
            Создайте меню вручную или импортируйте из POS-системы через интеграцию
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {menus.map((menu) => {
            const bindings = menu.bindings ?? []

            return (
              <div
                key={menu.id}
                className="bg-white rounded border border-neutral-900 p-5 transition-shadow group"
              >
                <div className="flex items-start justify-between mb-3">
                  <div className="min-w-0 flex-1">
                    <h3 className="font-semibold text-neutral-900 truncate">{menu.name}</h3>
                    <p className="text-xs text-neutral-400 mt-0.5">
                      {menu.source === 'pos_import' ? 'Импорт из POS' : 'Ручное создание'}
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={(e) => {
                      e.preventDefault()
                      handleDelete(menu.id)
                    }}
                    className="p-1.5 rounded text-neutral-300 hover:text-red-500 hover:bg-red-50 opacity-0 group-hover:opacity-100 transition-all"
                    aria-label="Удалить меню"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>

                <div className="rounded-xl border border-surface-border bg-neutral-50/70 p-3 mb-4">
                  <div className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 mb-2">
                    Активность по точкам продаж
                  </div>
                  {bindings.length === 0 ? (
                    <p className="text-sm text-neutral-500">Меню пока не привязано ни к одной точке продаж.</p>
                  ) : (
                    <div className="space-y-2">
                      {bindings.map((binding) => (
                        <div key={`${menu.id}-${binding.pos_id}`} className="flex items-center justify-between gap-3 rounded-lg bg-white px-3 py-2.5">
                          <div className="min-w-0">
                            <div className="text-sm font-medium text-neutral-900">
                              {resolveBindingPosName(binding.pos_id, binding.pos_name)}
                            </div>
                            <div className="text-xs text-neutral-500">
                              {binding.is_active ? 'Активно сейчас' : 'Не активно'}
                            </div>
                          </div>
                          <button
                            type="button"
                            role="switch"
                            aria-checked={binding.is_active}
                            onClick={() => handleToggleBinding(menu.id, binding.pos_id, !binding.is_active)}
                            className={cn(
                              'relative h-6 w-10 shrink-0 rounded-full transition-colors',
                              binding.is_active ? 'bg-accent' : 'bg-neutral-300',
                            )}
                          >
                            <span
                              className={cn(
                                'absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform',
                                binding.is_active && 'translate-x-4',
                              )}
                            />
                          </button>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="flex items-center justify-between">
                  <span className="text-xs text-neutral-400">
                    {menu.categories?.length ?? 0} категорий · {formatDate(menu.created_at)}
                  </span>
                  <Link
                    to={`/dashboard/menus/${menu.id}${botId ? `?botId=${botId}` : ''}`}
                    className="flex items-center gap-1 text-xs font-medium text-accent hover:text-accent-hover transition-colors"
                  >
                    Открыть
                    <ChevronRight className="w-3.5 h-3.5" />
                  </Link>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
