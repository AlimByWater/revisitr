import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/bots/queries', () => ({
  useBotQuery: vi.fn(),
}))

vi.mock('@/features/bots/api', () => ({
  botsApi: {
    updateSettings: vi.fn(),
  },
}))

import { useBotQuery } from '@/features/bots/queries'
import { botsApi } from '@/features/bots/api'
import BotFeedbackSettingsPage from './feedback'

const mockUseBotQuery = vi.mocked(useBotQuery)
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
    modules: ['feedback'],
    buttons: [],
    registration_form: [],
    module_configs: {
      feedback: {
        prompt_message: 'Напишите ваш вопрос:',
        success_message: 'Ваше сообщение отправлено.',
      },
    },
  },
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/dashboard/bots/1/feedback']}>
      <Routes>
        <Route path="/dashboard/bots/:botId/feedback" element={<BotFeedbackSettingsPage />} />
      </Routes>
    </MemoryRouter>,
  )
}

describe('BotFeedbackSettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockUseBotQuery.mockReturnValue({
      data: botFixture,
      isLoading: false,
      isError: false,
      mutate: vi.fn(),
    } as unknown as ReturnType<typeof useBotQuery>)
    mockUpdateSettings.mockResolvedValue()
  })

  it('saves feedback module messages', async () => {
    const user = userEvent.setup()
    renderPage()

    const prompt = await screen.findByPlaceholderText('Напишите ваш вопрос:')
    await user.clear(prompt)
    await user.type(prompt, 'Оставьте сообщение')

    await user.click(screen.getByRole('button', { name: 'Сохранить изменения' }))

    await waitFor(() => {
      expect(mockUpdateSettings).toHaveBeenCalledWith(
        1,
        expect.objectContaining({
          module_configs: expect.objectContaining({
            feedback: expect.objectContaining({
              prompt_message: 'Оставьте сообщение',
              success_message: 'Ваше сообщение отправлено.',
            }),
          }),
        }),
      )
    })
  })
})
