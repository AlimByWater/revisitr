import { Link, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { getApiErrorMessage } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { Eye, EyeOff } from 'lucide-react'

export default function RegisterPage() {
  const navigate = useNavigate()
  const register = useAuthStore((s) => s.register)

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [organization, setOrganization] = useState('')
  const [phone, setPhone] = useState('')
  const [showPassword, setShowPassword] = useState(false)
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
      navigate('/dashboard')
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Не удалось зарегистрироваться. Попробуйте позже.'))
    } finally {
      setIsLoading(false)
    }
  }

  const inputClass =
    'w-full px-0 py-3 bg-transparent text-base text-neutral-900 placeholder:text-neutral-400 border-0 border-b border-neutral-300 focus:outline-none focus:border-neutral-900 transition-colors duration-200 disabled:opacity-50 disabled:cursor-not-allowed'

  return (
    <div className="min-h-screen flex items-center justify-center bg-white px-6 py-12">
      <div className="w-full max-w-[420px]">
        {/* Logo */}
        <div className="text-center mb-12">
          <h1 className="text-5xl font-bold tracking-tight text-neutral-900 select-none">
            rev<span className="font-bold">/</span>sitr
          </h1>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="name" className="sr-only">
              Имя и фамилия
            </label>
            <input
              id="name"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Имя и фамилия"
              required
              autoFocus
              autoComplete="name"
              disabled={isLoading}
              className={inputClass}
            />
          </div>

          <div>
            <label htmlFor="email" className="sr-only">
              Email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Email"
              required
              autoComplete="email"
              disabled={isLoading}
              className={inputClass}
            />
          </div>

          <div className="relative">
            <label htmlFor="password" className="sr-only">
              Пароль
            </label>
            <input
              id="password"
              type={showPassword ? 'text' : 'password'}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Пароль (минимум 6 символов)"
              required
              minLength={6}
              autoComplete="new-password"
              disabled={isLoading}
              className={inputClass + ' pr-10'}
            />
            <button
              type="button"
              onClick={() => setShowPassword(!showPassword)}
              className="absolute right-0 top-1/2 -translate-y-1/2 p-1 text-neutral-400 hover:text-neutral-900 transition-colors duration-200"
              aria-label={showPassword ? 'Скрыть пароль' : 'Показать пароль'}
              tabIndex={-1}
            >
              {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
            </button>
          </div>

          <div>
            <label htmlFor="organization" className="sr-only">
              Название заведения
            </label>
            <input
              id="organization"
              type="text"
              value={organization}
              onChange={(e) => setOrganization(e.target.value)}
              placeholder="Название заведения"
              required
              disabled={isLoading}
              className={inputClass}
            />
          </div>

          <div>
            <label htmlFor="phone" className="sr-only">
              Телефон
            </label>
            <input
              id="phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="Телефон (необязательно)"
              autoComplete="tel"
              disabled={isLoading}
              className={inputClass}
            />
          </div>

          {error && (
            <div className="flex items-center gap-2.5 px-4 py-3 rounded bg-red-50 border border-red-100">
              <span className="text-red-500 font-mono text-xs shrink-0">✕</span>
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}

          <button
            type="submit"
            disabled={isLoading}
            className="w-full py-3.5 px-4 rounded bg-neutral-900 text-white text-base font-medium hover:bg-neutral-700 active:bg-[#EF3219] transition-colors duration-300 focus:outline-none disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Регистрация...' : 'Создать аккаунт'}
          </button>
        </form>

        <p className="text-center text-sm text-neutral-500 mt-8">
          Уже есть аккаунт?{' '}
          <Link
            to="/auth/login"
            className="text-[#EF3219] underline font-medium"
          >
            Войти
          </Link>
        </p>
      </div>
    </div>
  )
}
