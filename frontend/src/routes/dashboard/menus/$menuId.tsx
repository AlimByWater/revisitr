import { useEffect, useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useAddCategoryMutation, useAddItemMutation, useDeleteCategoryMutation, useDeleteItemMutation, useMenuQuery, useUpdateCategoryMutation, useUpdateItemMutation, useUpdateMenuMutation } from '@/features/menus/queries'
import { menusApi } from '@/features/menus/api'
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
import { campaignsApi } from '@/features/campaigns/api'
import { MenuDisplaySettingsPanel } from '@/routes/dashboard/bots/menu'
import {
  ArrowLeft,
  Check,
  ChevronDown,
  ChevronRight,
  Edit3,
  Info,
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
  const activeTab = searchParams.get('tab') === 'display' ? 'display' : 'content'
  const id = Number(menuId)
  const { data: menu, isLoading, isError, mutate } = useMenuQuery(isNaN(id) ? 0 : id)
  const { data: posLocations = [] } = usePOSQuery()
  const updateMenu = useUpdateMenuMutation()
  const addCategory = useAddCategoryMutation(id)
  const [draft, setDraft] = useState<Menu | null>(null)
  const [botPosIds, setBotPosIds] = useState<number[]>([])

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

  useEffect(() => {
    if (!botId) {
      setBotPosIds([])
      return
    }

    let mounted = true
    menusApi.getBotPOSLocations(Number(botId))
      .then((response) => {
        if (!mounted) return
        setBotPosIds(response.pos_ids ?? [])
      })
      .catch(() => {
        if (!mounted) return
        setBotPosIds([])
      })

    return () => {
      mounted = false
    }
  }, [botId])

  const backToModulesHref = botId ? `/dashboard/bots/${botId}?tab=modules` : '/dashboard/menus'
  const backToMenusHref = botId ? `/dashboard/menus?botId=${botId}` : '/dashboard/menus'
  const contentTabHref = botId
    ? `/dashboard/menus/${id}?botId=${botId}&tab=content`
    : `/dashboard/menus/${id}?tab=content`
  const displayTabHref = botId
    ? `/dashboard/menus/${id}?botId=${botId}&tab=display`
    : `/dashboard/menus/${id}?tab=display`

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
  const availablePosLocations = useMemo(
    () => (botId ? posLocations.filter((location) => botPosIds.includes(location.id)) : posLocations),
    [botId, botPosIds, posLocations],
  )

  const handleSaveMenu = async () => {
    if (!draft) return
    const allowedPosIds = new Set(availablePosLocations.map((location) => location.id))
    await updateMenu.mutate({
      id,
      data: {
        name: draft.name,
        intro_content: draft.intro_content,
        bindings: Array.from(bindingsByPosId.values()).filter((binding) => allowedPosIds.has(binding.pos_id)),
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
    <div className={cn(activeTab === 'display' ? 'max-w-7xl' : 'max-w-4xl', 'space-y-6')}>
      <div className="flex flex-wrap items-center gap-3 animate-in">
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

      {/* Heading extracted outside the white card */}
      <div className="flex flex-wrap items-end justify-between gap-4 animate-in animate-in-delay-1">
        <div className="min-w-0 flex-1">
          {activeTab === 'content' ? (
            <input
              type="text"
              value={draft.name}
              onChange={(event) =>
                setDraft((current) => (current ? { ...current, name: event.target.value } : current))
              }
              className="w-full text-3xl font-display font-bold text-neutral-900 tracking-tight bg-transparent border-none px-0 py-0 focus:outline-none focus:ring-0 placeholder:text-neutral-300"
              placeholder="Название меню"
            />
          ) : (
            <h1 className="font-display text-3xl font-bold tracking-tight text-neutral-900">{draft.name}</h1>
          )}
          <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">
            {draft.source === 'pos_import' ? 'Импорт из POS' : 'Ручное создание'}
            {draft.last_synced_at && ` · Синхронизировано: ${new Date(draft.last_synced_at).toLocaleDateString('ru-RU')}`}
          </p>
        </div>
        <div className="flex gap-2">
          {activeTab === 'content' && (
            <>
              <button
                type="button"
                onClick={() => setShowAddCategory(true)}
                className={cn(
                  'inline-flex items-center gap-1.5 py-2 px-3 rounded text-sm font-medium',
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
            </>
          )}
        </div>
      </div>

      <div className="flex gap-2 rounded-xl border border-surface-border bg-white p-1">
        <Link
          to={contentTabHref}
          className={cn(
            'flex-1 rounded-lg px-4 py-2 text-center text-sm font-medium transition-colors',
            activeTab === 'content'
              ? 'bg-neutral-900 text-white'
              : 'text-neutral-500 hover:bg-neutral-50 hover:text-neutral-900',
          )}
        >
          Категории и позиции
        </Link>
        <Link
          to={displayTabHref}
          className={cn(
            'flex-1 rounded-lg px-4 py-2 text-center text-sm font-medium transition-colors',
            activeTab === 'display'
              ? 'bg-neutral-900 text-white'
              : 'text-neutral-500 hover:bg-neutral-50 hover:text-neutral-900',
          )}
        >
          Отображение в боте
        </Link>
      </div>

      {activeTab === 'display' ? (
        botId ? (
          <MenuDisplaySettingsPanel botId={Number(botId)} activeMenuId={id} embedded />
        ) : (
          <section className="rounded-xl border border-surface-border bg-white p-6">
            <h2 className="text-lg font-semibold text-neutral-900">Нужен контекст бота</h2>
            <p className="mt-2 text-sm text-neutral-500">
              Настройки отображения сохраняются для конкретного бота. Откройте меню из раздела «Мои боты → Модули».
            </p>
          </section>
        )
      ) : (
      <>
        <section className="rounded-xl border border-blue-100 bg-blue-50/60 p-4 mb-6">
          <div className="flex items-start gap-3">
            <Info className="w-5 h-5 text-blue-500 shrink-0 mt-0.5" />
            <div>
              <div className="text-sm font-medium text-blue-900">Шаблон отображения</div>
              <div className="text-sm text-blue-700 mt-1">
                Настройте как меню будет выглядеть в боте (таб-категории, список или карусель) в{' '}
                <Link to={displayTabHref} className="underline font-medium hover:text-blue-900">
                  вкладке «Отображение в боте»
                </Link>.
              </div>
            </div>
          </div>
        </section>

        <section className="rounded-xl border border-surface-border bg-neutral-50/70 p-4">
          <div className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 mb-3">
            Привязка к точкам продаж
          </div>
          <div className="space-y-2">
            {availablePosLocations.map((location) => {
              const binding = bindingsByPosId.get(location.id)
              const isActive = binding?.is_active ?? false
              return (
                <div key={location.id} className="rounded border border-neutral-200 bg-white px-3 py-3">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <div className="text-sm font-medium text-neutral-900">{location.name}</div>
                      <div className="text-xs text-neutral-500">{location.address}</div>
                    </div>

                    <button
                      type="button"
                      role="switch"
                      aria-checked={isActive}
                      onClick={() =>
                        setDraft((current) => {
                          if (!current) return current
                          const existingBindings = current.bindings ?? []
                          const hasBinding = existingBindings.some((item) => item.pos_id === location.id)
                          return {
                            ...current,
                            bindings: hasBinding
                              ? existingBindings.map((item) => (
                                item.pos_id === location.id
                                  ? { ...item, is_active: !item.is_active }
                                  : item
                              ))
                              : [
                                ...existingBindings,
                                {
                                  menu_id: current.id,
                                  pos_id: location.id,
                                  pos_name: location.name,
                                  is_active: true,
                                },
                              ],
                          }
                        })
                      }
                      className={cn(
                        'relative h-6 w-10 shrink-0 rounded-full transition-colors',
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
            {availablePosLocations.length === 0 && (
              <p className="text-sm text-neutral-500">Для этого бота пока не привязаны точки продаж во вкладке «Подключение».</p>
            )}
          </div>
        </section>

      {showAddCategory && (
        <div className="bg-white rounded border border-neutral-900 p-5 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900 mb-3">Новая категория</h3>
          <div className="flex flex-wrap items-center gap-3">
            <input
              type="text"
              value={newCategory.name}
              onChange={(event) => setNewCategory((current) => ({ ...current, name: event.target.value }))}
              placeholder="Название категории"
              className={cn(inputClassName, 'flex-1 min-w-[240px]')}
            />
            <input
              type="text"
              value={newCategory.icon_emoji}
              onChange={(event) => setNewCategory((current) => ({ ...current, icon_emoji: event.target.value }))}
              placeholder="Эмодзи (например ☕)"
              className={cn(inputClassName, 'w-40')}
            />
            <input
              type="text"
              value={newCategory.icon_image_url}
              onChange={(event) => setNewCategory((current) => ({ ...current, icon_image_url: event.target.value }))}
              placeholder="URL иконки"
              className={cn(inputClassName, 'flex-1 min-w-[200px]')}
            />
            <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded border border-dashed border-neutral-200 px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
              Загрузить иконку
              <input
                type="file"
                accept="image/*"
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
      </>
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
  const [expanded, setExpanded] = useState(false)
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
  const deleteCategory = useDeleteCategoryMutation()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

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
        icon_emoji: draftCategory.icon_emoji,
        icon_image_url: draftCategory.icon_image_url,
      },
    })
    onUpdate()
  }

  const handleCategoryIconUpload = async (file?: File | null) => {
    if (!file) return
    const uploaded = await campaignsApi.uploadFile(file)
    setDraftCategory((current) => ({ ...current, icon_image_url: uploaded }))
  }

  const handleDeleteCategory = async () => {
    await deleteCategory.mutate(category.id)
    onUpdate()
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
        <div className="flex items-center gap-3 min-w-0">
          <span className="flex w-8 h-8 shrink-0 items-center justify-center rounded border border-neutral-200 bg-neutral-50 overflow-hidden">
            <Package className="w-4 h-4 text-neutral-300" />
          </span>
          <h3 className="text-base font-semibold text-neutral-900 truncate">{category.name}</h3>
          <span className="text-xs text-neutral-400 shrink-0">{items.length} позиций</span>
        </div>
        <ChevronDown className={cn('w-4 h-4 text-neutral-400 transition-transform shrink-0', expanded && 'rotate-180')} />
      </button>

      {expanded && (
        <div className="border-t border-neutral-200 space-y-4 p-6">
          <div className="flex flex-wrap items-center gap-3">
            <input
              type="text"
              value={draftCategory.icon_emoji}
              onChange={(event) => setDraftCategory((current) => ({ ...current, icon_emoji: event.target.value }))}
              placeholder="Эмодзи (например ☕)"
              className={cn(inputClassName, 'w-40')}
            />
            <input
              type="text"
              value={draftCategory.icon_image_url}
              onChange={(event) => setDraftCategory((current) => ({ ...current, icon_image_url: event.target.value }))}
              placeholder="URL иконки"
              className={cn(inputClassName, 'flex-1 min-w-[200px]')}
            />
            <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded border border-dashed border-neutral-200 px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
              Загрузить иконку
              <input
                type="file"
                accept="image/*"
                className="hidden"
                onChange={(event) => void handleCategoryIconUpload(event.target.files?.[0])}
              />
            </label>
          </div>
          <div className="flex flex-wrap items-center gap-3">
            <input
              type="text"
              value={draftCategory.name}
              onChange={(event) => setDraftCategory((current) => ({ ...current, name: event.target.value }))}
              placeholder="Название категории"
              className={cn(inputClassName, 'flex-1 min-w-[200px]')}
            />
            <button
              type="button"
              onClick={handleSaveCategory}
              className="inline-flex min-h-11 items-center justify-center rounded bg-neutral-900 px-4 text-sm font-medium text-white hover:bg-neutral-700"
            >
              Сохранить
            </button>
            {showDeleteConfirm ? (
              <div className="flex items-center gap-2 animate-in">
                <button
                  type="button"
                  onClick={handleDeleteCategory}
                  disabled={deleteCategory.isPending}
                  className="inline-flex min-h-11 items-center gap-1.5 rounded bg-red-500 px-3 text-xs font-medium text-white hover:bg-red-600 disabled:opacity-50"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  {items.length > 0 ? `Удалить категорию и ${items.length} позиций` : 'Удалить категорию'}
                </button>
                <button
                  type="button"
                  onClick={() => setShowDeleteConfirm(false)}
                  className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-500 hover:bg-neutral-100"
                >
                  <X className="h-3.5 w-3.5" />
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setShowDeleteConfirm(true)}
                className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-500"
                title="Удалить категорию"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            )}
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
                    type="text"
                    inputMode="decimal"
                    value={newItem.price === 0 ? '' : String(newItem.price)}
                    onChange={(event) => {
                      const raw = event.target.value.replace(/[^\d]/g, '')
                      const num = raw === '' ? 0 : parseInt(raw, 10)
                      setNewItem((current) => ({ ...current, price: num }))
                    }}
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
                <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded border border-dashed border-neutral-200 px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
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
  const deleteItem = useDeleteItemMutation()
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

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
        description: draft.description,
        price: Number(draft.price),
        weight: draft.weight,
        image_url: draft.image_url,
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

  const handleDelete = async () => {
    await deleteItem.mutate(item.id)
    onUpdate()
  }

  return (
    <div className="rounded border border-neutral-200 bg-neutral-50/60 p-4">
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
              type="text"
              inputMode="decimal"
              value={draft.price === 0 ? '' : String(draft.price)}
              onChange={(event) => {
                const raw = event.target.value.replace(/[^\d]/g, '')
                const num = raw === '' ? 0 : parseInt(raw, 10)
                setDraft((current) => ({ ...current, price: num }))
              }}
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
          <label className="inline-flex min-h-11 cursor-pointer items-center justify-center rounded border border-dashed border-neutral-200 px-3 text-sm text-neutral-600 hover:border-accent hover:text-accent">
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
        <div className="flex items-start gap-4">
          {item.image_url ? (
            <div className="h-16 w-16 shrink-0 overflow-hidden rounded-lg bg-neutral-100">
              <img src={item.image_url} alt={item.name} className="h-full w-full object-cover" />
            </div>
          ) : (
            <div className="flex h-16 w-16 shrink-0 items-center justify-center rounded-lg bg-neutral-100 text-neutral-300">
              <Package className="h-6 w-6" />
            </div>
          )}
          <div className="min-w-0 flex-1">
            <div className="flex items-baseline gap-2">
              <span className={cn('text-sm font-medium text-neutral-900', !item.is_available && 'line-through opacity-50')}>
                {item.name}
              </span>
              <span className="text-sm font-semibold text-neutral-900 shrink-0">
                {new Intl.NumberFormat('ru-RU', { style: 'currency', currency: 'RUB', maximumFractionDigits: 0 }).format(item.price)}
              </span>
            </div>
            {item.description && (
              <div className="mt-1 text-xs text-neutral-500 line-clamp-2">{item.description}</div>
            )}
            <div className="mt-2 flex flex-wrap items-center gap-2">
              {item.weight && (
                <span className="rounded bg-neutral-100 px-2 py-0.5 text-xs text-neutral-600">{item.weight}</span>
              )}
              {!item.is_available && (
                <span className="rounded bg-amber-100 px-2 py-0.5 text-xs text-amber-700">Не активно</span>
              )}
            </div>
          </div>
          <div className="flex items-center gap-1">
            {showDeleteConfirm ? (
              <div className="flex items-center gap-2 animate-in">
                <button
                  type="button"
                  onClick={handleDelete}
                  className="inline-flex items-center gap-1 rounded bg-red-500 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-600"
                >
                  <Trash2 className="h-3 w-3" />
                  Удалить
                </button>
                <button
                  type="button"
                  onClick={() => setShowDeleteConfirm(false)}
                  className="inline-flex items-center justify-center rounded px-2 py-1.5 text-xs text-neutral-500 hover:bg-neutral-100"
                >
                  <X className="h-3 w-3" />
                </button>
              </div>
            ) : (
              <button
                type="button"
                onClick={() => setShowDeleteConfirm(true)}
                className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-500"
                title="Удалить позицию"
              >
                <Trash2 className="h-4 w-4" />
              </button>
            )}
            <button
              type="button"
              onClick={() => setEditing(true)}
              className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-white hover:text-neutral-700"
              title="Редактировать позицию"
            >
              <Edit3 className="h-4 w-4" />
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
