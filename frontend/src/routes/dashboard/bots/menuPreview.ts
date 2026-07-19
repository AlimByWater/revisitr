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
  style?: Exclude<InlineButton['style'], ''>
  items: MenuPreviewItem[]
}

export interface CarouselEntry {
  item: MenuPreviewItem
  categoryHeading: string
}

export type MenuPresetKind = 'tabs' | 'list' | 'carousel'

export interface MenuPreviewState {
  title: string
  subtitle: string
  intro: string
  categories: MenuPreviewCategory[]
  presetKey: MenuPresetKind
  listLayout: 'summary' | 'expanded'
  listDensity: 'compact' | 'detailed'
  navButtonStyle: Exclude<InlineButton['style'], ''> | ''
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

  const tabFallbackStyle = (args.draft.tab_button_style as Exclude<InlineButton['style'], ''> | undefined) || undefined

  const categories = categoryOrder
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
        style: (categoryDraft?.style as Exclude<InlineButton['style'], ''>) || tabFallbackStyle,
        items: orderedItems,
      } satisfies MenuPreviewCategory
    })
    .filter(Boolean) as MenuPreviewCategory[]

  return {
    title,
    subtitle,
    intro,
    categories,
    presetKey: normalizePresetKey(args.selectedPresetKey),
    listLayout: args.draft.list_layout === 'expanded' ? 'expanded' : 'summary',
    listDensity: args.draft.list_density === 'detailed' ? 'detailed' : 'compact',
    navButtonStyle: (args.draft.nav_button_style as Exclude<InlineButton['style'], ''>) || '',
  }
}

function normalizePresetKey(value: string): MenuPresetKind {
  if (value === 'list' || value === 'carousel') return value
  return 'tabs'
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
      ...buildTabItemRows(selected.id, selected.items),
    ],
  }
}

export function buildMenuListContent(state: MenuPreviewState): MessageContent {
  const sections = [buildListHeaderText(state)]
  if (state.listLayout === 'expanded') {
    for (const category of state.categories) {
      sections.push(menuSections(categoryHeading(category), menuCategoryItemsText(category.items, state.listDensity)))
    }
  } else {
    sections.push('Выберите категорию')
  }

  return {
    parts: [{ type: 'text' as const, text: truncateMenuText(menuSections(...sections)) }],
    buttons: buildMenuListCategoryRows(state.categories),
  }
}

export function buildCategoryContent(category: MenuPreviewCategory, density: 'compact' | 'detailed'): MessageContent {
  const buttons: InlineButton[][] = category.items.map((item) => [
    { text: item.label, data: `menu:item:${item.id}:${category.id}:0` },
  ])
  buttons.push([{ text: 'Назад к меню', data: 'menu:root' }])

  const text = `${categoryHeading(category)}\n\n${menuCategoryItemsText(category.items, density)}`

  if (isMediaUrl(category.icon)) {
    return {
      parts: [{ type: 'photo', media_url: category.icon, text }],
      buttons,
    }
  }

  return {
    parts: [{ type: 'text' as const, text }],
    buttons,
  }
}

// List preset card: only "Назад к категории" button — matches backend handleMenuItemCallback.
export function buildItemSimpleContent(category: MenuPreviewCategory, itemID: number, page = 0): MessageContent | null {
  const item = category.items.find((entry) => entry.id === itemID)
  if (!item) return null

  const lines: string[] = [`<b>${item.label}</b>`]
  if (item.description?.trim()) lines.push(item.description.trim())
  lines.push(`Цена: ${formatMenuPrice(item.price)}`)
  if (item.weight?.trim()) lines.push(`Граммаж: <i>${item.weight.trim()}</i>`)

  return {
    parts: [{ type: 'text' as const, text: lines.join('\n'), parse_mode: 'HTML' }],
    buttons: [[{ text: 'Назад к категории', data: `menu:cat:${category.id}:${page}` }]],
  }
}

// Tabs preset card: arrows + "Назад к категории" → menu:tab + "✕ Закрыть" — matches backend menuItemCardContent.
export function buildItemCardContent(category: MenuPreviewCategory, itemID: number, navStyle: InlineButton['style'] = ''): MessageContent | null {
  const index = category.items.findIndex((item) => item.id === itemID)
  if (index === -1) return null

  const item = category.items[index]
  const sections = [
    categoryHeading(category),
    `<b>${item.label} — ${formatMenuPrice(item.price)}</b>`,
    item.weight?.trim() ? `<i>${item.weight.trim()}</i>` : '',
    item.description?.trim() || '',
  ]

  const buttons: InlineButton[][] = []
  const nav: InlineButton[] = []

  if (index > 0) {
    nav.push({ text: '←', data: `menu:cardnav:${category.items[index - 1].id}:${category.id}`, style: navStyle })
  }
  nav.push({ text: `${index + 1}/${category.items.length}`, data: 'menu:noop', style: navStyle })
  if (index < category.items.length - 1) {
    nav.push({ text: '→', data: `menu:cardnav:${category.items[index + 1].id}:${category.id}`, style: navStyle })
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

export function flattenCarouselEntries(state: MenuPreviewState): CarouselEntry[] {
  const entries: CarouselEntry[] = []
  for (const category of state.categories) {
    const heading = categoryHeading(category)
    for (const item of category.items) {
      entries.push({ item, categoryHeading: heading })
    }
  }
  return entries
}

// Carousel preset: flat ←/→ navigation across all items + "Назад к меню" — matches backend carouselContent.
export function buildCarouselContent(entries: CarouselEntry[], index: number, navStyle: InlineButton['style'] = ''): MessageContent | null {
  if (entries.length === 0 || index < 0 || index >= entries.length) return null
  const { item, categoryHeading: heading } = entries[index]
  const text = truncateMenuText(menuSections(
    heading,
    `<b>${item.label} — ${formatMenuPrice(item.price)}</b>`,
    item.weight?.trim() ? `<i>${item.weight.trim()}</i>` : '',
    item.description?.trim() || '',
  ))

  const nav: InlineButton[] = []
  if (index > 0) nav.push({ text: '←', data: `menu:car:${index - 1}`, style: navStyle })
  nav.push({ text: `${index + 1}/${entries.length}`, data: 'menu:car:noop', style: navStyle })
  if (index < entries.length - 1) nav.push({ text: '→', data: `menu:car:${index + 1}`, style: navStyle })

  const buttons: InlineButton[][] = [nav, [{ text: 'Назад к меню', data: 'menu:root' }]]

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
      ...buildTabItemRows(selected.id, selected.items),
    ],
  }
}

function buildTabsHeaderText(state: MenuPreviewState): string {
  return [state.title, state.subtitle].filter(Boolean).join('\n')
}

function buildListHeaderText(state: MenuPreviewState): string {
  return [state.title, state.subtitle].filter(Boolean).join('\n')
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

// Tabs preset opens carousel-style card → callback `menu:card:` (matches backend).
function buildTabItemRows(categoryID: number, items: MenuPreviewItem[]): InlineButton[][] {
  return items.map((item) => [{ text: item.label, data: `menu:card:${item.id}:${categoryID}` }])
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

function menuCategoryItemsText(items: MenuPreviewItem[], density: 'compact' | 'detailed'): string {
  if (items.length === 0) return 'Сейчас в этой категории нет доступных позиций.'

  const lines: string[] = []
  for (const item of items) {
    let line = `${item.label} — ${formatMenuPrice(item.price)}`
    if (density === 'detailed' && item.weight?.trim()) {
      line += ` • ${item.weight.trim()}`
    }
    lines.push(line)
    if (density === 'detailed' && item.description?.trim()) lines.push(item.description.trim())
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
  if (!category.icon || isMediaUrl(category.icon)) return category.label
  return `${category.icon} ${category.label}`.trim()
}

function isMediaUrl(value: string | undefined | null): value is string {
  if (!value) return false
  return value.startsWith('/') || value.startsWith('http://') || value.startsWith('https://')
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
