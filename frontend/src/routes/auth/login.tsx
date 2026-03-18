import { createFileRoute, Link } from '@tanstack/react-router'
import { useState } from 'react'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/auth/login')({
  component: LoginPage,
})

function LoginPage() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    // Auth logic will be implemented later
    console.log('Login:', { email, password })
  }

  return (
    <div className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold tracking-tight">
            revi<span className="text-accent">s</span>itr
          </h1>
          <p className="text-neutral-500 mt-2 text-sm">
            Войдите в панель управления
          </p>
        </div>

        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-8">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label
                htmlFor="email"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="admin@restaurant.com"
                required
                autoComplete="email"
                className={cn(
                  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                  'text-sm placeholder:text-neutral-400',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                )}
              />
            </div>

            <div>
              <label
                htmlFor="password"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Пароль
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Введите пароль"
                required
                autoComplete="current-password"
                className={cn(
                  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                  'text-sm placeholder:text-neutral-400',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                )}
              />
            </div>

            <button
              type="submit"
              className={cn(
                'w-full py-2.5 px-4 rounded-lg',
                'bg-neutral-900 text-white text-sm font-medium',
                'hover:bg-neutral-800 active:bg-neutral-950',
                'transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-neutral-900/20',
              )}
            >
              Войти
            </button>
          </form>

          <p className="text-center text-sm text-neutral-500 mt-6">
            Нет аккаунта?{' '}
            <Link to="/" className="text-accent hover:underline font-medium">
              Регистрация
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
