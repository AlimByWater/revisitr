import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/bots/queries', () => ({
  useBotQuery: vi.fn(),
}))

vi.mock('@/features/pos/queries', () => ({
  usePOSQuery: vi.fn(),
}))

vi.mock('@/features/loyalty/queries', () => ({
  useProgramsQuery: vi.fn(),
}))

vi.mock('@/features/bots/api', () => ({
  botsApi: {
    updateSettings: vi.fn(),
    update: vi.fn(),
  },
}))

vi.mock('@/features/menus/api', () => ({
  menusApi: {
    getBotPOSLocations: vi.fn(),
    setBotPOSLocations: vi.fn(),
  },
}))

vi.mock('@/features/campaigns/api', () => ({
  campaignsApi: {
    uploadFile: vi.fn(),
  },
}))

vi.mock('@/features/telegram-preview', () => ({
  TelegramPreview: () => <div data-testid="telegram-preview" />,
  MessageContentEditor: ({ value, onChange }: { value: any; onChange: (value: any) => void }) => (
    <button type="button" onClick={() => onChange(value)}>
      mock-message-editor
    </button>
  ),
}))

import { useBotQuery } from '@/features/bots/queries'
import { usePOSQuery } from '@/features/pos/queries'
import { useProgramsQuery } from '@/features/loyalty/queries'
import { menusApi } from '@/features/menus/api'
import BotDetailPage from './$botId'

const mockUseBotQuery = vi.mocked(useBotQuery)
const mockUsePOSQuery = vi.mocked(usePOSQuery)
const mockUseProgramsQuery = vi.mocked(useProgramsQuery)
const mockGetBotPOSLocations = vi.mocked(menusApi.getBotPOSLocations)

const botFixture = {
  id: 1,
  org_id: 1,
  name: 'Baratie',
  username: '',
  status: 'pending' as const,
  created_at: '2026-04-18T00:00:00Z',
  updated_at: '2026-04-18T00:00:00Z',
  program_id: undefined,
  created_by_telegram_id: undefined,
  settings: {
    modules: ['feedback', 'marketplace'],
    buttons: [
      { label: 'Меню', type: 'text', value: 'System', managed_by_module: 'menu', is_system: true },
      { label: 'Кастом', type: 'text', value: 'hello' },
    ],
    registration_form: [
      { name: 'phone', label: 'Телефон', type: 'phone', required: true },
      { name: 'favorite_drink', label: 'Любимый напиток', type: 'text', required: false },
    ],
    welcome_message: '',
    module_configs: {
      feedback: {
        prompt_message: 'Напишите ваш вопрос:',
        success_message: 'Ваше сообщение отправлено.',
      },
    },
    pos_selector_enabled: false,
    contacts_pos_ids: [],
  },
}

function renderPage(initialEntry = '/dashboard/bots/1') {
  return render(
    <MemoryRouter initialEntries={[initialEntry]}>
      <Routes>
        <Route path="/dashboard/bots/:botId" element={<BotDetailPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('BotDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()

    mockUseBotQuery.mockReturnValue({
      data: botFixture,
      isLoading: false,
      isError: false,
      error: undefined,
      mutate: vi.fn(),
      isValidating: false,
    } as unknown as ReturnType<typeof useBotQuery>)

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

    mockUseProgramsQuery.mockReturnValue({
      data: [
        {
          id: 3,
          org_id: 1,
          name: 'Бонусы',
          type: 'bonus',
          config: { welcome_bonus: 100, currency_name: 'баллы' },
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
    } as unknown as ReturnType<typeof useProgramsQuery>)
  })

  it('shows readiness banner when required setup is missing', async () => {
    mockGetBotPOSLocations.mockResolvedValue({ pos_ids: [] })

    renderPage('/dashboard/bots/1')

    expect(await screen.findByText('Для запуска бота не хватает настроек')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Привяжите точку продаж' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Выберите ссылку' })).toBeInTheDocument()
  })

  it('hides marketplace and renames feedback module to "Связаться"', async () => {
    mockGetBotPOSLocations.mockResolvedValue({ pos_ids: [10] })

    renderPage('/dashboard/bots/1?tab=modules')

    await waitFor(() => {
      expect(screen.getAllByText('Связаться').length).toBeGreaterThan(0)
    })

    expect(screen.queryByText('Маркетплейс')).not.toBeInTheDocument()
    expect(screen.getByPlaceholderText('Напишите ваш вопрос:')).toBeInTheDocument()
  })

  it('keeps standard field internal names fixed and blocks duplicate presets', async () => {
    mockGetBotPOSLocations.mockResolvedValue({ pos_ids: [10] })

    renderPage('/dashboard/bots/1?tab=general')

    expect(await screen.findByText('Поля регистрации')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '+ Пол' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: '+ Телефон' })).toBeDisabled()
    expect(screen.getByText('phone')).toBeInTheDocument()
    expect(screen.queryByPlaceholderText('Внутреннее имя, например phone')).not.toBeInTheDocument()
    expect(screen.getByPlaceholderText('Внутреннее имя, например favorite_drink')).toBeInTheDocument()
  })
})
