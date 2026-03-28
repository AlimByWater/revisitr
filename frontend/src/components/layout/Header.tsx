import { Link, useNavigate, useLocation } from 'react-router-dom'
import { User, Menu, LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/contexts/ThemeContext'

interface HeaderProps {
  onMenuToggle?: () => void
}

const NAV_LINKS = [
  { label: 'Панель управления', href: '/dashboard', underline: true },
  { label: 'Тарифы', href: '#' },
  { label: 'Поддержка', href: '#' },
  { label: 'Маркетинг «под ключ»', href: '#', accent: true },
  { label: 'Контакты', href: '#' },
]

export function Header({ onMenuToggle }: HeaderProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const { theme } = useTheme()

  const handleLogout = () => {
    logout()
    navigate('/auth/login')
  }

  const isAurora = theme === 'aurora'

  if (isAurora) {
    return (
      <header className="h-16 border-b flex items-center justify-between px-4 md:px-6 sticky top-0 z-30 header-glass">
        <div className="flex items-center gap-4 md:gap-8">
          <button onClick={onMenuToggle} type="button" className="lg:hidden w-9 h-9 rounded flex items-center justify-center text-white/60 hover:text-white hover:bg-white/10" aria-label="Открыть меню">
            <Menu className="w-5 h-5" />
          </button>
          <Link to="/dashboard" className="text-xl font-bold tracking-tight text-white/90 select-none">rev/sitr</Link>
        </div>
        <div className="flex items-center gap-3">
          <button onClick={handleLogout} type="button" className="w-8 h-8 rounded flex items-center justify-center text-white/30 hover:text-red-400 hover:bg-red-500/10" title="Выйти" aria-label="Выйти">
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </header>
    )
  }

  const isDashboard = location.pathname.startsWith('/dashboard')

  return (
    <header className="px-4 sm:px-8 lg:px-16 pt-4 sm:pt-6 pb-0">
      <div className="flex items-center justify-between">
        {/* Left section: mirrors sidebar width on desktop, logo centered within */}
        <div className="flex items-center shrink-0 lg:w-[220px] lg:justify-center">
          <div className="flex items-center gap-3 sm:gap-4">
            <button
              onClick={onMenuToggle}
              type="button"
              className="lg:hidden w-9 h-9 rounded border border-neutral-900 flex items-center justify-center text-neutral-600 hover:text-neutral-900 hover:bg-neutral-50 transition-colors"
              aria-label="Открыть меню"
            >
              <Menu className="w-5 h-5" />
            </button>

            <Link to="/dashboard" className="shrink-0">
              <img
                src="/revisitr/logo.png"
                alt="Revisitr"
                className="h-6 sm:h-7 w-auto"
              />
            </Link>
          </div>
        </div>

        {/* Center: nav — matches content area start */}
        <nav className="hidden lg:flex items-center gap-6 flex-1 pl-6">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.label}
              to={link.href}
              className={cn(
                'text-sm whitespace-nowrap transition-colors',
                link.accent
                  ? 'text-[#EF3219] font-medium hover:text-[#FF5C47]'
                  : link.underline && isDashboard
                    ? 'text-neutral-900 font-bold'
                    : 'text-neutral-500 hover:text-neutral-900',
              )}
            >
              {link.label}
            </Link>
          ))}
        </nav>

        {/* Right: profile */}
        <div className="flex items-center gap-2 sm:gap-3 shrink-0">
          <div className="w-9 h-9 sm:w-10 sm:h-10 rounded-full border border-neutral-900 flex items-center justify-center">
            <User className="w-4 h-4 sm:w-5 sm:h-5 text-neutral-900" />
          </div>
          <div className="hidden sm:flex flex-col">
            <span className="text-sm font-bold text-neutral-900 leading-tight uppercase">
              Gennady P.
            </span>
            <span className="text-[11px] font-bold text-[#EF3219] leading-tight">
              pro
            </span>
          </div>
        </div>
      </div>

      {/* Divider line */}
      <div className="mt-4 border-b border-neutral-900" />
    </header>
  )
}
