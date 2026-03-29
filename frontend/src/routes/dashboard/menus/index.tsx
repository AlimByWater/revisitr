import { useState } from 'react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useMenusQuery, useCreateMenuMutation, useDeleteMenuMutation } from '@/features/menus/queries'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { UtensilsCrossed, Plus, Trash2, ChevronRight } from 'lucide-react'

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export default function MenusPage() {
  const { data: menus, isLoading, isError, mutate } = useMenusQuery()
  const createMenu = useCreateMenuMutation()
  const deleteMenu = useDeleteMenuMutation()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')

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
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Меню</h1>
          <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">Управление меню заведений</p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className={cn(
            'flex items-center gap-1.5 py-2 px-4 rounded-lg text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover',
            'transition-all shadow-sm shadow-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          Создать меню
        </button>
      </div>

      {showCreate && (
        <div className="bg-white rounded-2xl border border-surface-border p-5 mb-6 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900 mb-3">Новое меню</h3>
          <div className="flex gap-3">
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Название меню"
              className={cn(
                'flex-1 px-3 py-2 rounded-lg border border-surface-border text-sm',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              )}
              onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            />
            <button
              type="button"
              onClick={handleCreate}
              disabled={createMenu.isPending || !newName.trim()}
              className={cn(
                'py-2 px-4 rounded-lg text-sm font-medium',
                'bg-accent text-white hover:bg-accent-hover',
                'disabled:opacity-50 transition-all',
              )}
            >
              {createMenu.isPending ? 'Создание...' : 'Создать'}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName('') }}
              className="py-2 px-3 rounded-lg text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {!menus || menus.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mb-4">
            <UtensilsCrossed className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">Нет меню</h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
            Создайте меню вручную или импортируйте из POS-системы через интеграцию
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {menus.map((menu) => (
            <div
              key={menu.id}
              className="bg-white rounded-2xl border border-surface-border p-5 hover:shadow-sm transition-shadow group"
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
                  onClick={(e) => { e.preventDefault(); handleDelete(menu.id) }}
                  className="p-1.5 rounded-lg text-neutral-300 hover:text-red-500 hover:bg-red-50 opacity-0 group-hover:opacity-100 transition-all"
                  aria-label="Удалить меню"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>

              <div className="flex items-center justify-between">
                <span className="text-xs text-neutral-400">
                  {menu.categories?.length ?? 0} категорий &middot; {formatDate(menu.created_at)}
                </span>
                <Link
                  to={`/dashboard/menus/${menu.id}`}
                  className="flex items-center gap-1 text-xs font-medium text-accent hover:text-accent-hover transition-colors"
                >
                  Открыть
                  <ChevronRight className="w-3.5 h-3.5" />
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
