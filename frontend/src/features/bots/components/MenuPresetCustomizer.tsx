import { ArrowDown, ArrowUp, ImageIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { EmojiPicker } from '@/features/emoji-packs'
import type { EmojiItem } from '@/features/emoji-packs/types'
import type { Menu } from '@/features/menus/types'
import { ButtonStylePicker } from '@/features/telegram-preview/components/ButtonStylePicker'
import { InfoHint } from '@/components/common/InfoHint'
import type {
  MenuPresetCategoryCustomization,
  MenuPresetCustomizations,
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
    rawById.set(categoryId, {
      category_id: categoryId,
      label: readString((item as { label?: unknown }).label),
      icon_image_url: readString((item as { icon_image_url?: unknown }).icon_image_url),
      icon_custom_emoji_id: readString((item as { icon_custom_emoji_id?: unknown }).icon_custom_emoji_id),
      style: readStyle((item as { style?: unknown }).style),
      emoji_only: Boolean((item as { emoji_only?: unknown }).emoji_only),
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

  const moveCategory = (categoryId: number, direction: -1 | 1) => {
    const index = categories.findIndex((category) => category.category_id === categoryId)
    const nextIndex = index + direction
    if (index < 0 || nextIndex < 0 || nextIndex >= categories.length) return
    const next = [...categories]
    const [item] = next.splice(index, 1)
    next.splice(nextIndex, 0, item)
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
              <h4 className="text-sm font-semibold text-neutral-900">Категории</h4>
              <InfoHint content="Здесь можно менять порядок, подписи и иконки категорий. Логика переходов внутри выбранного шаблона не меняется." />
            </div>
          </div>
          <div className="text-sm text-neutral-400">{categories.length} категорий</div>
        </div>

        <div className="hidden rounded-xl border border-surface-border bg-neutral-50 px-4 py-3 xl:grid xl:grid-cols-[5.5rem_minmax(0,1.4fr)_12rem_12rem] xl:gap-4">
          <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Порядок</div>
          <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Подпись</div>
          <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Иконка</div>
          <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Стиль</div>
        </div>

        {categories.map((category, index) => {
          const menuCategory = menu.categories?.find((item) => item.id === category.category_id)
          return (
            <div
              key={category.category_id}
              className="rounded-2xl border border-surface-border bg-white p-4"
            >
              <div className="grid gap-4 xl:grid-cols-[5.5rem_minmax(0,1.4fr)_12rem_12rem] xl:items-start">
                <div className="flex items-center gap-2 xl:pt-8">
                  <button
                    type="button"
                    onClick={() => moveCategory(category.category_id, -1)}
                    disabled={index === 0}
                    className="rounded-lg border border-surface-border p-2 text-neutral-500 disabled:opacity-40"
                    aria-label="Переместить вверх"
                  >
                    <ArrowUp className="h-4 w-4" />
                  </button>
                  <button
                    type="button"
                    onClick={() => moveCategory(category.category_id, 1)}
                    disabled={index === categories.length - 1}
                    className="rounded-lg border border-surface-border p-2 text-neutral-500 disabled:opacity-40"
                    aria-label="Переместить вниз"
                  >
                    <ArrowDown className="h-4 w-4" />
                  </button>
                </div>

                <label className="space-y-2">
                  <span className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 xl:hidden">
                    Подпись
                  </span>
                  <input
                    value={category.label ?? ''}
                    onChange={(event) =>
                      updateCategory(category.category_id, (current) => ({
                        ...current,
                        label: event.target.value,
                        emoji_only:
                          !event.target.value.trim() &&
                          Boolean(current.icon_custom_emoji_id || current.icon_image_url),
                      }))
                    }
                    className={inputClassName}
                    placeholder={menuCategory?.name ?? 'Категория'}
                  />
                  <div className="flex flex-wrap items-center gap-3 text-xs text-neutral-400">
                    {menuCategory?.name && menuCategory.name !== category.label && (
                      <span>Исходное название: {menuCategory.name}</span>
                    )}
                    <span>Если оставить подпись пустой, в табе останется только иконка.</span>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-neutral-500">
                    <span>Превью:</span>
                    <span className="inline-flex min-h-8 items-center gap-2 rounded-full border border-surface-border bg-neutral-50 px-3 py-1">
                      {category.icon_image_url ? (
                        <img
                          src={category.icon_image_url}
                          alt={category.label || menuCategory?.name || 'Категория'}
                          className="h-4 w-4 rounded object-cover"
                        />
                      ) : null}
                      {(category.label || '').trim() && <span>{category.label}</span>}
                      {!(category.label || '').trim() && !category.icon_image_url && <span>{menuCategory?.name || 'Категория'}</span>}
                    </span>
                  </div>
                </label>

                <div className="space-y-2">
                  <span className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 xl:hidden">
                    Иконка
                  </span>
                  <div className="flex min-h-12 items-center gap-2 rounded-xl border border-surface-border bg-neutral-50 px-3 py-2">
                    <EmojiPicker
                      selected={category.icon_image_url}
                      onSelect={(item: EmojiItem) =>
                        updateCategory(category.category_id, (current) => ({
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
                          alt="Эмодзи категории"
                          className="h-6 w-6 rounded object-cover"
                        />
                      ) : (
                        <ImageIcon className="h-4 w-4 text-neutral-500" />
                      )}
                    </EmojiPicker>
                    <div className="text-sm text-neutral-500">
                      {category.icon_image_url ? 'Иконка выбрана' : 'Добавить иконку'}
                    </div>
                    {category.icon_image_url && (
                      <button
                        type="button"
                        onClick={() =>
                          updateCategory(category.category_id, (current) => ({
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
                </div>

                <div className="space-y-2">
                  <span className="text-xs font-medium uppercase tracking-[0.18em] text-neutral-400 xl:hidden">
                    Стиль
                  </span>
                  <div className="flex min-h-12 items-center gap-3 rounded-xl border border-surface-border bg-neutral-50 px-3 py-2">
                    <ButtonStylePicker
                      value={category.style ?? ''}
                      onChange={(style) =>
                        updateCategory(category.category_id, (current) => ({
                          ...current,
                          style,
                        }))
                      }
                    />
                    <span className="text-sm text-neutral-500">Только для этой кнопки</span>
                  </div>
                </div>
              </div>
            </div>
          )
        })}
      </section>
    </div>
  )
}

function readString(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function readStyle(value: unknown): PresetButtonStyle {
  if (value === 'primary' || value === 'success' || value === 'danger') return value
  return ''
}
