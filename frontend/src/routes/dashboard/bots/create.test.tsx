import { act, fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/features/bots/api', () => ({
  botsApi: {
    createManaged: vi.fn(),
    getBotStatus: vi.fn(),
    create: vi.fn(),
  },
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

import { MemoryRouter } from 'react-router-dom'
import { botsApi } from '@/features/bots/api'
import CreateBotPage from './create'

const mockCreateManaged = vi.mocked(botsApi.createManaged)
const mockGetBotStatus = vi.mocked(botsApi.getBotStatus)
const mockCreate = vi.mocked(botsApi.create)

function renderPage() {
  return render(
    <MemoryRouter>
      <CreateBotPage />
    </MemoryRouter>,
  )
}

async function completeStep1(
  user: ReturnType<typeof userEvent.setup>,
  overrides: { name?: string; username?: string; description?: string } = {},
) {
  await user.type(screen.getByPlaceholderText('Мой ресторан'), overrides.name ?? 'Baratie')
  await user.type(screen.getByPlaceholderText('myrestaurant_bot'), overrides.username ?? 'baratiebot')

  if (overrides.description) {
    await user.type(screen.getByPlaceholderText('Описание бота (видно в профиле)'), overrides.description)
  }

  await user.click(screen.getByRole('button', { name: /^Далее$/ }))
}

async function openPendingState(
  user: ReturnType<typeof userEvent.setup>,
  overrides: { name?: string; username?: string; description?: string } = {},
) {
  await completeStep1(user, overrides)
  await user.click(screen.getByRole('button', { name: /^Пропустить$/ }))
  await user.click(screen.getByRole('button', { name: /Создать бота/i }))
  await screen.findByText('Перейдите в Telegram для подтверждения')
}

describe('CreateBotPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useRealTimers()

    mockCreateManaged.mockResolvedValue({
      bot_id: 77,
      deep_link: 'https://t.me/newbot?start=baratie',
      status: 'pending',
    })
    mockGetBotStatus.mockResolvedValue({ status: 'pending' })
    mockCreate.mockResolvedValue({
      id: 99,
      org_id: 1,
      name: 'Baratie',
      username: 'baratiebot',
      status: 'active',
      settings: {
        modules: [],
        buttons: [],
        registration_form: [],
        welcome_message: '',
      },
      created_at: '2026-04-14T00:00:00Z',
      updated_at: '2026-04-14T00:00:00Z',
    })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('validates basic wizard step and sanitizes username input', async () => {
    const user = userEvent.setup()
    renderPage()

    const nextButton = screen.getByRole('button', { name: /^Далее$/ })
    const usernameInput = screen.getByPlaceholderText('myrestaurant_bot')

    expect(nextButton).toBeDisabled()

    await user.type(screen.getByPlaceholderText('Мой ресторан'), 'Бар')
    await user.type(usernameInput, 'bot')

    expect(screen.getByText('Минимум 5 символов')).toBeInTheDocument()
    expect(nextButton).toBeDisabled()

    await user.clear(usernameInput)
    await user.type(usernameInput, 'bad-name')

    expect(usernameInput).toHaveValue('badname')
    expect(screen.getByText('Username должен заканчиваться на "bot"')).toBeInTheDocument()
    expect(nextButton).toBeDisabled()

    await user.clear(usernameInput)
    await user.type(usernameInput, 'brand_new_bot')

    expect(usernameInput).toHaveValue('brand_new_bot')
    expect(screen.queryByText('Минимум 5 символов')).not.toBeInTheDocument()
    expect(screen.queryByText('Username должен заканчиваться на "bot"')).not.toBeInTheDocument()
    expect(nextButton).toBeEnabled()
  })

  it('submits managed bot payload from wizard state', async () => {
    const user = userEvent.setup()
    renderPage()

    await completeStep1(user, {
      name: '  Baratie  ',
      username: 'baratiebot',
      description: '  Лучший бот для гостей  ',
    })

    await user.type(
      screen.getByPlaceholderText('Сообщение при первом запуске бота клиентом'),
      '  Добро пожаловать!  ',
    )
    await user.click(screen.getByRole('button', { name: /Добавить/i }))

    const fieldInputs = screen.getAllByPlaceholderText('Название поля')
    await user.type(fieldInputs[fieldInputs.length - 1], 'Favorite drink')
    await user.click(screen.getByRole('button', { name: /^Далее$/ }))

    await user.click(screen.getByRole('checkbox', { name: /Отзывы/i }))
    await user.click(screen.getByRole('button', { name: /Создать бота/i }))

    await waitFor(() => {
      expect(mockCreateManaged).toHaveBeenCalledWith({
        name: 'Baratie',
        username: 'baratiebot',
        description: 'Лучший бот для гостей',
        welcome_message: 'Добро пожаловать!',
        registration_form: [
          { name: 'first_name', label: 'Имя', type: 'text', required: true },
          { name: 'phone', label: 'Телефон', type: 'phone', required: true },
          { name: 'favorite_drink', label: 'Favorite drink', type: 'text', required: false },
        ],
        modules: ['loyalty', 'feedback'],
      })
    })

    expect(screen.getByRole('link', { name: /Создать в Telegram/i })).toHaveAttribute(
      'href',
      'https://t.me/newbot?start=baratie',
    )
  })

  it('polls bot status until active and redirects to bot page', async () => {
    vi.useFakeTimers()
    mockGetBotStatus
      .mockResolvedValueOnce({ status: 'pending' })
      .mockResolvedValueOnce({ status: 'active' })

    renderPage()

    fireEvent.change(screen.getByPlaceholderText('Мой ресторан'), { target: { value: 'Baratie' } })
    fireEvent.change(screen.getByPlaceholderText('myrestaurant_bot'), { target: { value: 'baratiebot' } })
    fireEvent.click(screen.getByRole('button', { name: /^Далее$/ }))
    fireEvent.click(screen.getByRole('button', { name: /^Пропустить$/ }))
    fireEvent.click(screen.getByRole('button', { name: /Создать бота/i }))

    await act(async () => {
      await Promise.resolve()
    })

    expect(screen.getByText('Перейдите в Telegram для подтверждения')).toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })
    expect(mockGetBotStatus).toHaveBeenCalledTimes(1)
    expect(screen.getByText('Ожидаем подтверждения...')).toBeInTheDocument()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000)
    })

    expect(screen.getByText('Бот создан!')).toBeInTheDocument()
    expect(mockGetBotStatus).toHaveBeenCalledTimes(2)

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1500)
    })

    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/bots/77')
  }, 10000)

  it('allows manual token fallback flow from pending state', async () => {
    const user = userEvent.setup()
    renderPage()

    await openPendingState(user, { name: 'Fallback bot', username: 'fallbackbot' })

    await user.click(screen.getByRole('button', { name: /У меня уже есть бот/i }))
    await user.type(screen.getByPlaceholderText('Вставьте токен от @BotFather'), '123456:ABCDEF')
    await user.click(screen.getByRole('button', { name: /Подключить бота/i }))

    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith({
        name: 'Fallback bot',
        token: '123456:ABCDEF',
      })
    })
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard/bots/99')
  })

  it('shows retry action when managed bot creation fails', async () => {
    mockCreateManaged
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce({
        bot_id: 88,
        deep_link: 'https://t.me/newbot?start=retry',
        status: 'pending',
      })

    const user = userEvent.setup()
    renderPage()

    await completeStep1(user, { name: 'Retry bot', username: 'retrybot' })
    await user.click(screen.getByRole('button', { name: /^Пропустить$/ }))
    await user.click(screen.getByRole('button', { name: /Создать бота/i }))

    await screen.findByText('Ошибка создания бота. Попробуйте снова.')
    expect(mockCreateManaged).toHaveBeenCalledTimes(1)

    await user.click(screen.getByRole('button', { name: /Повторить/i }))

    await waitFor(() => {
      expect(mockCreateManaged).toHaveBeenCalledTimes(2)
    })
    expect(await screen.findByText('Перейдите в Telegram для подтверждения')).toBeInTheDocument()
  })

  it('surfaces fallback token errors without redirecting', async () => {
    mockCreate.mockRejectedValueOnce(new Error('invalid token'))

    const user = userEvent.setup()
    renderPage()

    await openPendingState(user, { name: 'Broken fallback', username: 'brokenbot' })

    await user.click(screen.getByRole('button', { name: /У меня уже есть бот/i }))
    await user.type(screen.getByPlaceholderText('Вставьте токен от @BotFather'), 'bad-token')
    await user.click(screen.getByRole('button', { name: /Подключить бота/i }))

    await screen.findByText('Не удалось создать бота. Проверьте токен.')
    expect(mockNavigate).not.toHaveBeenCalled()
  })
})
