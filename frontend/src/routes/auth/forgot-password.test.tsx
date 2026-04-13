import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { describe, expect, it } from 'vitest'

import ForgotPasswordPage from './forgot-password'

describe('ForgotPasswordPage', () => {
  it('renders placeholder recovery message and login link', () => {
    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>,
    )

    expect(screen.getByText('Восстановление пароля')).toBeInTheDocument()
    expect(screen.getByText(/пока в разработке/i)).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Вернуться ко входу' })).toHaveAttribute('href', '/auth/login')
  })
})
