import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/stores/auth', () => ({
  useAuthStore: vi.fn(),
}))

const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual<typeof import('react-router-dom')>('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

import { useAuthStore } from '@/stores/auth'
import RegisterPage from './register'

const registerMock = vi.fn()

function mockStore() {
  vi.mocked(useAuthStore).mockImplementation((selector: any) => selector({ register: registerMock }))
}

describe('RegisterPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockStore()
  })

  it('submits registration form and redirects on success', async () => {
    registerMock.mockResolvedValue(undefined)
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('Имя и фамилия'), 'Иван Иванов')
    await user.type(screen.getByLabelText('Email'), 'ivan@example.com')
    await user.type(screen.getByLabelText('Пароль'), 'secret123')
    await user.type(screen.getByLabelText('Название заведения'), 'Baratie')
    await user.type(screen.getByLabelText(/Телефон/), '+79991234567')
    await user.click(screen.getByRole('button', { name: 'Создать аккаунт' }))

    await waitFor(() => {
      expect(registerMock).toHaveBeenCalledWith({
        name: 'Иван Иванов',
        email: 'ivan@example.com',
        password: 'secret123',
        organization: 'Baratie',
        phone: '+79991234567',
      })
    })
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
  })

  it('omits optional phone when field is empty', async () => {
    registerMock.mockResolvedValue(undefined)
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('Имя и фамилия'), 'Иван Иванов')
    await user.type(screen.getByLabelText('Email'), 'ivan@example.com')
    await user.type(screen.getByLabelText('Пароль'), 'secret123')
    await user.type(screen.getByLabelText('Название заведения'), 'Baratie')
    await user.click(screen.getByRole('button', { name: 'Создать аккаунт' }))

    await waitFor(() => {
      expect(registerMock).toHaveBeenCalledWith({
        name: 'Иван Иванов',
        email: 'ivan@example.com',
        password: 'secret123',
        organization: 'Baratie',
        phone: undefined,
      })
    })
  })

  it('shows fallback error message and stays on page when registration fails', async () => {
    registerMock.mockRejectedValue(new Error('network'))
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <RegisterPage />
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('Имя и фамилия'), 'Иван Иванов')
    await user.type(screen.getByLabelText('Email'), 'ivan@example.com')
    await user.type(screen.getByLabelText('Пароль'), 'secret123')
    await user.type(screen.getByLabelText('Название заведения'), 'Baratie')
    await user.click(screen.getByRole('button', { name: 'Создать аккаунт' }))

    expect(await screen.findByText('Не удалось зарегистрироваться. Попробуйте позже.')).toBeInTheDocument()
    expect(mockNavigate).not.toHaveBeenCalled()
  })
})
