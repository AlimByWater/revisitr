import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// Mock queries
vi.mock('@/features/rfm/queries', () => ({
  useRFMTemplatesQuery: vi.fn(),
  useRFMActiveTemplateQuery: vi.fn(),
  useRFMSetTemplateMutation: vi.fn(),
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useSearchParams: () => [new URLSearchParams(), vi.fn()],
  }
})

import {
  useRFMTemplatesQuery,
  useRFMActiveTemplateQuery,
  useRFMSetTemplateMutation,
} from '@/features/rfm/queries'
import RFMTemplatePage from './template'

const mockTemplates = [
  { key: 'coffeegng', name: 'Кофейня / Go-to', description: 'Для кофеен', r_thresholds: [7, 14, 30, 60] as [number, number, number, number], f_thresholds: [8, 5, 3, 2] as [number, number, number, number] },
  { key: 'qsr', name: 'Быстрое питание', description: 'Для QSR', r_thresholds: [5, 10, 21, 45] as [number, number, number, number], f_thresholds: [10, 6, 3, 2] as [number, number, number, number] },
  { key: 'tsr', name: 'Ресторан', description: 'Для ресторанов', r_thresholds: [14, 30, 60, 120] as [number, number, number, number], f_thresholds: [6, 4, 2, 1] as [number, number, number, number] },
  { key: 'bar', name: 'Бар', description: 'Для баров', r_thresholds: [7, 14, 30, 60] as [number, number, number, number], f_thresholds: [6, 4, 2, 1] as [number, number, number, number] },
]

const mockSetMutateAsync = vi.fn().mockResolvedValue({})

function setupMocks(overrides: Record<string, unknown> = {}) {
  vi.mocked(useRFMTemplatesQuery).mockReturnValue({
    data: overrides.templates ?? mockTemplates,
    isLoading: overrides.isLoading ?? false,
    isError: overrides.isError ?? false,
    error: undefined,
    isValidating: false,
    mutate: vi.fn(),
  } as unknown as ReturnType<typeof useRFMTemplatesQuery>)

  vi.mocked(useRFMActiveTemplateQuery).mockReturnValue({
    data: overrides.activeTemplate ?? {
      active_template_type: 'standard',
      active_template_key: 'coffeegng',
      template: mockTemplates[0],
    },
    isLoading: false,
    isError: false,
    error: undefined,
    isValidating: false,
    mutate: vi.fn(),
  } as unknown as ReturnType<typeof useRFMActiveTemplateQuery>)

  vi.mocked(useRFMSetTemplateMutation).mockReturnValue({
    data: undefined,
    mutate: vi.fn(),
    mutateAsync: overrides.setMutateAsync ?? mockSetMutateAsync,
    trigger: overrides.setMutateAsync ?? mockSetMutateAsync,
    isPending: false,
    isMutating: false,
    isError: false,
    isSuccess: false,
    error: undefined,
    reset: vi.fn(),
  } as unknown as ReturnType<typeof useRFMSetTemplateMutation>)
}

describe('RFMTemplatePage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders 4 standard template cards', () => {
    setupMocks()
    render(<RFMTemplatePage />)

    expect(screen.getByText('Кофейня / Go-to')).toBeInTheDocument()
    expect(screen.getByText('Быстрое питание')).toBeInTheDocument()
    expect(screen.getByText('Ресторан')).toBeInTheDocument()
    expect(screen.getByText('Бар')).toBeInTheDocument()
  })

  it('highlights active template with check icon', () => {
    setupMocks()
    const { container } = render(<RFMTemplatePage />)

    // Active template (coffeegng) should have the ring/border style
    const activeCard = screen.getByText('Кофейня / Go-to').closest('button')
    expect(activeCard?.className).toContain('ring')
  })

  it('shows confirm message on first click', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    // Click a non-active template
    await user.click(screen.getByText('Быстрое питание').closest('button')!)

    expect(screen.getByText('Нажмите ещё раз для подтверждения')).toBeInTheDocument()
  })

  it('saves template on double click (confirm)', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    const btn = screen.getByText('Быстрое питание').closest('button')!

    // First click → confirm
    await user.click(btn)
    expect(screen.getByText('Нажмите ещё раз для подтверждения')).toBeInTheDocument()

    // Second click → save
    await user.click(btn)

    await waitFor(() => {
      expect(mockSetMutateAsync).toHaveBeenCalledWith({
        template_type: 'standard',
        template_key: 'qsr',
      })
    })
  })

  it('switches to custom mode tab', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    await user.click(screen.getByText('Вручную'))

    expect(screen.getByText('Название шаблона')).toBeInTheDocument()
    expect(screen.getByText(/Recency/)).toBeInTheDocument()
    expect(screen.getByText(/Frequency/)).toBeInTheDocument()
  })

  it('shows validation errors for invalid R thresholds (non-ascending)', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    // Switch to custom mode
    await user.click(screen.getByText('Вручную'))

    // Find R5 input (first threshold input) and set invalid values
    const rInputs = screen.getAllByRole('spinbutton').slice(0, 4)

    // Set R5=50, R4=10 (non-ascending)
    await user.clear(rInputs[0])
    await user.type(rInputs[0], '50')
    await user.clear(rInputs[1])
    await user.type(rInputs[1], '10')

    // Click save
    await user.click(screen.getByText('Сохранить шаблон'))

    expect(screen.getByText(/Recency: пороги должны строго возрастать/)).toBeInTheDocument()
  })

  it('shows validation errors for invalid F thresholds (non-descending)', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    await user.click(screen.getByText('Вручную'))

    const fInputs = screen.getAllByRole('spinbutton').slice(4, 8)

    // Set F5=2, F4=5 (non-descending — should be F5 > F4)
    await user.clear(fInputs[0])
    await user.type(fInputs[0], '2')
    await user.clear(fInputs[1])
    await user.type(fInputs[1], '5')

    await user.click(screen.getByText('Сохранить шаблон'))

    expect(screen.getByText(/Frequency: пороги должны строго убывать/)).toBeInTheDocument()
  })

  it('saves valid custom template', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    await user.click(screen.getByText('Вручную'))

    // Fill name
    const nameInput = screen.getByPlaceholderText('Мой шаблон')
    await user.clear(nameInput)
    await user.type(nameInput, 'Тест')

    // Defaults are [7, 14, 30, 60] for R and [8, 5, 3, 2] for F — these are valid
    await user.click(screen.getByText('Сохранить шаблон'))

    await waitFor(() => {
      expect(mockSetMutateAsync).toHaveBeenCalledWith({
        template_type: 'custom',
        custom_name: 'Тест',
        r_thresholds: [7, 14, 30, 60],
        f_thresholds: [8, 5, 3, 2],
      })
    })
  })

  it('shows error state when templates fail to load', () => {
    setupMocks({ isError: true, templates: undefined })
    render(<RFMTemplatePage />)

    expect(screen.getByText('Не удалось загрузить шаблоны')).toBeInTheDocument()
  })

  it('navigates back to dashboard on back button', async () => {
    setupMocks()
    const user = userEvent.setup()
    render(<RFMTemplatePage />)

    await user.click(screen.getByText('RFM-сегментация'))

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/rfm')
  })
})
