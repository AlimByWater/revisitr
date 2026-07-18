import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/bots/queries', () => ({
  useBotQuery: vi.fn(),
}))

vi.mock('@/features/lunch/queries', () => ({
  useLunchProgramQuery: vi.fn(),
}))

vi.mock('@/features/lunch/api', () => ({
  lunchApi: {
    updateProgram: vi.fn(),
    setAvailability: vi.fn(),
    createCourse: vi.fn(),
    updateCourse: vi.fn(),
    deleteCourse: vi.fn(),
    createFormat: vi.fn(),
    updateFormat: vi.fn(),
    deleteFormat: vi.fn(),
  },
}))

vi.mock('@/features/orders/queries', () => ({
  useOrdersQuery: vi.fn(),
}))

vi.mock('@/features/orders/api', () => ({
  ordersApi: {
    listOrders: vi.fn(),
    updateOrderStatus: vi.fn(),
  },
}))

vi.mock('@/features/menus/api', () => ({
  menusApi: {
    getBotPOSLocations: vi.fn(),
  },
}))

vi.mock('@/features/menus/queries', () => ({
  useMenusQuery: vi.fn(),
  useMenuQuery: vi.fn(),
}))

import { useBotQuery } from '@/features/bots/queries'
import { useLunchProgramQuery } from '@/features/lunch/queries'
import { lunchApi } from '@/features/lunch/api'
import { ordersApi } from '@/features/orders/api'
import { useOrdersQuery } from '@/features/orders/queries'
import { menusApi } from '@/features/menus/api'
import { useMenuQuery, useMenusQuery } from '@/features/menus/queries'
import BotLunchSettingsPage from './lunch'

const mockUseBotQuery = vi.mocked(useBotQuery)
const mockUseLunchProgramQuery = vi.mocked(useLunchProgramQuery)
const mockUseOrdersQuery = vi.mocked(useOrdersQuery)
const mockUseMenusQuery = vi.mocked(useMenusQuery)
const mockUseMenuQuery = vi.mocked(useMenuQuery)
const mockGetBotPOSLocations = vi.mocked(menusApi.getBotPOSLocations)

const botFixture = {
  id: 1,
  org_id: 1,
  name: 'Baratie',
  username: 'baratie_bot',
  status: 'active' as const,
  created_at: '2026-04-18T00:00:00Z',
  updated_at: '2026-04-18T00:00:00Z',
  settings: { modules: ['lunch'], buttons: [], registration_form: [], module_configs: {} },
}

const menuFixture = {
  id: 3,
  org_id: 1,
  name: 'Основное меню',
  source: 'manual' as const,
  created_at: '2026-04-18T00:00:00Z',
  updated_at: '2026-04-18T00:00:00Z',
  categories: [
    {
      id: 30,
      menu_id: 3,
      name: 'Супы',
      sort_order: 0,
      created_at: '2026-04-18T00:00:00Z',
      items: [
        {
          id: 100,
          category_id: 30,
          name: 'Борщ',
          price: 180,
          tags: [],
          is_available: true,
          sort_order: 0,
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
        },
      ],
    },
  ],
}

const programFixture = {
  id: 5,
  bot_id: 1,
  name: 'Бизнес-ланч',
  description: '',
  is_active: true,
  created_at: '2026-04-18T00:00:00Z',
  updated_at: '2026-04-18T00:00:00Z',
  courses: [
    {
      id: 10,
      program_id: 5,
      code: '1',
      title: 'Первое',
      menu_category_id: 30,
      sort_order: 0,
      items: [{ course_id: 10, menu_item_id: 100, surcharge: 0 }],
    },
  ],
  formats: [
    {
      id: 7,
      program_id: 5,
      name: 'Только первое',
      price_mode: 'fixed' as const,
      base_price: 350,
      is_active: true,
      sort_order: 0,
      course_ids: [10],
    },
  ],
  availability: [{ weekday: 1, time_from: '12:00', time_to: '16:00' }],
}

const orderFixture = {
  id: 3,
  bot_id: 1,
  bot_client_id: 42,
  source: 'lunch' as const,
  format_id: 7,
  format_name: 'Только первое',
  table_num: '7',
  total_price: 350,
  status: 'new' as const,
  created_at: '2026-07-17T12:30:00Z',
  items: [
    {
      id: 1,
      order_id: 3,
      course_id: 10,
      course_title: 'Первое',
      menu_item_id: 100,
      item_name: 'Борщ',
      price: 180,
      surcharge: 0,
    },
  ],
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/dashboard/bots/1/lunch']}>
      <Routes>
        <Route path="/dashboard/bots/:botId/lunch" element={<BotLunchSettingsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('BotLunchSettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseBotQuery.mockReturnValue({
      data: botFixture,
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useBotQuery>)
    mockUseLunchProgramQuery.mockReturnValue({
      data: programFixture,
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useLunchProgramQuery>)
    mockUseMenusQuery.mockReturnValue({
      data: [menuFixture],
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useMenusQuery>)
    mockUseMenuQuery.mockReturnValue({
      data: menuFixture,
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof useMenuQuery>)
    mockUseOrdersQuery.mockReturnValue({
      data: [orderFixture],
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useOrdersQuery>)
    mockGetBotPOSLocations.mockResolvedValue({ pos_ids: [] })
    vi.mocked(lunchApi.updateProgram).mockResolvedValue(programFixture)
    vi.mocked(lunchApi.setAvailability).mockResolvedValue()
    vi.mocked(lunchApi.updateFormat).mockResolvedValue()
    vi.mocked(ordersApi.updateOrderStatus).mockResolvedValue()
  })

  it('renders program, courses and formats', async () => {
    renderPage()

    expect(await screen.findByDisplayValue('Бизнес-ланч')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Первое')).toBeInTheDocument()
    expect(screen.getByDisplayValue('Только первое')).toBeInTheDocument()
  })

  it('saves program with availability slots', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Сохранить программу' }))

    await waitFor(() => {
      expect(lunchApi.updateProgram).toHaveBeenCalledWith(
        1,
        expect.objectContaining({ name: 'Бизнес-ланч', is_active: true }),
      )
      expect(lunchApi.setAvailability).toHaveBeenCalledWith(1, [
        expect.objectContaining({ weekday: 1, time_from: '12:00', time_to: '16:00' }),
      ])
    })
  })

  it('blocks saving a fixed-price format with zero price (FR-A6)', async () => {
    const user = userEvent.setup()
    renderPage()

    const priceInput = await screen.findByDisplayValue('350')
    await user.clear(priceInput)

    expect(screen.getByText('Для фиксированной цены укажите сумму больше 0')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Сохранить формат' })).toBeDisabled()
  })

  it('shows new orders and marks one as processed', async () => {
    const user = userEvent.setup()
    renderPage()

    expect(await screen.findByText(/№3 · Стол 7/)).toBeInTheDocument()
    expect(screen.getByText(/Первое: Борщ/)).toBeInTheDocument()

    await user.click(screen.getByRole('button', { name: 'Отработан' }))

    await waitFor(() => {
      expect(ordersApi.updateOrderStatus).toHaveBeenCalledWith(3, 'sent')
    })
  })

  it('saves an edited format', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Сохранить формат' }))

    await waitFor(() => {
      expect(lunchApi.updateFormat).toHaveBeenCalledWith(
        7,
        expect.objectContaining({ price_mode: 'fixed', base_price: 350, course_ids: [10] }),
      )
    })
  })
})
