import { useEffect, useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  useAddCategoryMutation,
  useAddItemMutation,
  useMenuQuery,
  useUpdateCategoryMutation,
  useUpdateItemMutation,
  useUpdateMenuMutation,
} from '@/features/menus/queries'
import type {
  CreateMenuItemRequest,
  Menu,
  MenuCategory,
  MenuItem,
  MenuPOSBindingRequest,
} from '@/features/menus/types'
import { usePOSQuery } from '@/features/pos/queries'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { MessageContentEditor, type MessageContent } from '@/features/telegram-preview'
import { campaignsApi } from '@/features/campaigns/api'
import {
  ArrowLeft,
  Check,
  ChevronDown,
  ChevronRight,
  Edit3,
  Package,
  Plus,
  Save,
  Trash2,
  X,
} from 'lucide-react'

const inputClassName = cn(
  'w-full px-3 py-2 rounded border border-neutral-200 text-sm',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
)

function normalizeMenuDraft(menu: Menu): Menu {
  return {
    ...menu,
    intro_content: menu.intro_content ?? {
      parts: [{ type: 'text', text: '', parse_mode: 'Markdown' }],
    },
    bindings: menu.bindings ?? [],
  }
}

export default function MenuDetailPage() {
  const { menuId } = useParams<{ menuId: string }>()
  const [searchParams] = useSearchParams()
  const botId = searchParams.get('botId')
  const id = Number(menuId)
  const { data: menu, isLoading, isError, mutate } = useMenuQuery(isNaN(id) ? 0 : id)
  const { data: posLocations = [] } = usePOSQuery()
  const updateMenu = useUpdateMenuMutation()
  const addCategory = useAddCategoryMutation(id)
  const [draft, setDraft] = useState<Menu | null>(null)

  const [newCategory, setNewCategory] = useState({
    name: '',
    icon_emoji: '',
    icon_image_url: '',
  })
  const [showAddCategory, setShowAddCategory] = useState(false)

  useEffect(() => {
    if (menu) {
      setDraft(normalizeMenuDraft(menu))
    }
  }, [menu])

  const backToModulesHref = botId ? `/dashboard/bots/${botId}?tab=modules` : '/dashboard/menus'
  const backToMenusHref = botId ? `/dashboard/menus?botId=${botId}` : '/dashboard/menus'

  const bindingsByPosId = useMemo(() => {
    const current = new Map<number, MenuPOSBindingRequest>()
    for (const binding of draft?.bindings ?? []) {
      current.set(binding.pos_id, {
        pos_id: binding.pos_id,
        is_active: binding.is_active,
      })
    }
    return current
  }, [draft?.bindings])

  const handleSaveMenu = async () => {
    if (!draft) return
    await updateMenu.mutate({
      id,
      data: {
        name: draft.name,
        intro_content: draft.intro_content,
        bindings: Array.from(bindingsByPosId.values()),
      },
    })
    mutate()
  }

  const handleAddCategory = async () => {
    if (!newCategory.name.trim()) return
    await addCategory.mutate({
      name: newCategory.name.trim(),
      icon_emoji: newCategory.icon_emoji.trim() || undefined,
      icon_image_url: newCategory.icon_image_url.trim() || undefined,
    })
    setNewCategory({ name: '', icon_emoji: '', icon_image_url: '' })
    setShowAddCategory(false)
    mutate()
  }

  const handleNewCategoryIconUpload = async (file?: File | null) => {
    if (!file) return
    const uploaded = await campaignsApi.uploadFile(file)
    setNewCategory((current) => ({ ...current, icon_image_url: uploaded }))
  }

  if (isLoading || !draft) {
    return (
      <div className="max-w-4xl">
        <div className="h-4 w-32 shimmer rounded mb-6" />
        <CardSkeleton />
      </div>
    )
  }

  if (isError || !menu) {
    return (
      <div className="max-w-3xl">
        <Link
          to={backToMenusHref}
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Назад
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
    <div className="max-w-4xl space-y-6">
      <div className="flex flex-wrap items-center gap-3">
        <Link
          to={backToModulesHref}
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
        >
          <ArrowLeft className="w-4 h-4" />
          {botId ? 'Назад к модулям' : 'Назад к списку'}
        </Link>
        {botId && (
          <Link
            to={backToMenusHref}
            className="inline-flex items-center gap-1.5 text-sm text-accent hover:text-accent-hover transition-colors"
          >
            Все меню
            <ChevronRight className="h-4 w-4" />
          </Link>
        )}
      </div>

      <div className="bg-white rounded border border-neutral-900 p-6 animate-in">
        <div className="flex flex-wrap items-start justify-between gap-4 mb-6">
          <div className="min-w-0">
            <input
              type="text"
              value={draft.name}
              onChange={(event) =>
                setDraft((current) => (current ? { ...current, name: event.target.value } : current))
              }
              className={cn(inputClassName, 'text-2xl font-serif font-bold border-none px-0 py-0 focus:ring-0')}
            />
            <p className="text-xs text-neutral-400 mt-1">
              {draft.source === 'pos_import' ? 'Импорт из POS' : 'Ручное создание'}
              {draft.last_synced_at && ` · Синхронизировано: ${new Date(draft.last_synced_at).toLocaleDateString('ru-RU')}`}
            </p>
          </div>
          <div className="flex gap-2">
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
            <button
              type="button"
              onClick={handleSaveMenu}
              className="inline-flex items-center gap-1.5 rounded bg-neutral-900 px-4 py-2 text-sm font-medium text-white hover:bg-neutral-700"
            >
              <Save className="h-4 w-4" />
              Сохранить
            </button>
          </div>
        </div>

        <section className="rounded-xl border border-surface-border bg-neutral-50/70 p-4 mb-6">
          <div className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 mb-3">
            Первое сообщение
          </div>
          <MessageContentEditor
            value={draft.intro_content as MessageContent}
            onChange={(content) =>
              setDraft((current) => (current ? { ...current, intro_content: content } : current))
            }
            onUpload={campaignsApi.uploadFile}
            maxParts={4}
          />
        </section>

        <section className="rounded-xl border border-surface-border bg-neutral-50/70 p-4">
          <div className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 mb-3">
            Привязка к точкам продаж
          </div>
          <div className="space-y-2">
            {posLocations.map((location) => {
              const binding = bindingsByPosId.get(location.id)
              const isBound = Boolean(binding)
              const isActive = binding?.is_active ?? false
              return (
                <div key={location.id} className="rounded-lg border border-surface-border bg-white px-3 py-3">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <label className="flex items-center gap-3">
                      <input
                        type="checkbox"
                        checked={isBound}
                        onChange={() =>
                          setDraft((current) => {
                            if (!current) return current
                            const nextBindings = [...(current.bindings ?? [])]
                            const index = nextBindings.findIndex((item) => item.pos_id === location.id)
                            if (index >= 0) {
                              nextBindings.splice(index, 1)
                            } else {
                              nextBindings.push({
                                menu_id: current.id,
                                pos_id: location.id,
                                pos_name: location.name,
                                is_active: false,
                              })
                            }
                            return { ...current, bindings: nextBindings }
                          })
                        }
                        className="h-4 w-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
                      />
                      <div>
                        <div className="text-sm font-medium text-neutral-900">{location.name}</div>
                        <div className="text-xs text-neutral-500">{location.address}</div>
                      </div>
                    </label>

                    <button
                      type="button"
                      role="switch"
                      aria-checked={isActive}
                      disabled={!isBound}
                      onClick={() =>
                        setDraft((current) => {
                          if (!current) return current
                          return {
                            ...current,
                            bindings: (current.bindings ?? []).map((item) => (
                              item.pos_id === location.id
                                ? { ...item, is_active: !item.is_active }
                                : item
                            )),
                          }
                        })
                      }
                      className={cn(
                        'relative h-6 w-10 shrink-0 rounded-full transition-colors disabled:opacity-40',
                        isActive ? 'bg-accent' : 'bg-neutral-300',
                      )}
                    >
                      <span
                        className={cn(
                          'absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform',
                          isActive && 'translate-x-4',
                        )}
                      />
                    </button>
                  </div>
                </div>
              )
            })}
          </div>
        </section>
      </div>

      {showAddCategory && (
        <div className="bg-white rounded border border-neutral-900 p-5 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900 mb-3">Новая категория</h3>
          <div className="grid gap-3 md:grid-cols-4">
            <input
              type="text"
              value={newCategory.name}
              onChange={(event) => setNewCategory((current) => ({ ...current, name: event.target.value }))}
              placeholder="Название категории"
              className={inputClassName}
            />
            <input
              type="text"
              value={newCategory.icon_emoji}
              onChange={(event) => setNewCategory((current) => ({ ...current, icon_emoji: event.target.value }))}
              placeholder="Emoji"
              className={inputClassName}
            />
            <input
              type="text"
              value={newCategory.icon_image_url}
              onChange={(event) => setNewCategory((current) => ({ ...current, icon_image_url: event.target.value }))}
              placeholder="URL иконки"
              className={inputClassName}
            />
            <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded-lg border border-dashed border-surface-border px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
              Загрузить иконку
              <input
                type="file"
                accept="image/png,image/svg+xml,image/*"
                className="hidden"
                onChange={(event) => void handleNewCategoryIconUpload(event.target.files?.[0])}
              />
            </label>
          </div>
          <div className="mt-4 flex gap-3">
            <button
              type="button"
              onClick={handleAddCategory}
              disabled={addCategory.isPending || !newCategory.name.trim()}
              className="rounded bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover disabled:opacity-50"
            >
              Добавить
            </button>
            <button
              type="button"
              onClick={() => {
                setShowAddCategory(false)
                setNewCategory({ name: '', icon_emoji: '', icon_image_url: '' })
              }}
              className="rounded px-4 py-2 text-sm text-neutral-500 hover:text-neutral-700"
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
          {categories.map((category) => (
            <CategorySection
              key={category.id}
              menuId={id}
              category={category}
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
  category: MenuCategory
  onUpdate: () => void
}) {
  const [expanded, setExpanded] = useState(true)
  const [showAddItem, setShowAddItem] = useState(false)
  const [draftCategory, setDraftCategory] = useState({
    name: category.name,
    icon_emoji: category.icon_emoji ?? '',
    icon_image_url: category.icon_image_url ?? '',
  })
  const [newItem, setNewItem] = useState<CreateMenuItemRequest>({
    name: '',
    description: '',
    price: 0,
    weight: '',
    image_url: '',
    tags: [],
  })

  const addItem = useAddItemMutation(menuId, category.id)
  const updateCategory = useUpdateCategoryMutation()

  useEffect(() => {
    setDraftCategory({
      name: category.name,
      icon_emoji: category.icon_emoji ?? '',
      icon_image_url: category.icon_image_url ?? '',
    })
  }, [category.icon_emoji, category.icon_image_url, category.name])

  const handleSaveCategory = async () => {
    await updateCategory.mutate({
      categoryId: category.id,
      data: {
        name: draftCategory.name,
        icon_emoji: draftCategory.icon_emoji || undefined,
        icon_image_url: draftCategory.icon_image_url || undefined,
      },
    })
    onUpdate()
  }

  const handleCategoryIconUpload = async (file?: File | null) => {
    if (!file) return
    const uploaded = await campaignsApi.uploadFile(file)
    setDraftCategory((current) => ({ ...current, icon_image_url: uploaded }))
  }

  const handleAddItem = async () => {
    if (!newItem.name.trim()) return
    await addItem.mutate({
      ...newItem,
      name: newItem.name.trim(),
      description: newItem.description?.trim() || undefined,
      weight: newItem.weight?.trim() || undefined,
      image_url: newItem.image_url?.trim() || undefined,
    })
    setNewItem({
      name: '',
      description: '',
      price: 0,
      weight: '',
      image_url: '',
      tags: [],
    })
    setShowAddItem(false)
    onUpdate()
  }

  const handleNewItemImageUpload = async (file?: File | null) => {
    if (!file) return
    const uploaded = await campaignsApi.uploadFile(file)
    setNewItem((current) => ({ ...current, image_url: uploaded }))
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
          <h3 className="text-base font-semibold text-neutral-900">
            {draftCategory.icon_emoji ? `${draftCategory.icon_emoji} ` : ''}
            {category.name}
          </h3>
          <span className="text-xs text-neutral-400">{items.length} позиций</span>
        </div>
        <ChevronDown className={cn('w-4 h-4 text-neutral-400 transition-transform', expanded && 'rotate-180')} />
      </button>

      {expanded && (
        <div className="border-t border-neutral-200 space-y-4 p-6">
          <div className="grid gap-3 md:grid-cols-[minmax(0,1fr)_120px_minmax(0,1fr)_auto_auto]">
            <input
              type="text"
              value={draftCategory.name}
              onChange={(event) => setDraftCategory((current) => ({ ...current, name: event.target.value }))}
              placeholder="Название категории"
              className={inputClassName}
            />
            <input
              type="text"
              value={draftCategory.icon_emoji}
              onChange={(event) => setDraftCategory((current) => ({ ...current, icon_emoji: event.target.value }))}
              placeholder="Emoji"
              className={inputClassName}
            />
            <input
              type="text"
              value={draftCategory.icon_image_url}
              onChange={(event) => setDraftCategory((current) => ({ ...current, icon_image_url: event.target.value }))}
              placeholder="URL иконки"
              className={inputClassName}
            />
            <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded-lg border border-dashed border-surface-border px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
              Загрузить
              <input
                type="file"
                accept="image/png,image/svg+xml,image/*"
                className="hidden"
                onChange={(event) => void handleCategoryIconUpload(event.target.files?.[0])}
              />
            </label>
            <button
              type="button"
              onClick={handleSaveCategory}
              className="inline-flex min-h-11 items-center justify-center rounded-lg bg-neutral-900 px-4 text-sm font-medium text-white hover:bg-neutral-700"
            >
              Сохранить
            </button>
          </div>

          {items.length === 0 ? (
            <p className="text-sm text-neutral-400 text-center py-6">Нет позиций</p>
          ) : (
            <div className="space-y-3">
              {items.map((item) => (
                <MenuItemRow key={item.id} item={item} onUpdate={onUpdate} />
              ))}
            </div>
          )}

          <div className="border-t border-neutral-200/70 pt-4">
            {showAddItem ? (
              <div className="space-y-3">
                <div className="grid gap-3 md:grid-cols-2">
                  <input
                    type="text"
                    value={newItem.name}
                    onChange={(event) => setNewItem((current) => ({ ...current, name: event.target.value }))}
                    placeholder="Название"
                    className={inputClassName}
                  />
                  <input
                    type="number"
                    value={newItem.price || ''}
                    onChange={(event) => setNewItem((current) => ({ ...current, price: Number(event.target.value) }))}
                    placeholder="Цена"
                    className={inputClassName}
                  />
                  <input
                    type="text"
                    value={newItem.weight ?? ''}
                    onChange={(event) => setNewItem((current) => ({ ...current, weight: event.target.value }))}
                    placeholder="Граммаж"
                    className={inputClassName}
                  />
                  <input
                    type="text"
                    value={newItem.image_url ?? ''}
                    onChange={(event) => setNewItem((current) => ({ ...current, image_url: event.target.value }))}
                    placeholder="URL картинки"
                    className={inputClassName}
                  />
                </div>
                <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded-lg border border-dashed border-surface-border px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
                  Загрузить фотографию
                  <input
                    type="file"
                    accept="image/*"
                    className="hidden"
                    onChange={(event) => void handleNewItemImageUpload(event.target.files?.[0])}
                  />
                </label>
                <textarea
                  rows={3}
                  value={newItem.description ?? ''}
                  onChange={(event) => setNewItem((current) => ({ ...current, description: event.target.value }))}
                  placeholder="Описание"
                  className={inputClassName}
                />
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={handleAddItem}
                    disabled={addItem.isPending}
                    className="inline-flex items-center gap-1.5 rounded bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover disabled:opacity-50"
                  >
                    <Check className="h-4 w-4" />
                    Добавить позицию
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setShowAddItem(false)
                      setNewItem({
                        name: '',
                        description: '',
                        price: 0,
                        weight: '',
                        image_url: '',
                        tags: [],
                      })
                    }}
                    className="inline-flex items-center gap-1.5 rounded px-4 py-2 text-sm text-neutral-500 hover:text-neutral-700"
                  >
                    <X className="h-4 w-4" />
                    Отмена
                  </button>
                </div>
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
  const [draft, setDraft] = useState({
    name: item.name,
    description: item.description ?? '',
    price: item.price,
    weight: item.weight ?? '',
    image_url: item.image_url ?? '',
    is_available: item.is_available,
  })
  const updateItem = useUpdateItemMutation()

  useEffect(() => {
    setDraft({
      name: item.name,
      description: item.description ?? '',
      price: item.price,
      weight: item.weight ?? '',
      image_url: item.image_url ?? '',
      is_available: item.is_available,
    })
  }, [item])

  const handleSave = async () => {
    await updateItem.mutate({
      itemId: item.id,
      data: {
        name: draft.name,
        description: draft.description || undefined,
        price: Number(draft.price),
        weight: draft.weight || undefined,
        image_url: draft.image_url || undefined,
        is_available: draft.is_available,
      },
    })
    setEditing(false)
    onUpdate()
  }

  const handleItemImageUpload = async (file?: File | null) => {
    if (!file) return
    const uploaded = await campaignsApi.uploadFile(file)
    setDraft((current) => ({ ...current, image_url: uploaded }))
  }

  return (
    <div className="rounded-xl border border-surface-border bg-neutral-50/60 p-4">
      {editing ? (
        <div className="space-y-3">
          <div className="grid gap-3 md:grid-cols-2">
            <input
              type="text"
              value={draft.name}
              onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))}
              className={inputClassName}
              placeholder="Название"
            />
            <input
              type="number"
              value={draft.price}
              onChange={(event) => setDraft((current) => ({ ...current, price: Number(event.target.value) }))}
              className={inputClassName}
              placeholder="Цена"
            />
            <input
              type="text"
              value={draft.weight}
              onChange={(event) => setDraft((current) => ({ ...current, weight: event.target.value }))}
              className={inputClassName}
              placeholder="Граммаж"
            />
            <input
              type="text"
              value={draft.image_url}
              onChange={(event) => setDraft((current) => ({ ...current, image_url: event.target.value }))}
              className={inputClassName}
              placeholder="URL картинки"
            />
          </div>
          <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded-lg border border-dashed border-surface-border px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
            Загрузить фотографию
            <input
              type="file"
              accept="image/*"
              className="hidden"
              onChange={(event) => void handleItemImageUpload(event.target.files?.[0])}
            />
          </label>
          <textarea
            rows={3}
            value={draft.description}
            onChange={(event) => setDraft((current) => ({ ...current, description: event.target.value }))}
            className={inputClassName}
            placeholder="Описание"
          />
          <label className="flex items-center gap-2 text-sm text-neutral-700">
            <input
              type="checkbox"
              checked={draft.is_available}
              onChange={(event) => setDraft((current) => ({ ...current, is_available: event.target.checked }))}
              className="h-4 w-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
            />
            Позиция активна
          </label>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleSave}
              className="inline-flex items-center gap-1.5 rounded bg-accent px-4 py-2 text-sm font-medium text-white hover:bg-accent-hover"
            >
              <Check className="h-4 w-4" />
              Сохранить
            </button>
            <button
              type="button"
              onClick={() => setEditing(false)}
              className="inline-flex items-center gap-1.5 rounded px-4 py-2 text-sm text-neutral-500 hover:text-neutral-700"
            >
              <X className="h-4 w-4" />
              Отмена
            </button>
          </div>
        </div>
      ) : (
        <div className="flex items-start justify-between gap-4">
          <div className="min-w-0 flex-1">
            <div className={cn('text-sm font-medium text-neutral-900', !item.is_available && 'line-through opacity-50')}>
              {item.name}
            </div>
            {item.description && (
              <div className="text-xs text-neutral-500 mt-1">{item.description}</div>
            )}
            <div className="mt-2 flex flex-wrap gap-3 text-xs text-neutral-500">
              <span>{new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(item.price)}</span>
              {item.weight ? <span>{item.weight}</span> : null}
              {item.image_url ? <span>Есть фото</span> : null}
              {!item.is_available ? <span>Не активно</span> : null}
            </div>
          </div>
          <button
            type="button"
            onClick={() => setEditing(true)}
            className="inline-flex min-h-11 min-w-11 items-center justify-center rounded-lg text-neutral-400 hover:bg-white hover:text-neutral-700"
          >
            <Edit3 className="h-4 w-4" />
          </button>
        </div>
      )}
    </div>
  )
}
