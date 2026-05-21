import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/bots/queries', () => ({
  useBotQuery: vi.fn(),
}))

vi.mock('@/features/pos/queries', () => ({
  usePOSQuery: vi.fn(),
}))

vi.mock('@/features/menus/api', () => ({
  menusApi: {
    getBotPOSLocations: vi.fn(),
  },
}))

vi.mock('@/features/bots/api', () => ({
  botsApi: {
    updateSettings: vi.fn(),
  },
}))

vi.mock('@/features/campaigns/api', () => ({
  campaignsApi: {
    uploadFile: vi.fn(),
  },
}))

vi.mock('@/features/telegram-preview', () => ({
  MessageContentEditor: ({ value, onChange }: { value: any; onChange: (value: any) => void }) => (
    <button type="button" onClick={() => onChange(value)}>
      mock-message-editor
    </button>
  ),
}))

import { useBotQuery } from '@/features/bots/queries'
import { usePOSQuery } from '@/features/pos/queries'
import { menusApi } from '@/features/menus/api'
import { botsApi } from '@/features/bots/api'
import BotBookingSettingsPage from './booking'

const mockUseBotQuery = vi.mocked(useBotQuery)
const mockUsePOSQuery = vi.mocked(usePOSQuery)
const mockGetBotPOSLocations = vi.mocked(menusApi.getBotPOSLocations)
const mockUpdateSettings = vi.mocked(botsApi.updateSettings)

const botFixture = {
  id: 1,
  org_id: 1,
  name: 'Baratie',
  username: 'baratie_bot',
  status: 'active' as const,
  created_at: '2026-04-18T00:00:00Z',
  updated_at: '2026-04-18T00:00:00Z',
  settings: {
    modules: ['booking'],
    buttons: [],
    registration_form: [{ name: 'first_name', label: 'Как вас зовут?', type: 'text', required: true }],
    module_configs: {
      booking: {
        date_from_days: 1,
        date_to_days: 7,
        time_slots: [{ start: '10:00', end: '11:00' }],
        party_size_options: ['1', '2'],
        pos_ids: [10],
      },
    },
  },
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/dashboard/bots/1/booking']}>
      <Routes>
        <Route path="/dashboard/bots/:botId/booking" element={<BotBookingSettingsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('BotBookingSettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseBotQuery.mockReturnValue({
      data: botFixture,
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useBotQuery>)
    mockUsePOSQuery.mockReturnValue({
      data: [{ id: 10, org_id: 1, name: 'Маросейка', address: '', phone: '', schedule: {}, is_active: true }],
      isLoading: false,
      isError: false,
    } as unknown as ReturnType<typeof usePOSQuery>)
    mockGetBotPOSLocations.mockResolvedValue({ pos_ids: [10] })
    mockUpdateSettings.mockResolvedValue()
  })

  it('keeps number input focused while editing and saves numeric config', async () => {
    const user = userEvent.setup()
    renderPage()

    const fromInput = await screen.findByLabelText('Бронь доступна от')
    await user.clear(fromInput)
    await user.type(fromInput, '12')

    expect(fromInput).toHaveValue('12')
    expect(document.activeElement).toBe(fromInput)

    await user.click(screen.getByRole('button', { name: 'Сохранить изменения' }))

    await waitFor(() => {
      expect(mockUpdateSettings).toHaveBeenCalledWith(
        1,
        expect.objectContaining({
          module_configs: expect.objectContaining({
            booking: expect.objectContaining({ date_from_days: 12 }),
          }),
        }),
      )
    })
  })
})
