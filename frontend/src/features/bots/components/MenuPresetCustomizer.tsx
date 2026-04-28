import { useMemo, useState } from 'react'
import { ChevronDown, Eye, EyeOff, GripVertical, ImageIcon } from 'lucide-react'
import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core'
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { cn } from '@/lib/utils'
import { EmojiPicker } from '@/features/emoji-packs'
import type { EmojiItem } from '@/features/emoji-packs/types'
import type { Menu, MenuItem } from '@/features/menus/types'
import { ButtonStylePicker } from '@/features/telegram-preview/components/ButtonStylePicker'
import { InfoHint } from '@/components/common/InfoHint'
import type {
  MenuPresetCategoryCustomization,
  MenuPresetCustomizations,
  MenuPresetItemCustomization,
  PresetButtonStyle,
} from '../types'

const inputClassName = cn(
  'w-full rounded-lg border border-surface-border bg-white px-4 py-2.5 text-sm',
  'placeholder:text-neutral-400 transition-colors',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
)

interface MenuPresetCustomizerProps {
  menu: Menu | null
  value: MenuPresetCustomizations
  onChange: (value: MenuPresetCustomizations) => void
}

export function createDefaultMenuPresetCustomizations(menu: Menu | null): MenuPresetCustomizations {
  const categories = (menu?.categories ?? []).map((category) => ({
    category_id: category.id,
    label: category.name,
    icon_image_url: category.icon_image_url,
    icon_custom_emoji_id: '',
    style: '' as PresetButtonStyle,
    emoji_only: false,
    item_order: (category.items ?? []).map((item) => item.id),
    items: [] as MenuPresetItemCustomization[],
  }))

  return {
    title: menu?.name ?? '',
    subtitle: '',
    category_order: categories.map((category) => category.category_id),
    categories,
    tab_button_style: '',
    nav_button_style: '',
  }
}

export function normalizeMenuPresetCustomizations(
  raw: Record<string, unknown> | undefined,
  menu: Menu | null,
): MenuPresetCustomizations {
  const defaults = createDefaultMenuPresetCustomizations(menu)
  const rawCategories = Array.isArray(raw?.categories) ? raw?.categories : []
  const rawById = new Map<number, Partial<MenuPresetCategoryCustomization>>()

  for (const item of rawCategories) {
    if (!item || typeof item !== 'object') continue
    const categoryId = Number((item as { category_id?: unknown }).category_id)
    if (!Number.isFinite(categoryId)) continue

    const rawItemOrder = Array.isArray((item as { item_order?: unknown }).item_order)
      ? ((item as { item_order?: unknown[] }).item_order as unknown[])
          .map((v) => Number(v))
          .filter((v) => Number.isFinite(v))
      : undefined

    const rawItems = Array.isArray((item as { items?: unknown }).items)
      ? ((item as { items?: unknown[] }).items as unknown[])
          .filter((v): v is Record<string, unknown> => v != null && typeof v === 'object')
          .map((v) => ({
            item_id: Number(v.item_id),
            label: readString(v.label),
            hidden: Boolean(v.hidden),
          }))
          .filter((v) => Number.isFinite(v.item_id))
      : undefined

    rawById.set(categoryId, {
      category_id: categoryId,
      label: readString((item as { label?: unknown }).label),
      icon_image_url: readString((item as { icon_image_url?: unknown }).icon_image_url),
      icon_custom_emoji_id: readString((item as { icon_custom_emoji_id?: unknown }).icon_custom_emoji_id),
      style: readStyle((item as { style?: unknown }).style),
      emoji_only: Boolean((item as { emoji_only?: unknown }).emoji_only),
      item_order: rawItemOrder,
      items: rawItems,
    })
  }

  const rawOrder = Array.isArray(raw?.category_order)
    ? raw.category_order
        .map((item) => Number(item))
        .filter((item) => Number.isFinite(item))
    : []

  const orderedIds = rawOrder.length > 0 ? rawOrder : defaults.category_order ?? []
  const orderIndex = new Map(orderedIds.map((id, index) => [id, index]))

  const categories = [...(defaults.categories ?? [])]
    .sort((left, right) => {
      const leftIndex = orderIndex.get(left.category_id)
      const rightIndex = orderIndex.get(right.category_id)
      if (leftIndex == null && rightIndex == null) return 0
      if (leftIndex == null) return 1
      if (rightIndex == null) return -1
      return leftIndex - rightIndex
    })
    .map((category) => {
      const override = rawById.get(category.category_id)
      return {
        ...category,
        label: override?.emoji_only ? '' : (override?.label || category.label),
        icon_image_url: override?.icon_image_url || category.icon_image_url || '',
        icon_custom_emoji_id: override?.icon_custom_emoji_id || '',
        style: readStyle(override?.style || category.style || ''),
        emoji_only: Boolean(override?.emoji_only),
        item_order: override?.item_order ?? category.item_order,
        items: override?.items ?? category.items ?? [],
      }
    })

  return {
    title: readString(raw?.title) || defaults.title,
    subtitle: readString(raw?.subtitle) || '',
    category_order: categories.map((category) => category.category_id),
    categories,
    tab_button_style: readStyle(raw?.tab_button_style) || '',
    nav_button_style: readStyle(raw?.nav_button_style) || '',
  }
}

export function isMenuPresetCustomizationDirty(
  value: MenuPresetCustomizations,
  menu: Menu | null,
): boolean {
  return JSON.stringify(sanitizeMenuPresetCustomizations(value, menu)) !== JSON.stringify({})
}

export function sanitizeMenuPresetCustomizations(
  value: MenuPresetCustomizations,
  menu: Menu | null,
): Record<string, unknown> {
  const defaults = createDefaultMenuPresetCustomizations(menu)
  const result: Record<string, unknown> = {}

  if ((value.title ?? '').trim() !== (defaults.title ?? '').trim()) {
    result.title = (value.title ?? '').trim()
  }
  if ((value.subtitle ?? '').trim() !== '') {
    result.subtitle = (value.subtitle ?? '').trim()
  }
  if ((value.tab_button_style ?? '') !== '') {
    result.tab_button_style = value.tab_button_style
  }
  if ((value.nav_button_style ?? '') !== '') {
    result.nav_button_style = value.nav_button_style
  }

  const currentOrder = value.category_order ?? []
  const defaultOrder = defaults.category_order ?? []
  if (JSON.stringify(currentOrder) !== JSON.stringify(defaultOrder)) {
    result.category_order = currentOrder
  }

  const defaultById = new Map(
    (defaults.categories ?? []).map((category) => [category.category_id, category]),
  )
  const categories = (value.categories ?? [])
    .map((category) => {
      const fallback = defaultById.get(category.category_id)
      const entry: Record<string, unknown> = {
        category_id: category.category_id,
      }

      if ((category.label ?? '').trim() !== (fallback?.label ?? '').trim()) {
        entry.label = (category.label ?? '').trim()
      }
      if ((category.icon_image_url ?? '') !== (fallback?.icon_image_url ?? '')) {
        entry.icon_image_url = category.icon_image_url ?? ''
      }
      if ((category.icon_custom_emoji_id ?? '') !== (fallback?.icon_custom_emoji_id ?? '')) {
        entry.icon_custom_emoji_id = category.icon_custom_emoji_id ?? ''
      }
      if ((category.style ?? '') !== (fallback?.style ?? '')) {
        entry.style = category.style ?? ''
      }
      if (Boolean(category.emoji_only)) {
        entry.emoji_only = true
      }

      const defaultItemOrder = fallback?.item_order ?? []
      const currentItemOrder = category.item_order ?? []
      if (JSON.stringify(currentItemOrder) !== JSON.stringify(defaultItemOrder)) {
        entry.item_order = currentItemOrder
      }

      const dirtyItems = (category.items ?? [])
        .filter((item) => item.label?.trim() || item.hidden)
        .map((item) => {
          const cleaned: Record<string, unknown> = { item_id: item.item_id }
          if (item.label?.trim()) cleaned.label = item.label.trim()
          if (item.hidden) cleaned.hidden = true
          return cleaned
        })
      if (dirtyItems.length > 0) {
        entry.items = dirtyItems
      }

      return Object.keys(entry).length > 1 ? entry : null
    })
    .filter(Boolean)

  if (categories.length > 0) {
    result.categories = categories
  }

  return result
}

export function MenuPresetCustomizer({
  menu,
  value,
  onChange,
}: MenuPresetCustomizerProps) {
  const categories = value.categories ?? []
  const [expandedCategoryIds, setExpandedCategoryIds] = useState<number[]>([])
  const categorySensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )
  const sortableIds = useMemo(
    () => categories.map((category) => `category-${category.category_id}`),
    [categories],
  )

  if (!menu || categories.length === 0) {
    return (
      <div className="rounded-xl border border-dashed border-surface-border bg-neutral-50 p-4 text-sm text-neutral-500">
        Сначала привяжите к боту меню с категориями. После этого здесь появится кастомизация табов и превью.
      </div>
    )
  }

  const update = (patch: Partial<MenuPresetCustomizations>) => {
    onChange({
      ...value,
      ...patch,
    })
  }

  const updateCategory = (
    categoryId: number,
    updater: (category: MenuPresetCategoryCustomization) => MenuPresetCategoryCustomization,
  ) => {
    const next = categories.map((category) =>
      category.category_id === categoryId ? updater(category) : category,
    )
    update({
      categories: next,
      category_order: next.map((category) => category.category_id),
    })
  }

  const toggleExpanded = (categoryId: number) => {
    setExpandedCategoryIds((current) =>
      current.includes(categoryId)
        ? current.filter((id) => id !== categoryId)
        : [...current, categoryId],
    )
  }

  const handleCategoryDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = sortableIds.indexOf(String(active.id))
    const newIndex = sortableIds.indexOf(String(over.id))
    if (oldIndex < 0 || newIndex < 0) return

    const next = arrayMove(categories, oldIndex, newIndex)
    update({
      categories: next,
      category_order: next.map((category) => category.category_id),
    })
  }

  return (
    <div className="space-y-8">
      <section className="grid gap-4 xl:grid-cols-[minmax(0,1.15fr)_minmax(18rem,0.85fr)]">
        <div className="rounded-2xl border border-surface-border bg-neutral-50/70 p-5">
          <div className="mb-4">
            <div className="flex items-center gap-2">
              <h4 className="text-sm font-semibold text-neutral-900">Тексты блока</h4>
              <InfoHint content="Здесь меняются заголовок и короткое пояснение, которое пользователь увидит над категориями в боте." />
            </div>
          </div>
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-2">
              <span className="text-sm font-medium text-neutral-700">Заголовок</span>
              <input
                value={value.title ?? ''}
                onChange={(event) => update({ title: event.target.value })}
                className={inputClassName}
                placeholder="Меню"
              />
            </label>
            <label className="space-y-2">
              <span className="text-sm font-medium text-neutral-700">Подзаголовок</span>
              <input
                value={value.subtitle ?? ''}
                onChange={(event) => update({ subtitle: event.target.value })}
                className={inputClassName}
                placeholder="Короткое пояснение над категориями"
              />
            </label>
          </div>
        </div>

        <div className="rounded-2xl border border-surface-border bg-neutral-50/70 p-5">
          <div className="mb-4">
            <div className="flex items-center gap-2">
              <h4 className="text-sm font-semibold text-neutral-900">Глобальный стиль кнопок</h4>
              <InfoHint content="Общий стиль для табов и навигации. Если нужно, ниже можно задать отдельный стиль только для конкретной категории." />
            </div>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between gap-4 rounded-xl border border-surface-border bg-white px-4 py-3">
              <div>
                <div className="text-sm font-medium text-neutral-800">Стиль табов</div>
                <div className="mt-1 text-sm text-neutral-500">Цвет для кнопок категорий</div>
              </div>
              <ButtonStylePicker
                value={value.tab_button_style ?? ''}
                onChange={(style) => update({ tab_button_style: style })}
              />
            </div>

            <div className="flex items-center justify-between gap-4 rounded-xl border border-surface-border bg-white px-4 py-3">
              <div>
                <div className="text-sm font-medium text-neutral-800">Стиль навигации</div>
                <div className="mt-1 text-sm text-neutral-500">Стрелки и счётчик карточки блюда</div>
              </div>
              <ButtonStylePicker
                value={value.nav_button_style ?? ''}
                onChange={(style) => update({ nav_button_style: style })}
              />
            </div>
          </div>
        </div>
      </section>

      <section className="space-y-4">
        <div className="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <h4 className="text-sm font-semibold text-neutral-900">Категории и позиции</h4>
              <InfoHint content="Перетаскивайте строки, чтобы менять порядок. Откройте категорию, чтобы настроить иконку, цвет и управлять отображением позиций." />
            </div>
            <p className="mt-1 text-sm text-neutral-500">
              Перетаскивайте категории за ручку слева. Внутри — настройки отображения и позиции.
            </p>
          </div>
          <div className="text-sm text-neutral-400">{categories.length} категорий</div>
        </div>

        <DndContext
          sensors={categorySensors}
          collisionDetection={closestCenter}
          onDragEnd={handleCategoryDragEnd}
        >
          <SortableContext items={sortableIds} strategy={verticalListSortingStrategy}>
            <div className="space-y-3">
              {categories.map((category) => {
                const menuCategory = menu.categories?.find((item) => item.id === category.category_id)
                return (
                  <SortableCategoryRow
                    key={category.category_id}
                    id={`category-${category.category_id}`}
                    category={category}
                    menuCategory={menuCategory}
                    isExpanded={expandedCategoryIds.includes(category.category_id)}
                    onToggleExpanded={() => toggleExpanded(category.category_id)}
                    onUpdateCategory={updateCategory}
                  />
                )
              })}
            </div>
          </SortableContext>
        </DndContext>
      </section>
    </div>
  )
}

interface SortableCategoryRowProps {
  id: string
  category: MenuPresetCategoryCustomization
  menuCategory?: NonNullable<Menu['categories']>[number]
  isExpanded: boolean
  onToggleExpanded: () => void
  onUpdateCategory: (
    categoryId: number,
    updater: (category: MenuPresetCategoryCustomization) => MenuPresetCategoryCustomization,
  ) => void
}

function SortableCategoryRow({
  id,
  category,
  menuCategory,
  isExpanded,
  onToggleExpanded,
  onUpdateCategory,
}: SortableCategoryRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }
  const canonicalName = menuCategory?.name || 'Без названия'
  const canonicalIcon = menuCategory?.icon_image_url || ''
  const hasLabelOverride = Boolean(category.label?.trim()) && category.label?.trim() !== menuCategory?.name?.trim()
  const hasIconOverride = Boolean(category.icon_image_url) && category.icon_image_url !== canonicalIcon
  const displayIcon = category.icon_image_url || canonicalIcon

  const menuItems = menuCategory?.items ?? []
  const itemOrder = category.item_order ?? menuItems.map((item) => item.id)
  const itemCustomizations = category.items ?? []
  const itemCustomById = new Map(itemCustomizations.map((item) => [item.item_id, item]))

  const orderedItems = useMemo(() => {
    const orderIndex = new Map(itemOrder.map((id, index) => [id, index]))
    return [...menuItems].sort((a, b) => {
      const ai = orderIndex.get(a.id)
      const bi = orderIndex.get(b.id)
      if (ai == null && bi == null) return 0
      if (ai == null) return 1
      if (bi == null) return -1
      return ai - bi
    })
  }, [menuItems, itemOrder])

  const itemSensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates }),
  )
  const itemSortableIds = useMemo(
    () => orderedItems.map((item) => `item-${item.id}`),
    [orderedItems],
  )

  const handleItemDragEnd = (event: DragEndEvent) => {
    const { active, over } = event
    if (!over || active.id === over.id) return

    const oldIndex = itemSortableIds.indexOf(String(active.id))
    const newIndex = itemSortableIds.indexOf(String(over.id))
    if (oldIndex < 0 || newIndex < 0) return

    const reordered = arrayMove(orderedItems, oldIndex, newIndex)
    onUpdateCategory(category.category_id, (current) => ({
      ...current,
      item_order: reordered.map((item) => item.id),
    }))
  }

  const updateItem = (itemId: number, patch: Partial<MenuPresetItemCustomization>) => {
    onUpdateCategory(category.category_id, (current) => {
      const existing = (current.items ?? []).find((i) => i.item_id === itemId)
      const updated: MenuPresetItemCustomization = {
        item_id: itemId,
        label: patch.label ?? existing?.label,
        hidden: patch.hidden ?? existing?.hidden,
      }
      const items = existing
        ? (current.items ?? []).map((i) => (i.item_id === itemId ? updated : i))
        : [...(current.items ?? []), updated]
      return { ...current, items }
    })
  }

  const hiddenCount = orderedItems.filter((item) => itemCustomById.get(item.id)?.hidden).length
  const itemCount = menuItems.length

  return (
    <div ref={setNodeRef} style={style}>
      <div
        className={cn(
          'overflow-hidden rounded-2xl border border-surface-border bg-white',
          isDragging && 'shadow-lg ring-1 ring-accent/20',
        )}
      >
        <div className="flex items-center gap-3 px-4 py-3">
          <button
            type="button"
            className="inline-flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-neutral-400 transition-colors hover:bg-neutral-100 hover:text-neutral-600 cursor-grab active:cursor-grabbing touch-none"
            aria-label={`Перетащить категорию ${canonicalName}`}
            {...listeners}
            {...attributes}
          >
            <GripVertical className="h-4 w-4" />
          </button>

          <button
            type="button"
            onClick={onToggleExpanded}
            className="flex min-w-0 flex-1 items-center gap-3 text-left"
          >
            <span className="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-neutral-50/70">
              {displayIcon ? (
                <img
                  src={displayIcon}
                  alt={canonicalName}
                  className="h-full w-full object-cover"
                />
              ) : (
                <span className="text-xs font-medium text-neutral-300" aria-hidden="true">—</span>
              )}
            </span>

            <div className="min-w-0 flex-1">
              <div className="truncate text-sm font-medium text-neutral-900">{canonicalName}</div>
              <div className="flex items-center gap-2 text-xs text-neutral-400">
                <span>{itemCount} позиций</span>
                {hiddenCount > 0 && (
                  <span className="text-amber-600">{hiddenCount} скрыто</span>
                )}
              </div>
            </div>

            <span
              className={cn(
                'h-3 w-3 shrink-0 rounded-full border border-white shadow-sm',
                styleIndicatorClass(category.style ?? ''),
              )}
              aria-hidden="true"
            />
            <ChevronDown
              className={cn(
                'h-4 w-4 shrink-0 text-neutral-400 transition-transform',
                isExpanded && 'rotate-180',
              )}
            />
          </button>
        </div>

        {isExpanded && (
          <div className="border-t border-surface-border bg-neutral-50/40 px-5 py-4 space-y-4">
            <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_16.5rem] lg:items-start">
              <div className="space-y-2">
                <label className="space-y-2">
                  <span className="text-sm font-medium text-neutral-700">Название в боте</span>
                  <input
                    value={category.label ?? ''}
                    onChange={(event) =>
                      onUpdateCategory(category.category_id, (current) => ({
                        ...current,
                        label: event.target.value,
                        emoji_only:
                          !event.target.value.trim() &&
                          Boolean(current.icon_custom_emoji_id || current.icon_image_url),
                      }))
                    }
                    className={cn(inputClassName, 'text-base')}
                    placeholder={menuCategory?.name ?? 'Категория'}
                  />
                </label>
                {hasLabelOverride ? (
                  <div className="flex items-center gap-2 rounded-lg bg-amber-50 border border-amber-100 px-3 py-1.5 text-xs text-amber-700">
                    <span>Оригинал: <span className="font-medium">{menuCategory?.name}</span></span>
                    <button
                      type="button"
                      onClick={() =>
                        onUpdateCategory(category.category_id, (current) => ({
                          ...current,
                          label: menuCategory?.name ?? '',
                        }))
                      }
                      className="ml-auto text-xs font-medium text-amber-600 hover:text-amber-800"
                    >
                      Сбросить
                    </button>
                  </div>
                ) : (
                  <div className="text-xs text-neutral-400">
                    Используется название из меню. Если оставить пустым — в табе останется только иконка.
                  </div>
                )}
              </div>

              <div className="rounded-xl border border-surface-border bg-white px-4 py-3">
                <div className="space-y-3">
                  <div>
                    <div className="mb-2 text-sm font-medium text-neutral-800">Иконка в боте</div>
                    <div className="flex min-h-10 items-center gap-2">
                      <EmojiPicker
                        selected={category.icon_image_url}
                        onSelect={(item: EmojiItem) =>
                          onUpdateCategory(category.category_id, (current) => ({
                            ...current,
                            icon_image_url: item.image_url,
                            icon_custom_emoji_id: item.tg_custom_emoji_id ?? '',
                            emoji_only: !current.label?.trim(),
                          }))
                        }
                      >
                        {category.icon_image_url ? (
                          <img
                            src={category.icon_image_url}
                            alt="Иконка категории в боте"
                            className="h-8 w-8 rounded object-cover"
                          />
                        ) : (
                          <div className="flex h-8 w-8 items-center justify-center rounded bg-neutral-50 text-neutral-400">
                            <ImageIcon className="h-4 w-4 text-neutral-400" />
                          </div>
                        )}
                      </EmojiPicker>
                      {!category.icon_image_url && !canonicalIcon && (
                        <div className="text-sm text-neutral-500">Добавить иконку</div>
                      )}
                      {!category.icon_image_url && canonicalIcon && (
                        <div className="text-sm text-neutral-400">Используется иконка из меню</div>
                      )}
                      {category.icon_image_url && (
                        <button
                          type="button"
                          onClick={() =>
                            onUpdateCategory(category.category_id, (current) => ({
                              ...current,
                              icon_image_url: '',
                              icon_custom_emoji_id: '',
                              emoji_only: false,
                            }))
                          }
                          className="ml-auto text-xs font-medium text-neutral-500 hover:text-neutral-700"
                        >
                          Убрать
                        </button>
                      )}
                    </div>
                    {hasIconOverride && canonicalIcon && (
                      <div className="mt-2 flex items-center gap-2 rounded-lg bg-amber-50 border border-amber-100 px-3 py-1.5 text-xs text-amber-700">
                        <span className="flex items-center gap-1.5">
                          В меню: <img src={canonicalIcon} alt="" className="h-5 w-5 rounded object-cover inline" />
                        </span>
                        <button
                          type="button"
                          onClick={() =>
                            onUpdateCategory(category.category_id, (current) => ({
                              ...current,
                              icon_image_url: canonicalIcon,
                              icon_custom_emoji_id: '',
                            }))
                          }
                          className="ml-auto text-xs font-medium text-amber-600 hover:text-amber-800"
                        >
                          Сбросить
                        </button>
                      </div>
                    )}
                  </div>

                  <div className="border-t border-surface-border pt-3">
                    <div className="mb-2 text-sm font-medium text-neutral-800">Цвет кнопки</div>
                    <div className="flex min-h-10 items-center gap-3">
                      <ButtonStylePicker
                        value={category.style ?? ''}
                        onChange={(style) =>
                          onUpdateCategory(category.category_id, (current) => ({
                            ...current,
                            style,
                          }))
                        }
                      />
                      <span className="text-sm text-neutral-500">Только для этой категории</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            {orderedItems.length > 0 && (
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <div className="text-sm font-medium text-neutral-700">Позиции</div>
                  <div className="text-xs text-neutral-400">
                    {orderedItems.length - hiddenCount} из {orderedItems.length} показано
                  </div>
                </div>
                <DndContext
                  sensors={itemSensors}
                  collisionDetection={closestCenter}
                  onDragEnd={handleItemDragEnd}
                >
                  <SortableContext items={itemSortableIds} strategy={verticalListSortingStrategy}>
                    <div className="space-y-1">
                      {orderedItems.map((menuItem) => {
                        const custom = itemCustomById.get(menuItem.id)
                        return (
                          <SortableItemRow
                            key={menuItem.id}
                            id={`item-${menuItem.id}`}
                            menuItem={menuItem}
                            customization={custom}
                            onUpdate={(patch) => updateItem(menuItem.id, patch)}
                          />
                        )
                      })}
                    </div>
                  </SortableContext>
                </DndContext>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}

interface SortableItemRowProps {
  id: string
  menuItem: MenuItem
  customization?: MenuPresetItemCustomization
  onUpdate: (patch: Partial<MenuPresetItemCustomization>) => void
}

function SortableItemRow({ id, menuItem, customization, onUpdate }: SortableItemRowProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id })
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  }

  const isHidden = customization?.hidden ?? false
  const hasLabelOverride = Boolean(customization?.label?.trim()) && customization?.label?.trim() !== menuItem.name.trim()
  const [editing, setEditing] = useState(false)

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={cn(
        'flex items-center gap-2 rounded-xl border border-surface-border bg-white px-3 py-2',
        isDragging && 'shadow-md ring-1 ring-accent/20',
        isHidden && 'opacity-50',
      )}
    >
      <button
        type="button"
        className="inline-flex h-7 w-7 shrink-0 items-center justify-center rounded text-neutral-300 hover:text-neutral-500 cursor-grab active:cursor-grabbing touch-none"
        {...listeners}
        {...attributes}
      >
        <GripVertical className="h-3.5 w-3.5" />
      </button>

      <div className="min-w-0 flex-1">
        {editing ? (
          <input
            value={customization?.label ?? menuItem.name}
            onChange={(e) => onUpdate({ label: e.target.value })}
            onBlur={() => setEditing(false)}
            onKeyDown={(e) => e.key === 'Enter' && setEditing(false)}
            className="w-full rounded border border-surface-border bg-white px-2 py-1 text-sm focus:outline-none focus:ring-1 focus:ring-accent/30"
            placeholder={menuItem.name}
            autoFocus
          />
        ) : (
          <button
            type="button"
            onClick={() => setEditing(true)}
            className="text-left w-full"
          >
            <div className={cn('text-sm text-neutral-900', isHidden && 'line-through')}>
              {hasLabelOverride ? customization!.label : menuItem.name}
            </div>
            <div className="flex items-center gap-2 text-xs text-neutral-400">
              <span>{formatPrice(menuItem.price)}</span>
              {menuItem.weight && <span>{menuItem.weight}</span>}
              {hasLabelOverride && (
                <span className="text-blue-500">оригинал: {menuItem.name}</span>
              )}
            </div>
          </button>
        )}
      </div>

      {hasLabelOverride && !editing && (
        <button
          type="button"
          onClick={() => onUpdate({ label: '' })}
          className="text-xs text-amber-600 hover:text-amber-800 shrink-0"
          title="Сбросить название"
        >
          Сбросить
        </button>
      )}

      <button
        type="button"
        onClick={() => onUpdate({ hidden: !isHidden })}
        className={cn(
          'inline-flex h-7 w-7 shrink-0 items-center justify-center rounded transition-colors',
          isHidden
            ? 'text-neutral-400 hover:text-neutral-600'
            : 'text-neutral-400 hover:text-neutral-600',
        )}
        title={isHidden ? 'Показать в боте' : 'Скрыть в боте'}
      >
        {isHidden ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
      </button>
    </div>
  )
}

function formatPrice(price: number): string {
  return new Intl.NumberFormat('ru-RU', {
    style: 'currency',
    currency: 'RUB',
    maximumFractionDigits: 0,
  }).format(price)
}

function readString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function readStyle(value: unknown): PresetButtonStyle {
  if (value === 'primary' || value === 'success' || value === 'danger') return value
  return ''
}

function styleIndicatorClass(style: PresetButtonStyle): string {
  if (style === 'primary') return 'bg-blue-500'
  if (style === 'success') return 'bg-green-500'
  if (style === 'danger') return 'bg-red-500'
  return 'bg-neutral-300'
}
