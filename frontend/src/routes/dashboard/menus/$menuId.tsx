import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  useMenuQuery,
  useAddCategoryMutation,
  useAddItemMutation,
  useUpdateItemMutation,
} from '@/features/menus/queries'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ArrowLeft, Plus, ChevronDown, Package, Edit3, Check, X } from 'lucide-react'
import type { MenuItem } from '@/features/menus/types'

function formatCurrency(amount: number) {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(amount)
}

const inputClassName = cn(
  'w-full px-3 py-2 rounded border border-neutral-200 text-sm',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
)

export default function MenuDetailPage() {
  const { menuId } = useParams<{ menuId: string }>()
  const id = Number(menuId)
  const { data: menu, isLoading, isError, mutate } = useMenuQuery(isNaN(id) ? 0 : id)

  const [newCatName, setNewCatName] = useState('')
  const [showAddCategory, setShowAddCategory] = useState(false)
  const addCategory = useAddCategoryMutation(id)

  const handleAddCategory = async () => {
    if (!newCatName.trim()) return
    await addCategory.mutate({ name: newCatName.trim() })
    setNewCatName('')
    setShowAddCategory(false)
    mutate()
  }

  if (isLoading) {
    return (
      <div className="max-w-3xl">
        <div className="h-4 w-32 shimmer rounded mb-6" />
        <CardSkeleton />
      </div>
    )
  }

  if (isError || !menu) {
    return (
      <div className="max-w-3xl">
        <Link
          to="/dashboard/menus"
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Назад к списку
        </Link>
        <ErrorState
          title="Меню не найдено"
          message="Проверьте URL или вернитесь к списку."
          onRetry={() => mutate()}
        />
      </div>
    )
  }

  const categories = menu.categories ?? []

  return (
    <div className="max-w-3xl">
      <Link
        to="/dashboard/menus"
        className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        Назад к списку
      </Link>

      <div className="bg-white rounded border border-neutral-900 p-6 mb-6 animate-in">
        <div className="flex items-center justify-between">
          <div>
            <h1 className="font-serif font-serif text-3xl font-bold text-neutral-900 tracking-tight">
              {menu.name}
            </h1>
            <p className="text-xs text-neutral-400 mt-1">
              {menu.source === 'pos_import' ? 'Импорт из POS' : 'Ручное создание'}
              {menu.last_synced_at && ` \u00b7 Синхронизировано: ${new Date(menu.last_synced_at).toLocaleDateString('ru-RU')}`}
            </p>
          </div>
          <button
            type="button"
            onClick={() => setShowAddCategory(true)}
            className={cn(
              'flex items-center gap-1.5 py-2 px-3 rounded text-sm font-medium',
              'bg-accent text-white hover:bg-accent-hover transition-all',
            )}
          >
            <Plus className="w-4 h-4" />
            Категория
          </button>
        </div>
      </div>

      {showAddCategory && (
        <div className="bg-white rounded border border-neutral-900 p-5 mb-6 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900 mb-3">Новая категория</h3>
          <div className="flex gap-3">
            <input
              type="text"
              value={newCatName}
              onChange={(e) => setNewCatName(e.target.value)}
              placeholder="Название категории"
              className={cn(inputClassName, 'flex-1')}
              onKeyDown={(e) => e.key === 'Enter' && handleAddCategory()}
            />
            <button
              type="button"
              onClick={handleAddCategory}
              disabled={addCategory.isPending || !newCatName.trim()}
              className={cn(
                'py-2 px-4 rounded text-sm font-medium',
                'bg-accent text-white hover:bg-accent-hover',
                'disabled:opacity-50 transition-all',
              )}
            >
              Добавить
            </button>
            <button
              type="button"
              onClick={() => { setShowAddCategory(false); setNewCatName('') }}
              className="py-2 px-3 text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {categories.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <div className="w-12 h-12 rounded bg-neutral-100 flex items-center justify-center mb-3">
            <Package className="w-6 h-6 text-neutral-400" />
          </div>
          <p className="text-sm text-neutral-500">Нет категорий. Добавьте первую категорию.</p>
        </div>
      ) : (
        <div className="space-y-4">
          {categories.map((cat) => (
            <CategorySection
              key={cat.id}
              menuId={id}
              category={cat}
              onUpdate={() => mutate()}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function CategorySection({
  menuId,
  category,
  onUpdate,
}: {
  menuId: number
  category: { id: number; name: string; items?: MenuItem[] }
  onUpdate: () => void
}) {
  const [expanded, setExpanded] = useState(true)
  const [showAddItem, setShowAddItem] = useState(false)
  const [newItem, setNewItem] = useState({ name: '', price: '' })
  const addItem = useAddItemMutation(menuId, category.id)

  const handleAddItem = async () => {
    if (!newItem.name.trim() || !newItem.price) return
    await addItem.mutate({
      name: newItem.name.trim(),
      price: Number(newItem.price),
    })
    setNewItem({ name: '', price: '' })
    setShowAddItem(false)
    onUpdate()
  }

  const items = category.items ?? []

  return (
    <section className="bg-white rounded border border-neutral-900 overflow-hidden animate-in">
      <button
        type="button"
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center justify-between px-6 py-4 hover:bg-neutral-50 transition-colors"
      >
        <div className="flex items-center gap-2">
          <h3 className="text-base font-semibold text-neutral-900">{category.name}</h3>
          <span className="text-xs text-neutral-400">{items.length} позиций</span>
        </div>
        <ChevronDown className={cn('w-4 h-4 text-neutral-400 transition-transform', expanded && 'rotate-180')} />
      </button>

      {expanded && (
        <div className="border-t border-neutral-200">
          {items.length === 0 ? (
            <p className="text-sm text-neutral-400 text-center py-6">Нет позиций</p>
          ) : (
            <div className="divide-y divide-neutral-200/50">
              {items.map((item) => (
                <MenuItemRow key={item.id} item={item} onUpdate={onUpdate} />
              ))}
            </div>
          )}

          <div className="px-6 py-3 border-t border-neutral-200/50">
            {showAddItem ? (
              <div className="flex gap-2">
                <input
                  type="text"
                  value={newItem.name}
                  onChange={(e) => setNewItem({ ...newItem, name: e.target.value })}
                  placeholder="Название"
                  className={cn(inputClassName, 'flex-1')}
                />
                <input
                  type="number"
                  value={newItem.price}
                  onChange={(e) => setNewItem({ ...newItem, price: e.target.value })}
                  placeholder="Цена"
                  className={cn(inputClassName, 'w-28')}
                />
                <button
                  type="button"
                  onClick={handleAddItem}
                  disabled={addItem.isPending}
                  className="p-2 rounded bg-accent text-white hover:bg-accent-hover disabled:opacity-50 transition-all"
                >
                  <Check className="w-4 h-4" />
                </button>
                <button
                  type="button"
                  onClick={() => { setShowAddItem(false); setNewItem({ name: '', price: '' }) }}
                  className="p-2 rounded text-neutral-400 hover:text-neutral-600 transition-colors"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setShowAddItem(true)}
                className="flex items-center gap-1.5 text-sm text-accent hover:text-accent-hover font-medium transition-colors"
              >
                <Plus className="w-3.5 h-3.5" />
                Добавить позицию
              </button>
            )}
          </div>
        </div>
      )}
    </section>
  )
}

function MenuItemRow({ item, onUpdate }: { item: MenuItem; onUpdate: () => void }) {
  const [editing, setEditing] = useState(false)
  const [editPrice, setEditPrice] = useState(String(item.price))
  const updateItem = useUpdateItemMutation()

  const handleSave = async () => {
    await updateItem.mutate({ itemId: item.id, data: { price: Number(editPrice) } })
    setEditing(false)
    onUpdate()
  }

  return (
    <div className="flex items-center justify-between px-6 py-3 hover:bg-neutral-50/50 transition-colors group">
      <div className="min-w-0 flex-1">
        <p className={cn('text-sm text-neutral-800', !item.is_available && 'line-through opacity-50')}>
          {item.name}
        </p>
        {item.description && (
          <p className="text-xs text-neutral-400 mt-0.5 truncate">{item.description}</p>
        )}
      </div>
      <div className="flex items-center gap-2 ml-4">
        {editing ? (
          <>
            <input
              type="number"
              value={editPrice}
              onChange={(e) => setEditPrice(e.target.value)}
              className={cn(inputClassName, 'w-24 text-right')}
              onKeyDown={(e) => e.key === 'Enter' && handleSave()}
            />
            <button type="button" onClick={handleSave} className="p-1 text-green-600 hover:text-green-700">
              <Check className="w-4 h-4" />
            </button>
            <button type="button" onClick={() => setEditing(false)} className="p-1 text-neutral-400 hover:text-neutral-600">
              <X className="w-4 h-4" />
            </button>
          </>
        ) : (
          <>
            <span className="text-sm font-medium text-neutral-900 tabular-nums whitespace-nowrap">
              {new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(item.price)}
            </span>
            <button
              type="button"
              onClick={() => { setEditing(true); setEditPrice(String(item.price)) }}
              className="p-1 rounded text-neutral-300 hover:text-neutral-500 opacity-0 group-hover:opacity-100 transition-all"
            >
              <Edit3 className="w-3.5 h-3.5" />
            </button>
          </>
        )}
      </div>
    </div>
  )
}
