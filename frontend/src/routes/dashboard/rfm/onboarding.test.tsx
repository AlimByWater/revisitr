import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// Mock queries
vi.mock('@/features/rfm/queries', () => ({
  useRFMOnboardingQuestionsQuery: vi.fn(),
  useRFMRecommendMutation: vi.fn(),
  useRFMSetTemplateMutation: vi.fn(),
  useRFMActiveTemplateQuery: vi.fn(),
  useRFMTemplatesQuery: vi.fn(),
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return { ...actual, useNavigate: () => mockNavigate }
})

import {
  useRFMOnboardingQuestionsQuery,
  useRFMRecommendMutation,
  useRFMSetTemplateMutation,
  useRFMActiveTemplateQuery,
  useRFMTemplatesQuery,
} from '@/features/rfm/queries'
import RFMOnboardingPage from './onboarding'

const mockQuestions = [
  { id: 1, text: 'Тип заведения?', answers: [{ id: 1, text: 'Кофейня' }, { id: 2, text: 'Ресторан' }] },
  { id: 2, text: 'Средний чек?', answers: [{ id: 3, text: 'До 500₽' }, { id: 4, text: '500–1500₽' }] },
  { id: 3, text: 'Частота?', answers: [{ id: 5, text: 'Каждый день' }, { id: 6, text: 'Раз в неделю' }] },
]

const mockRecommendation = {
  recommended: { key: 'coffeegng', name: 'Кофейня / Go-to', description: 'Для кофеен', r_thresholds: [7, 14, 30, 60] as [number, number, number, number], f_thresholds: [8, 5, 3, 2] as [number, number, number, number] },
  all_scores: { coffeegng: 10, qsr: 5 },
}

function setupMocks(overrides: Record<string, unknown> = {}) {
  vi.mocked(useRFMActiveTemplateQuery).mockReturnValue({
    data: overrides.activeTemplate ?? null,
    isLoading: false,
    isError: false,
    mutate: vi.fn(),
  } as ReturnType<typeof useRFMActiveTemplateQuery>)

  vi.mocked(useRFMOnboardingQuestionsQuery).mockReturnValue({
    data: overrides.questions ?? mockQuestions,
    isLoading: false,
    isError: overrides.isError ?? false,
    mutate: vi.fn(),
  } as ReturnType<typeof useRFMOnboardingQuestionsQuery>)

  vi.mocked(useRFMRecommendMutation).mockReturnValue({
    data: overrides.recommendation ?? undefined,
    mutate: overrides.recommendMutate ?? vi.fn(),
    mutateAsync: overrides.recommendMutateAsync ?? vi.fn(),
    isPending: false,
    isError: false,
    isSuccess: false,
  } as ReturnType<typeof useRFMRecommendMutation>)

  vi.mocked(useRFMSetTemplateMutation).mockReturnValue({
    mutate: vi.fn(),
    mutateAsync: overrides.setTemplateMutateAsync ?? vi.fn().mockResolvedValue({}),
    isPending: false,
    isError: false,
    isSuccess: false,
  } as ReturnType<typeof useRFMSetTemplateMutation>)

  vi.mocked(useRFMTemplatesQuery).mockReturnValue({
    data: overrides.allTemplates ?? [mockRecommendation.recommended],
    isLoading: false,
    isError: false,
    mutate: vi.fn(),
  } as ReturnType<typeof useRFMTemplatesQuery>)
}

describe('RFMOnboardingPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers({ shouldAdvanceTime: true })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders first question with answers', () => {
    setupMocks()
    render(<RFMOnboardingPage />)

    expect(screen.getByText('Тип заведения?')).toBeInTheDocument()
    expect(screen.getByText('Кофейня')).toBeInTheDocument()
    expect(screen.getByText('Ресторан')).toBeInTheDocument()
  })

  it('shows progress text with step 1 of 3', () => {
    setupMocks()
    render(<RFMOnboardingPage />)

    expect(screen.getByText(/1 из 3/)).toBeInTheDocument()
  })

  it('advances to next question after answer + delay', async () => {
    setupMocks()
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    render(<RFMOnboardingPage />)

    await user.click(screen.getByText('Кофейня'))
    vi.advanceTimersByTime(350)

    await waitFor(() => {
      expect(screen.getByText('Средний чек?')).toBeInTheDocument()
    })
  })

  it('shows error state when questions fail to load', () => {
    setupMocks({ isError: true, questions: undefined })
    render(<RFMOnboardingPage />)

    expect(screen.getByText('Не удалось загрузить вопросы')).toBeInTheDocument()
  })

  it('redirects to /dashboard/rfm if template already set', () => {
    setupMocks({
      activeTemplate: {
        active_template_type: 'standard',
        active_template_key: 'coffeegng',
        template: mockRecommendation.recommended,
      },
    })
    render(<RFMOnboardingPage />)

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/rfm', { replace: true })
  })

  it('shows recommendation after all 3 answers', async () => {
    const recommendMutate = vi.fn()
    setupMocks({ recommendation: mockRecommendation, recommendMutate })
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime })
    render(<RFMOnboardingPage />)

    // Answer Q1
    await user.click(screen.getByText('Кофейня'))
    vi.advanceTimersByTime(350)

    // Answer Q2
    await waitFor(() => expect(screen.getByText('Средний чек?')).toBeInTheDocument())
    await user.click(screen.getByText('До 500₽'))
    vi.advanceTimersByTime(350)

    // Answer Q3
    await waitFor(() => expect(screen.getByText('Частота?')).toBeInTheDocument())
    await user.click(screen.getByText('Каждый день'))
    vi.advanceTimersByTime(350)

    // Recommendation should have been called
    expect(recommendMutate).toHaveBeenCalled()

    // Result screen should show
    await waitFor(() => {
      expect(screen.getByText(/Кофейня \/ Go-to/)).toBeInTheDocument()
      expect(screen.getByText('Использовать')).toBeInTheDocument()
    })
  })

  it('shows loading spinner when questions not yet loaded', () => {
    setupMocks({ questions: undefined })
    // Override to return null data (loading)
    vi.mocked(useRFMOnboardingQuestionsQuery).mockReturnValue({
      data: undefined,
      isLoading: true,
      isError: false,
      mutate: vi.fn(),
    } as ReturnType<typeof useRFMOnboardingQuestionsQuery>)

    const { container } = render(<RFMOnboardingPage />)
    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
  })
})
