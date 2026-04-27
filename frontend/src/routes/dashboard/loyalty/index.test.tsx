import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

const mockNavigate = vi.fn()

vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

vi.mock('@/features/loyalty/queries', () => ({
  useProgramsQuery: vi.fn(),
  useUpdateProgramMutation: vi.fn(() => ({ mutate: vi.fn() })),
}))

vi.mock('@/components/loyalty/CreateProgramModal', () => ({
  CreateProgramModal: () => <div>create-program-modal</div>,
}))

import { useProgramsQuery, useUpdateProgramMutation } from '@/features/loyalty/queries'
import LoyaltyProgramsPage from './index'

const mockUseProgramsQuery = vi.mocked(useProgramsQuery)
const mockUseUpdateProgramMutation = vi.mocked(useUpdateProgramMutation)

function renderPage(initialEntry = '/dashboard/loyalty?botId=1') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/dashboard/loyalty" element={<LoyaltyProgramsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('LoyaltyProgramsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseProgramsQuery.mockReturnValue({
      data: [
        {
          id: 3,
          org_id: 1,
          name: 'Baratie Rewards',
          type: 'bonus',
          config: { welcome_bonus: 150, currency_name: 'дублонов' },
          levels: [],
          is_active: true,
          created_at: '2026-04-18T00:00:00Z',
          updated_at: '2026-04-18T00:00:00Z',
        },
      ],
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useProgramsQuery>)
    mockUseUpdateProgramMutation.mockReturnValue({ mutate: vi.fn() } as unknown as ReturnType<typeof useUpdateProgramMutation>)
  })

  it('keeps bot context when opening a loyalty program', async () => {
    renderPage('/dashboard/loyalty?botId=1')

    expect(screen.getByRole('link', { name: 'Назад к модулям бота' })).toHaveAttribute(
      'href',
      '/dashboard/bots/1?tab=modules',
    )

    // Program card is a <Link> — verify the href preserves botId context
    const programLink = screen.getByRole('link', { name: /Baratie Rewards/ })
    expect(programLink).toHaveAttribute('href', '/dashboard/loyalty/3?botId=1')
  })
})
