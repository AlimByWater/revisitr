import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/orders/queries', () => ({
  useOrgOrdersQuery: vi.fn(),
}))

vi.mock('@/features/orders/api', () => ({
  ordersApi: {
    listOrgOrders: vi.fn(),
    updateOrderStatus: vi.fn(),
  },
}))

import { ordersApi } from '@/features/orders/api'
import { useOrgOrdersQuery } from '@/features/orders/queries'
import OrdersPage from './orders'

const mockUseOrgOrdersQuery = vi.mocked(useOrgOrdersQuery)

const orderFixture = {
  id: 3,
  bot_id: 2,
  bot_client_id: 42,
  source: 'lunch' as const,
  format_id: 7,
  format_name: 'Только первое',
  table_num: '7',
  total_price: 350,
  status: 'new' as const,
  created_at: '2026-07-18T12:30:00Z',
  bot_name: 'Baratie',
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
    <MemoryRouter initialEntries={['/dashboard/orders']}>
      <OrdersPage />
    </MemoryRouter>,
  )
}

describe('OrdersPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseOrgOrdersQuery.mockReturnValue({
      data: [orderFixture],
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useOrgOrdersQuery>)
    vi.mocked(ordersApi.updateOrderStatus).mockResolvedValue()
  })

  it('renders orders with source and bot name', async () => {
    renderPage()

    expect(await screen.findByText(/№3 · Стол 7/)).toBeInTheDocument()
    // «Ланч» встречается дважды: кнопка фильтра и бейдж источника на карточке
    expect(screen.getAllByText('Ланч')).toHaveLength(2)
    expect(screen.getByText(/Baratie · Только первое/)).toBeInTheDocument()
    expect(screen.getByText(/Первое: Борщ/)).toBeInTheDocument()
  })

  it('marks an order as processed', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(await screen.findByRole('button', { name: 'Отработан' }))

    await waitFor(() => {
      expect(ordersApi.updateOrderStatus).toHaveBeenCalledWith(3, 'sent')
    })
  })

  it('filters by source', async () => {
    const user = userEvent.setup()
    renderPage()

    await user.click(screen.getByRole('button', { name: 'Ланч' }))

    expect(mockUseOrgOrdersQuery).toHaveBeenLastCalledWith('lunch', 'new')
  })
})
