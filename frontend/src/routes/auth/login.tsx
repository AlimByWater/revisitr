import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

export const Route = createFileRoute('/auth/login')({
  component: LoginPage,
})

function LoginPage() {
  const router = useRouter()
  const login = useAuthStore((s) => s.login)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    try {
      await login({ email, password })
      router.navigate({ to: '/dashboard' })
    } catch (err: unknown) {
      if (
        err &&
        typeof err === 'object' &&
        'response' in err &&
        err.response &&
        typeof err.response === 'object' &&
        'data' in err.response &&
        err.response.data &&
        typeof err.response.data === 'object' &&
        'message' in err.response.data
      ) {
        setError(String((err.response as { data: { message: string } }).data.message))
      } else {
        setError('Не удалось войти. Проверьте email и пароль.')
      }
    } finally {
      setIsLoading(false)
    }
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
                disabled={isLoading}
                className={cn(
                  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                  'text-sm placeholder:text-neutral-400',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
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
                minLength={6}
                autoComplete="current-password"
                disabled={isLoading}
                className={cn(
                  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                  'text-sm placeholder:text-neutral-400',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                )}
              />
            </div>

            {error && (
              <p className="text-sm text-red-600">{error}</p>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className={cn(
                'w-full py-2.5 px-4 rounded-lg',
                'bg-neutral-900 text-white text-sm font-medium',
                'hover:bg-neutral-800 active:bg-neutral-950',
                'transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-neutral-900/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {isLoading ? 'Вход...' : 'Войти'}
            </button>
          </form>

          <p className="text-center text-sm text-neutral-500 mt-6">
            Нет аккаунта?{' '}
            <Link to="/auth/register" className="text-accent hover:underline font-medium">
              Регистрация
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
