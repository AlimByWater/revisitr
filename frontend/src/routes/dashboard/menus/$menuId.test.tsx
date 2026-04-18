import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/menus/queries', () => ({
  useMenuQuery: vi.fn(),
  useAddCategoryMutation: vi.fn(),
  useAddItemMutation: vi.fn(),
  useUpdateCategoryMutation: vi.fn(),
  useUpdateItemMutation: vi.fn(),
  useUpdateMenuMutation: vi.fn(),
}))

vi.mock('@/features/pos/queries', () => ({
  usePOSQuery: vi.fn(),
}))

vi.mock('@/features/telegram-preview', () => ({
  MessageContentEditor: () => <div>mock-message-editor</div>,
}))

vi.mock('@/features/campaigns/api', () => ({
  campaignsApi: {
    uploadFile: vi.fn(),
  },
}))

import {
  useAddCategoryMutation,
  useAddItemMutation,
  useMenuQuery,
  useUpdateCategoryMutation,
  useUpdateItemMutation,
  useUpdateMenuMutation,
} from '@/features/menus/queries'
import { usePOSQuery } from '@/features/pos/queries'
import MenuDetailPage from './$menuId'

const mockUseMenuQuery = vi.mocked(useMenuQuery)
const mockUseAddCategoryMutation = vi.mocked(useAddCategoryMutation)
const mockUseAddItemMutation = vi.mocked(useAddItemMutation)
const mockUseUpdateCategoryMutation = vi.mocked(useUpdateCategoryMutation)
const mockUseUpdateItemMutation = vi.mocked(useUpdateItemMutation)
const mockUseUpdateMenuMutation = vi.mocked(useUpdateMenuMutation)
const mockUsePOSQuery = vi.mocked(usePOSQuery)

function renderPage(initialEntry = '/dashboard/menus/5?botId=12') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/dashboard/menus/:menuId" element={<MenuDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

function mutationStub() {
  return {
    mutate: vi.fn(),
    mutateAsync: vi.fn(),
    isPending: false,
    isError: false,
    isSuccess: false,
    isMutating: false,
    reset: vi.fn(),
    trigger: vi.fn(),
    data: undefined,
    error: undefined,
  }
}

describe('MenuDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    mockUseAddCategoryMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useAddCategoryMutation>)

    mockUseAddItemMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useAddItemMutation>)

    mockUseUpdateCategoryMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useUpdateCategoryMutation>)

    mockUseUpdateItemMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useUpdateItemMutation>)

    mockUseUpdateMenuMutation.mockReturnValue({
      ...mutationStub(),
    } as unknown as ReturnType<typeof useUpdateMenuMutation>)

    mockUsePOSQuery.mockReturnValue({
      data: [
        {
          id: 10,
          org_id: 1,
          name: 'Маросейка',
          address: 'Маросейка, 12',
          phone: '+7 000 000-00-00',
          schedule: {},
          is_active: true,
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
        },
      ],
      isLoading: false,
      isError: false,
      error: undefined,
      mutate: vi.fn(),
      isValidating: false,
    } as unknown as ReturnType<typeof usePOSQuery>)
  })

  it('renders intro editor, bindings, category icon fields, and full item data', () => {
    mockUseMenuQuery.mockReturnValue({
      data: {
        id: 5,
        org_id: 1,
        name: 'Основное меню',
        source: 'manual',
        intro_content: {
          parts: [{ type: 'text', text: 'Добро пожаловать', parse_mode: 'Markdown' }],
        },
        created_at: '2026-04-18T00:00:00Z',
        updated_at: '2026-04-18T00:00:00Z',
        bindings: [{ menu_id: 5, pos_id: 10, pos_name: 'Маросейка', is_active: true }],
        categories: [
          {
            id: 101,
            menu_id: 5,
            name: 'Кофе',
            icon_emoji: '☕',
            icon_image_url: 'https://cdn.test/icon.png',
            sort_order: 1,
            created_at: '2026-04-18T00:00:00Z',
            items: [
              {
                id: 201,
                category_id: 101,
                name: 'Капучино',
                description: 'Классический капучино',
                price: 350,
                weight: '250 мл',
                image_url: 'https://cdn.test/cappuccino.png',
                tags: [],
                is_available: true,
                sort_order: 1,
                created_at: '2026-04-18T00:00:00Z',
                updated_at: '2026-04-18T00:00:00Z',
              },
            ],
          },
        ],
      },
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
      error: undefined,
      isValidating: false,
    } as unknown as ReturnType<typeof useMenuQuery>)

    renderPage()

    expect(screen.getByRole('link', { name: /Назад к модулям/i })).toHaveAttribute(
      'href',
      '/dashboard/bots/12?tab=modules',
    )
    expect(screen.getByText('Первое сообщение')).toBeInTheDocument()
    expect(screen.getByText('Привязка к точкам продаж')).toBeInTheDocument()
    expect(screen.getByDisplayValue('☕')).toBeInTheDocument()
    expect(screen.getByDisplayValue('https://cdn.test/icon.png')).toBeInTheDocument()
    expect(screen.getByText('250 мл')).toBeInTheDocument()
    expect(screen.getByText('Есть фото')).toBeInTheDocument()
  })
})
