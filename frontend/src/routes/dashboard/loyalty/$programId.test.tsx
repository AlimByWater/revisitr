import { render, screen } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/loyalty/queries', () => ({
  useProgramQuery: vi.fn(),
  useUpdateProgramMutation: vi.fn(() => ({ mutateAsync: vi.fn() })),
  useUpdateLevelsMutation: vi.fn(() => ({ mutateAsync: vi.fn() })),
}))

vi.mock('@/features/loyalty/api', () => ({
  deleteLevel: vi.fn(),
}))

import {
  useProgramQuery,
  useUpdateLevelsMutation,
  useUpdateProgramMutation,
} from '@/features/loyalty/queries'
import ProgramDetailPage from './$programId'

const mockUseProgramQuery = vi.mocked(useProgramQuery)
const mockUseUpdateProgramMutation = vi.mocked(useUpdateProgramMutation)
const mockUseUpdateLevelsMutation = vi.mocked(useUpdateLevelsMutation)

function renderPage(initialEntry = '/dashboard/loyalty/3?botId=1') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/dashboard/loyalty/:programId" element={<ProgramDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('ProgramDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseProgramQuery.mockReturnValue({
      data: {
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
      isLoading: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useProgramQuery>)
    mockUseUpdateProgramMutation.mockReturnValue({ mutateAsync: vi.fn() } as unknown as ReturnType<typeof useUpdateProgramMutation>)
    mockUseUpdateLevelsMutation.mockReturnValue({ mutateAsync: vi.fn() } as unknown as ReturnType<typeof useUpdateLevelsMutation>)
  })

  it('returns to bot modules when opened from bot module settings', () => {
    renderPage('/dashboard/loyalty/3?botId=1')

    expect(screen.getByRole('link', { name: 'Назад к модулям бота' })).toHaveAttribute(
      'href',
      '/dashboard/bots/1?tab=modules',
    )
  })

  it('returns to loyalty programs without bot context', () => {
    renderPage('/dashboard/loyalty/3')

    expect(screen.getByRole('link', { name: 'Назад к программам' })).toHaveAttribute(
      'href',
      '/dashboard/loyalty',
    )
  })
})
