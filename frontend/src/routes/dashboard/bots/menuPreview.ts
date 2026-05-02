import type { InlineButton, MessageContent } from '@/features/telegram-preview'
import type { Menu } from '@/features/menus/types'
import type { MenuPresetCustomizations } from '@/features/bots/types'

export interface MenuPreviewItem {
  id: number
  label: string
  price: number
  weight?: string
  description?: string
  imageUrl?: string
  hidden?: boolean
}

export interface MenuPreviewCategory {
  id: number
  label: string
  icon: string
  emojiOnly: boolean
  style?: InlineButton['style']
  items: MenuPreviewItem[]
}

export interface MenuPreviewState {
  title: string
  subtitle: string
  intro: string
  categories: MenuPreviewCategory[]
  isListPreset: boolean
}

export function buildMenuPreviewState(args: {
  introContent: MessageContent | null
  menu: Menu | null
  draft: MenuPresetCustomizations
  selectedPresetKey: string
}): MenuPreviewState {
  const intro = args.introContent?.parts?.[0]?.text?.trim() || 'Добро пожаловать! 🍽️'
  const title = args.draft.title?.trim() || args.menu?.name?.trim() || ''
  const subtitle = args.draft.subtitle?.trim() || ''
  const menuCategories = [...(args.menu?.categories ?? [])].sort(
    (left, right) => left.sort_order - right.sort_order,
  )
  const categoryDrafts = args.draft.categories ?? []
  const categoryOrder = args.draft.category_order?.length
    ? args.draft.category_order
    : menuCategories.map((category) => category.id)

  const categories: MenuPreviewCategory[] = categoryOrder
    .map((categoryId) => {
      const categoryDraft = categoryDrafts.find((category) => category.category_id === categoryId)
      const menuCategory = menuCategories.find((category) => category.id === categoryId)
      if (!menuCategory) return null

      const itemOverrides = new Map(
        (categoryDraft?.items ?? []).map((item) => [item.item_id, item]),
      )
      const baseItems = [...(menuCategory.items ?? [])].sort((left, right) => left.sort_order - right.sort_order)
      const itemOrder = categoryDraft?.item_order?.length
        ? categoryDraft.item_order
        : baseItems.map((item) => item.id)
      const itemById = new Map(baseItems.map((item) => [item.id, item]))

      const orderedItems = itemOrder
        .map((itemId) => itemById.get(itemId))
        .filter((item): item is NonNullable<typeof item> => Boolean(item))
        .map((item) => {
          const override = itemOverrides.get(item.id)
          return {
            id: item.id,
            label: override?.label?.trim() || item.name,
            price: item.price,
            weight: item.weight ?? undefined,
            description: item.description ?? undefined,
            imageUrl: item.image_url ?? undefined,
            hidden: Boolean(override?.hidden) || !item.is_available,
          }
        })
        .filter((item) => !item.hidden)

      return {
        id: categoryId,
        label: categoryDraft?.label?.trim() || menuCategory.name,
        icon: categoryDraft?.icon_image_url || menuCategory.icon_image_url || '',
        emojiOnly: Boolean(categoryDraft?.emoji_only && (categoryDraft?.icon_image_url || menuCategory.icon_image_url)),
        style: (categoryDraft?.style as InlineButton['style']) || undefined,
        items: orderedItems,
      } satisfies MenuPreviewCategory
    })
    .filter((category): category is MenuPreviewCategory => category !== null)

  return { title, subtitle, intro, categories, isListPreset: args.selectedPresetKey === 'list' }
}

export function buildMenuInitialContent(state: MenuPreviewState): MessageContent {
  const selected = state.categories[0]
  if (!selected) {
    return {
      parts: [{ type: 'text' as const, text: buildTabsHeaderText(state) }],
    }
  }

  const header = buildTabsHeaderText(state)
  const text = truncateMenuText([header, renderMenuASCIIBlock(selected.items)].filter(Boolean).join('\n'))

  return {
    parts: [{ type: 'text' as const, text }],
    buttons: [
      ...buildCategoryTabRows(state.categories, selected.id),
      ...buildCategoryItemRows(selected.id, selected.items),
    ],
  }
}

export function buildMenuListContent(state: MenuPreviewState): MessageContent {
  const sections = [buildListHeaderText(state)]

  for (const category of state.categories) {
    sections.push(menuSections(categoryHeading(category), menuCategoryItemsText(category.items)))
  }

  return {
    parts: [{ type: 'text' as const, text: truncateMenuText(menuSections(...sections)) }],
    buttons: buildMenuListCategoryRows(state.categories),
  }
}

export function buildCategoryContent(category: MenuPreviewCategory): MessageContent {
  const buttons: InlineButton[][] = category.items.map((item) => [
    { text: item.label, data: `menu:item:${item.id}:${category.id}:0` },
  ])
  buttons.push([{ text: 'Назад к меню', data: 'menu:root' }])

  return {
    parts: [{ type: 'text' as const, text: `${categoryHeading(category)}\n\nВыберите позицию:` }],
    buttons,
  }
}

export function buildItemCardContent(category: MenuPreviewCategory, itemID: number): MessageContent | null {
  const index = category.items.findIndex((item) => item.id === itemID)
  if (index === -1) return null

  const item = category.items[index]
  const sections = [
    categoryHeading(category),
    `${item.label} — ${formatMenuPrice(item.price)}`,
    item.weight?.trim() || '',
    item.description?.trim() || '',
  ]

  const buttons: InlineButton[][] = []
  const nav: InlineButton[] = []

  if (index > 0) {
    nav.push({ text: '←', data: `menu:cardnav:${category.items[index - 1].id}:${category.id}` })
  }
  nav.push({ text: `${index + 1}/${category.items.length}`, data: 'menu:noop' })
  if (index < category.items.length - 1) {
    nav.push({ text: '→', data: `menu:cardnav:${category.items[index + 1].id}:${category.id}` })
  }
  if (nav.length > 0) buttons.push(nav)
  buttons.push([{ text: 'Назад к категории', data: `menu:tab:${category.id}` }])
  buttons.push([{ text: '✕ Закрыть', data: 'menu:cardclose' }])

  const text = truncateMenuText(menuSections(...sections))

  if (item.imageUrl) {
    return {
      parts: [{ type: 'photo', media_url: item.imageUrl, text, parse_mode: 'HTML' }],
      buttons,
    }
  }

  return {
    parts: [{ type: 'text' as const, text, parse_mode: 'HTML' }],
    buttons,
  }
}

export function buildMenuTabsContent(state: MenuPreviewState, selectedCategoryID: number): MessageContent {
  const selected = state.categories.find((category) => category.id === selectedCategoryID) ?? state.categories[0]
  if (!selected) return buildMenuInitialContent(state)

  return {
    parts: [{ type: 'text' as const, text: truncateMenuText([buildTabsHeaderText(state), renderMenuASCIIBlock(selected.items)].filter(Boolean).join('\n')) }],
    buttons: [
      ...buildCategoryTabRows(state.categories, selected.id),
      ...buildCategoryItemRows(selected.id, selected.items),
    ],
  }
}

function buildTabsHeaderText(state: MenuPreviewState): string {
  return [state.title, state.subtitle].filter(Boolean).join('\n')
}

function buildListHeaderText(state: MenuPreviewState): string {
  const inferredSubtitle = !state.subtitle &&
    state.intro &&
    state.intro !== 'Выберите категорию:' &&
    state.intro !== state.title
    ? state.intro
    : ''

  return [state.title, state.subtitle || inferredSubtitle].filter(Boolean).join('\n')
}

function buildCategoryTabRows(categories: MenuPreviewCategory[], activeCategoryID: number): InlineButton[][] {
  const rows: InlineButton[][] = []
  let current: InlineButton[] = []

  for (const category of categories) {
    current.push({
      text: category.emojiOnly && category.icon ? '⠀' : category.label,
      data: `menu:tab:${category.id}`,
      style: category.id === activeCategoryID ? 'success' : category.style || '',
      icon_image_url: category.icon || undefined,
    })

    if (current.length === 4) {
      rows.push(current)
      current = []
    }
  }

  if (current.length > 0) rows.push(current)
  return rows
}

function buildCategoryItemRows(categoryID: number, items: MenuPreviewItem[]): InlineButton[][] {
  return items.map((item) => [{ text: item.label, data: `menu:item:${item.id}:${categoryID}:0` }])
}

function buildMenuListCategoryRows(categories: MenuPreviewCategory[]): InlineButton[][] {
  return categories.map((category) => [
    {
      text: category.emojiOnly && category.icon ? '⠀' : category.label,
      data: `menu:cat:${category.id}:0`,
      style: category.style || '',
      icon_image_url: category.icon || undefined,
    },
  ])
}

function menuCategoryItemsText(items: MenuPreviewItem[]): string {
  if (items.length === 0) return 'Сейчас в этой категории нет доступных позиций.'

  const lines: string[] = []
  for (const item of items) {
    let line = `${item.label} — ${formatMenuPrice(item.price)}`
    if (item.weight?.trim()) {
      line += ` • ${item.weight.trim()}`
    }
    lines.push(line)
    if (item.description?.trim()) lines.push(item.description.trim())
  }
  return lines.join('\n')
}

function renderMenuASCIIBlock(items: MenuPreviewItem[]): string {
  const lines = [
    '╔═.·:·. ﹏﹏𓂁🦈𓂁﹏﹏.·:·.═╗',
    '║',
  ]

  for (const item of items) {
    lines.push(`║ ⟼ ${item.label}`)
  }

  lines.push('║', '╚═.·:·. ﹏﹏𓂁🪸𓂁﹏﹏.·:·.═╝')
  return lines.join('\n')
}

function categoryHeading(category: MenuPreviewCategory): string {
  return category.icon ? `${category.icon} ${category.label}`.trim() : category.label
}

function formatMenuPrice(price: number): string {
  return `${Math.round(price)} ₽`
}

function menuSections(...sections: string[]): string {
  return sections.map((section) => section.trim()).filter(Boolean).join('\n\n')
}

function truncateMenuText(text: string): string {
  const runes = Array.from(text.trim())
  if (runes.length <= 4000) return text.trim()
  return `${runes.slice(0, 4000).join('')}\n\n... ещё позиции`
}
