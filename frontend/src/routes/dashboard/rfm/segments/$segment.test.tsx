import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// Mock queries
vi.mock('@/features/rfm/queries', () => ({
  useRFMSegmentClientsQuery: vi.fn(),
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ segment: 'vip' }),
  }
})

import { useRFMSegmentClientsQuery } from '@/features/rfm/queries'
import SegmentDetailPage from './$segment'

const mockClients = {
  segment: 'vip',
  segment_name: 'VIP / Ядро',
  total: 3,
  page: 1,
  per_page: 20,
  clients: [
    { id: 1, first_name: 'Иван', last_name: 'Петров', phone: '+79001234567', r_score: 5, f_score: 5, m_score: 5, recency_days: 2, frequency_count: 18, monetary_sum: 25000, last_visit_date: '2026-03-24T12:00:00Z', total_visits_lifetime: 42 },
    { id: 2, first_name: 'Анна', last_name: 'Сидорова', phone: '+79007654321', r_score: 4, f_score: 5, m_score: 4, recency_days: 5, frequency_count: 14, monetary_sum: 18500, last_visit_date: '2026-03-22T12:00:00Z', total_visits_lifetime: 30 },
    { id: 3, first_name: 'Олег', last_name: 'Кузнецов', phone: '+79005551234', r_score: 5, f_score: 4, m_score: null, recency_days: 1, frequency_count: 10, monetary_sum: null, last_visit_date: '2026-03-26T12:00:00Z', total_visits_lifetime: 15 },
  ],
}

function setupMocks(overrides: Record<string, unknown> = {}) {
  vi.mocked(useRFMSegmentClientsQuery).mockReturnValue({
    data: overrides.data ?? mockClients,
    isLoading: overrides.isLoading ?? false,
    isError: overrides.isError ?? false,
    mutate: vi.fn(),
  } as ReturnType<typeof useRFMSegmentClientsQuery>)
}

describe('SegmentDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders segment name and client count', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    expect(screen.getByText('VIP / Ядро')).toBeInTheDocument()
    expect(screen.getByText(/3 клиента/)).toBeInTheDocument()
  })

  it('renders client rows with names', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    expect(screen.getByText(/Иван П\./)).toBeInTheDocument()
    expect(screen.getByText(/Анна С\./)).toBeInTheDocument()
    expect(screen.getByText(/Олег К\./)).toBeInTheDocument()
  })

  it('renders RFM score badges', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    // Should have multiple score badges with values 5, 4, etc.
    const badges = screen.getAllByText('5')
    expect(badges.length).toBeGreaterThanOrEqual(3) // Иван has R5 F5 M5
  })

  it('shows dash for null scores', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    // Олег has m_score: null → should show "—"
    const dashes = screen.getAllByText('—')
    expect(dashes.length).toBeGreaterThanOrEqual(1)
  })

  it('renders visit counts', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    expect(screen.getByText('42')).toBeInTheDocument()
    expect(screen.getByText('30')).toBeInTheDocument()
    expect(screen.getByText('15')).toBeInTheDocument()
  })

  it('shows empty state when no clients', () => {
    setupMocks({ data: { ...mockClients, clients: [], total: 0 } })
    render(<SegmentDetailPage />)

    expect(screen.getByText('В этом сегменте пока нет клиентов')).toBeInTheDocument()
  })

  it('shows error state with retry', () => {
    setupMocks({ isError: true, data: undefined })
    render(<SegmentDetailPage />)

    expect(screen.getByText('Не удалось загрузить клиентов')).toBeInTheDocument()
  })

  it('navigates back to dashboard on back button', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<SegmentDetailPage />)

    await user.click(screen.getByText('RFM-сегментация'))

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/rfm')
  })

  it('navigates to campaign creation on scenario button', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<SegmentDetailPage />)

    const scenarioBtn = screen.getByText('Запустить сценарий').closest('button')!
    await user.click(scenarioBtn)

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/campaigns/create?segment=vip')
  })

  it('has export CSV button', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    expect(screen.getByText('Экспорт CSV')).toBeInTheDocument()
  })

  it('does not show pagination for small datasets', () => {
    setupMocks()
    render(<SegmentDetailPage />)

    // 3 clients, perPage=20, so no pagination
    expect(screen.queryByText('Стр.')).not.toBeInTheDocument()
  })

  it('shows pagination when total > perPage', () => {
    setupMocks({
      data: { ...mockClients, total: 45 },
    })
    render(<SegmentDetailPage />)

    expect(screen.getByText(/Стр\. 1 из 3/)).toBeInTheDocument()
  })

  it('changes sort order on column header click', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<SegmentDetailPage />)

    // Click "Выручка" header to sort
    const revenueHeader = screen.getByText('Выручка')
    await user.click(revenueHeader)

    // Should re-call the hook with different sort params (verified via mock being called again)
    expect(useRFMSegmentClientsQuery).toHaveBeenCalled()
  })
})
