import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

export const Route = createFileRoute('/auth/register')({
  component: RegisterPage,
})

function RegisterPage() {
  const router = useRouter()
  const register = useAuthStore((s) => s.register)

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [organization, setOrganization] = useState('')
  const [phone, setPhone] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setIsLoading(true)

    try {
      await register({
        name,
        email,
        password,
        organization,
        phone: phone || undefined,
      })
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
        setError('Не удалось зарегистрироваться. Попробуйте позже.')
      }
    } finally {
      setIsLoading(false)
    }
  }

  const inputClassName = cn(
    'w-full px-4 py-2.5 rounded-lg border border-surface-border',
    'text-sm placeholder:text-neutral-400',
    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
    'transition-colors',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  )

  return (
    <div className="min-h-screen flex items-center justify-center px-4 py-8">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold tracking-tight">
            revi<span className="text-accent">s</span>itr
          </h1>
          <p className="text-neutral-500 mt-2 text-sm">
            Создайте аккаунт для управления лояльностью
          </p>
        </div>

        <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-8">
          <form onSubmit={handleSubmit} className="space-y-5">
            <div>
              <label
                htmlFor="name"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Имя
              </label>
              <input
                id="name"
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Иван Иванов"
                required
                autoComplete="name"
                disabled={isLoading}
                className={inputClassName}
              />
            </div>

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
                className={inputClassName}
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
                placeholder="Минимум 6 символов"
                required
                minLength={6}
                autoComplete="new-password"
                disabled={isLoading}
                className={inputClassName}
              />
            </div>

            <div>
              <label
                htmlFor="organization"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Название организации
              </label>
              <input
                id="organization"
                type="text"
                value={organization}
                onChange={(e) => setOrganization(e.target.value)}
                placeholder="Ресторан 'Уют'"
                required
                disabled={isLoading}
                className={inputClassName}
              />
            </div>

            <div>
              <label
                htmlFor="phone"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Телефон{' '}
                <span className="text-neutral-400 font-normal">(необязательно)</span>
              </label>
              <input
                id="phone"
                type="tel"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+7 (999) 123-45-67"
                autoComplete="tel"
                disabled={isLoading}
                className={inputClassName}
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
              {isLoading ? 'Регистрация...' : 'Создать аккаунт'}
            </button>
          </form>

          <p className="text-center text-sm text-neutral-500 mt-6">
            Уже есть аккаунт?{' '}
            <Link to="/auth/login" className="text-accent hover:underline font-medium">
              Войти
            </Link>
          </p>
        </div>
      </div>
    </div>
  )
}
