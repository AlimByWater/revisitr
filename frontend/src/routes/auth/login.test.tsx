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
import LoginPage from './login'

const loginMock = vi.fn()

function mockStore() {
  vi.mocked(useAuthStore).mockImplementation((selector: any) => selector({ login: loginMock }))
}

describe('LoginPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockStore()
  })

  it('submits credentials and redirects to dashboard on success', async () => {
    loginMock.mockResolvedValue(undefined)
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('Email'), 'owner@example.com')
    await user.type(screen.getByLabelText('Пароль'), 'secret123')
    await user.click(screen.getByRole('button', { name: 'Войти' }))

    await waitFor(() => {
      expect(loginMock).toHaveBeenCalledWith({ email: 'owner@example.com', password: 'secret123' })
    })
    expect(mockNavigate).toHaveBeenCalledWith('/dashboard')
  })

  it('shows backend error message and does not redirect on failure', async () => {
    loginMock.mockRejectedValue({ response: { data: { message: 'Неверный пароль' } } })
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    )

    await user.type(screen.getByLabelText('Email'), 'owner@example.com')
    await user.type(screen.getByLabelText('Пароль'), 'wrongpass')
    await user.click(screen.getByRole('button', { name: 'Войти' }))

    expect(await screen.findByText('Неверный пароль')).toBeInTheDocument()
    expect(mockNavigate).not.toHaveBeenCalled()
  })

  it('toggles password visibility', async () => {
    const user = userEvent.setup()

    render(
      <MemoryRouter>
        <LoginPage />
      </MemoryRouter>,
    )

    const passwordInput = screen.getByLabelText('Пароль')
    expect(passwordInput).toHaveAttribute('type', 'password')

    await user.click(screen.getByRole('button', { name: 'Показать пароль' }))
    expect(passwordInput).toHaveAttribute('type', 'text')
  })
})
