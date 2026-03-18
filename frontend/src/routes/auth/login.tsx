import { Link, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

// Restaurant floor plan — tables in a dining room
const FLOOR_PLAN = [
  '┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐',
  '│     │   │     │   │     │   │     │   │     │',
  '│  ·  │   │  ·  │   │  ·  │   │  ·  │   │  ·  │',
  '│     │   │     │   │     │   │     │   │     │',
  '└─────┘   └─────┘   └─────┘   └─────┘   └─────┘',
  '     ┌─────┐   ┌─────┐   ┌─────┐   ┌─────┐',
  '     │     │   │     │   │     │   │     │',
  '     │  ·  │   │  ·  │   │  ·  │   │  ·  │',
  '     │     │   │     │   │     │   │     │',
  '     └─────┘   └─────┘   └─────┘   └─────┘',
].join('\n')

const FEATURES = [
  { title: 'Telegram-боты', desc: '— для каждого заведения' },
  { title: 'Программы лояльности', desc: '— бонусные и скидочные' },
  { title: 'Рассылки', desc: '— ручные и авто-сценарии' },
  { title: 'Аналитика', desc: '— выручка, клиенты, отчёты' },
]

export default function LoginPage() {
  const navigate = useNavigate()
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
      navigate('/dashboard')
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

  const inputClass = cn(
    'w-full px-4 py-3 rounded-xl border border-surface-border bg-white',
    'text-sm placeholder:text-neutral-300',
    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
    'transition-all duration-200',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  )

  return (
    <div className="min-h-screen flex">
      {/* ─── Brand panel ─── */}
      <div className="hidden lg:flex lg:w-[48%] bg-sidebar relative overflow-hidden flex-col justify-between p-10 xl:p-14">
        {/* ASCII floor plan texture */}
        <div
          className="absolute inset-0 flex items-center justify-center overflow-hidden select-none pointer-events-none"
          aria-hidden="true"
        >
          <pre className="text-white/[0.035] text-[11px] font-mono leading-relaxed whitespace-pre">
            {[FLOOR_PLAN, FLOOR_PLAN, FLOOR_PLAN, FLOOR_PLAN, FLOOR_PLAN, FLOOR_PLAN].join('\n')}
          </pre>
        </div>

        {/* Content */}
        <div className="relative z-10">
          {/* Brand */}
          <div className="flex items-center gap-3 mb-20 animate-in">
            <span className="text-2xl font-bold text-white tracking-tight">
              revi<span className="text-accent">s</span>itr
            </span>
            <span className="text-[10px] font-bold text-accent bg-accent/10 px-1.5 py-0.5 rounded uppercase tracking-wider">
              PRO
            </span>
          </div>

          {/* Headline */}
          <div className="mb-10 animate-in animate-in-delay-1">
            <h2 className="font-serif text-[3rem] xl:text-[4rem] font-bold text-white leading-[1.08] tracking-tight">
              Управляйте<br />
              лояльностью<br />
              <span className="text-accent">гостей</span>
            </h2>
          </div>

          {/* Decorative divider */}
          <div
            className="font-mono text-accent/25 text-xs mb-8 tracking-[0.4em] animate-in animate-in-delay-2"
            aria-hidden="true"
          >
            ━━━━━━━━━━━━━━
          </div>

          {/* Features */}
          <div className="space-y-3.5 animate-in animate-in-delay-3">
            {FEATURES.map((f) => (
              <div key={f.title} className="flex items-start gap-3">
                <span className="text-accent font-mono text-sm mt-[3px]">▸</span>
                <p className="text-sm">
                  <span className="text-white/80 font-medium">{f.title}</span>
                  <span className="text-white/35 ml-1.5">{f.desc}</span>
                </p>
              </div>
            ))}
          </div>
        </div>

        {/* Footer */}
        <div className="relative z-10 animate-in animate-in-delay-4">
          <div className="font-mono text-white/[0.06] text-[10px] mb-3" aria-hidden="true">
            {'─'.repeat(50)}
          </div>
          <p className="text-white/20 text-[11px] font-mono uppercase tracking-wider">
            © 2026 Revisitr · Платформа для HoReCa
          </p>
        </div>
      </div>

      {/* ─── Form panel ─── */}
      <div className="flex-1 flex items-center justify-center px-6 py-12 bg-white">
        <div className="w-full max-w-[400px]">
          {/* Mobile brand */}
          <div className="lg:hidden text-center mb-10 animate-in">
            <h1 className="text-3xl font-bold tracking-tight">
              revi<span className="text-accent">s</span>itr
            </h1>
            <p className="text-neutral-400 mt-2 text-sm">
              Платформа лояльности для HoReCa
            </p>
          </div>

          {/* Form header */}
          <div className="mb-8 animate-in">
            <h2 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
              Вход
            </h2>
            <p className="text-neutral-400 mt-2 text-sm">
              Войдите в панель управления
            </p>
          </div>

          <form onSubmit={handleSubmit} className="space-y-5 animate-in animate-in-delay-1">
            <div>
              <label htmlFor="email" className="block text-sm font-medium text-neutral-700 mb-1.5">
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
                className={inputClass}
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-sm font-medium text-neutral-700 mb-1.5">
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
                className={inputClass}
              />
            </div>

            {error && (
              <div className="flex items-center gap-2.5 px-4 py-3 rounded-xl bg-red-50 border border-red-100">
                <span className="text-red-500 font-mono text-xs shrink-0">✕</span>
                <p className="text-sm text-red-600">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className={cn(
                'w-full py-3 px-4 rounded-xl',
                'bg-accent text-white text-sm font-semibold',
                'hover:bg-accent-hover active:bg-accent/80',
                'transition-all duration-150',
                'focus:outline-none focus:ring-2 focus:ring-accent/20',
                'shadow-md shadow-accent/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {isLoading ? 'Вход...' : 'Войти'}
            </button>
          </form>

          <p className="text-center text-sm text-neutral-400 mt-8 animate-in animate-in-delay-2">
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
