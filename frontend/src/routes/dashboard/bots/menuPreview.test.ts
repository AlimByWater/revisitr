import { describe, expect, it } from 'vitest'
import type { MenuPresetCustomizations } from '@/features/bots/types'
import type { Menu } from '@/features/menus/types'
import {
  buildCarouselContent,
  buildCategoryContent,
  buildItemCardContent,
  buildItemSimpleContent,
  buildMenuInitialContent,
  buildMenuListContent,
  buildMenuPreviewState,
  buildMenuTabsContent,
  flattenCarouselEntries,
} from './menuPreview'

const menu: Menu = {
  id: 1,
  org_id: 10,
  name: 'Baratie Signature Menu',
  source: 'manual',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  categories: [
    {
      id: 20,
      menu_id: 1,
      name: 'Основные блюда',
      sort_order: 2,
      created_at: '2026-01-01T00:00:00Z',
      items: [
        {
          id: 201,
          category_id: 20,
          name: 'Sea King Steak',
          price: 222,
          is_available: true,
          sort_order: 2,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          tags: [],
        },
        {
          id: 202,
          category_id: 20,
          name: 'All Blue Sashimi',
          price: 199,
          is_available: true,
          sort_order: 1,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          tags: [],
        },
      ],
    },
    {
      id: 10,
      menu_id: 1,
      name: 'Закуски',
      sort_order: 1,
      created_at: '2026-01-01T00:00:00Z',
      items: [
        {
          id: 101,
          category_id: 10,
          name: 'Grand Line Bruschetta',
          price: 149,
          is_available: true,
          sort_order: 1,
          created_at: '2026-01-01T00:00:00Z',
          updated_at: '2026-01-01T00:00:00Z',
          tags: [],
          weight: '120 г',
        },
      ],
    },
  ],
}

const draft = {} satisfies MenuPresetCustomizations

describe('menu preview builders', () => {
  it('builds tabs preview with active category items inline', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'tabs',
    })

    const content = buildMenuTabsContent(state, 20)
    expect(content.parts[0]?.text).toContain('Baratie Signature Menu')
    expect(content.parts[0]?.text).toContain('Sea King Steak')
    expect(content.parts[0]?.text).toContain('All Blue Sashimi')
    expect(content.buttons?.[0]?.[0]?.data).toBe('menu:tab:10')
    expect(content.buttons?.[0]?.[1]?.style).toBe('success')
    expect(content.buttons?.[1]?.[0]?.data).toBe('menu:card:202:20')
  })

  it('builds tabs default screen from first category order', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'tabs',
    })

    const content = buildMenuInitialContent(state)
    expect(content.parts[0]?.text).toContain('Grand Line Bruschetta')
    expect(content.parts[0]?.text).not.toContain('Sea King Steak')
  })

  it('builds list preview with category sections and category buttons', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'list',
    })

    const content = buildMenuListContent(state)
    expect(content.parts[0]?.text).toContain('Baratie Signature Menu')
    expect(content.parts[0]?.text).toContain('Выберите категорию')
    expect(content.parts[0]?.text).not.toContain('Закуски')
    expect(content.buttons).toHaveLength(2)
    expect(content.buttons?.[0]?.[0]?.data).toBe('menu:cat:10:0')
  })

  it('builds expanded list preview with compact item details by default', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft: {
        list_layout: 'expanded',
        list_density: 'compact',
      },
      selectedPresetKey: 'list',
    })

    const content = buildMenuListContent(state)
    expect(content.parts[0]?.text).toContain('Grand Line Bruschetta — 149 ₽')
    expect(content.parts[0]?.text).not.toContain('120 г')
    expect(content.parts[0]?.text).not.toContain('Тигровые')
  })

  it('builds category screen with back button', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'list',
    })

    const category = state.categories[0]
    const content = buildCategoryContent(category, 'detailed')
    expect(content.parts[0]?.text).toContain('Grand Line Bruschetta — 149 ₽ • 120 г')
    expect(content.buttons?.[content.buttons.length - 1]?.[0]?.data).toBe('menu:root')
  })

  it('switches tabs content to a different category without outgoing bubble', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'tabs',
    })

    const content = buildMenuTabsContent(state, 20)
    expect(content.parts[0]?.text).toContain('All Blue Sashimi')
    expect(content.parts[0]?.text).not.toContain('Grand Line Bruschetta')
    expect(content.buttons?.[0]?.[1]?.style).toBe('success')
  })

  it('builds tabs item card content with arrows, back and close buttons', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'tabs',
    })

    const category = state.categories.find((item) => item.id === 20)
    expect(category).toBeTruthy()
    const content = buildItemCardContent(category!, 201)
    expect(content?.parts[0]?.text).toContain('Sea King Steak')
    expect(content?.buttons?.[0]?.[0]?.data).toBe('menu:cardnav:202:20')
    expect(content?.buttons?.[1]?.[0]?.data).toBe('menu:tab:20')
    expect(content?.buttons?.[2]?.[0]?.data).toBe('menu:cardclose')
  })

  it('builds list item card with only back-to-category button', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'list',
    })

    const category = state.categories.find((item) => item.id === 10)
    expect(category).toBeTruthy()
    const content = buildItemSimpleContent(category!, 101)
    expect(content?.parts[0]?.text).toContain('Grand Line Bruschetta')
    expect(content?.buttons).toHaveLength(1)
    expect(content?.buttons?.[0]?.[0]?.data).toBe('menu:cat:10:0')
    expect(content?.buttons?.[0]?.[0]?.text).toBe('Назад к категории')
  })

  it('builds carousel preset content with flat navigation across all items', () => {
    const state = buildMenuPreviewState({
      introContent: { parts: [{ type: 'text', text: 'Baratie Signature Menu' }] },
      menu,
      draft,
      selectedPresetKey: 'carousel',
    })

    const entries = flattenCarouselEntries(state)
    expect(entries).toHaveLength(3)

    const first = buildCarouselContent(entries, 0)
    expect(first?.buttons?.[0]?.[0]?.text).toBe('1/3')
    expect(first?.buttons?.[0]?.[1]?.data).toBe('menu:car:1')
    expect(first?.buttons?.[1]?.[0]?.data).toBe('menu:root')

    const middle = buildCarouselContent(entries, 1)
    expect(middle?.buttons?.[0]?.[0]?.data).toBe('menu:car:0')
    expect(middle?.buttons?.[0]?.[2]?.data).toBe('menu:car:2')
  })
})
