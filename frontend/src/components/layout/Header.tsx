import { useState, useRef, useEffect } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { User, Menu, LogOut, Settings } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/contexts/ThemeContext'

interface HeaderProps {
  onMenuToggle?: () => void
}

export function Header({ onMenuToggle }: HeaderProps) {
  const navigate = useNavigate()
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const user = useAuthStore((s) => s.user)
  const { theme, toggleTheme } = useTheme()
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  const handleLogout = async () => {
    setMenuOpen(false)
    await logout()
    navigate('/auth/login')
  }

  const isAurora = theme === 'aurora'
  const isDashboard = location.pathname === '/dashboard' || location.pathname === '/dashboard/'

  // Close dropdown on outside click
  useEffect(() => {
    if (!menuOpen) return
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [menuOpen])

  if (isAurora) {
    return (
      <header
        className="h-16 border-b flex items-center justify-between px-4 md:px-6 sticky top-0 z-30 transition-all duration-500 header-glass"
      >
        <div className="flex items-center gap-4 md:gap-8">
          {/* Mobile burger */}
          <button
            onClick={onMenuToggle}
            type="button"
            className="lg:hidden h-11 w-11 rounded-lg flex items-center justify-center transition-colors text-white/60 hover:text-white hover:bg-white/10"
            aria-label="Открыть меню"
          >
            <Menu className="w-5 h-5" />
          </button>

          <Link to="/dashboard" className="text-xl font-bold tracking-tight group/logo select-none">
            <span className="text-white/90">
              revi
            </span>
            <span
              className="transition-all duration-300 text-violet-400 group-hover/logo:drop-shadow-[0_0_8px_rgba(139,92,246,0.6)]"
            >
              s
            </span>
            <span className="text-white/90">
              itr
            </span>
          </Link>

          <nav className="hidden md:flex items-center gap-6">
            <Link
              to="/dashboard"
              className={cn(
                'text-sm font-medium transition-colors',
                location.pathname.startsWith('/dashboard') && !location.pathname.startsWith('/dashboard/account')
                  ? 'text-white'
                  : 'text-white/50 hover:text-white/90',
              )}
            >
              Панель управления
            </Link>
          </nav>
        </div>

        <div className="flex items-center gap-3 md:gap-5">
          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            type="button"
            className="theme-toggle"
            title="Светлая тема"
            aria-label="Переключить тему"
          />

          <a
            href="#"
            className="hidden md:block text-sm font-medium transition-colors text-white/40 hover:text-violet-400"
          >
            Тарифы
          </a>
          <a
            href="#"
            className="hidden md:block text-sm font-medium transition-colors text-white/40 hover:text-violet-400"
          >
            Поддержка
          </a>

          {/* User dropdown */}
          <div
            ref={menuRef}
            className="relative flex items-center gap-2 md:gap-3 pl-3 md:pl-4 border-l border-white/10"
          >
            <button
              type="button"
              onClick={() => setMenuOpen(!menuOpen)}
              className="flex items-center gap-2 rounded-lg px-2 py-1.5 transition-colors hover:bg-white/10"
            >
              <div className="w-8 h-8 rounded-full flex items-center justify-center bg-white/10">
                <User className="w-4 h-4 text-white/60" />
              </div>
              <span className="text-sm font-medium hidden lg:block text-white/70">
                {user?.name || 'Администратор'}
              </span>
            </button>

            {menuOpen && (
              <div className="absolute right-0 top-full mt-2 w-48 rounded-xl border shadow-lg py-1 z-50 bg-neutral-900 border-white/10">
                <Link
                  to="/dashboard/account"
                  onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-2.5 px-4 py-2.5 text-sm transition-colors text-white/70 hover:bg-white/10 hover:text-white"
                >
                  <Settings className="w-4 h-4" />
                  Настройки аккаунта
                </Link>
                <div className="my-1 border-t border-white/10" />
                <button
                  type="button"
                  onClick={handleLogout}
                  className="flex items-center gap-2.5 px-4 py-2.5 text-sm w-full transition-colors text-red-400 hover:bg-red-500/10"
                >
                  <LogOut className="w-4 h-4" />
                  Выйти
                </button>
              </div>
            )}
          </div>
        </div>
      </header>
    )
  }

  return (
    <header className="z-30 transition-all duration-500 px-4 sm:px-8 lg:px-16 pt-4 sm:pt-6 pb-0">
      <div className="flex items-center">
        {/* Left: logo area */}
        <div className="lg:w-[220px] lg:justify-center flex items-center shrink-0">
          {/* Mobile burger */}
          <button
            onClick={onMenuToggle}
            type="button"
            className="lg:hidden h-11 w-11 rounded-lg flex items-center justify-center transition-colors text-neutral-600 hover:text-neutral-900 hover:bg-neutral-100 border border-neutral-900 mr-3"
            aria-label="Открыть меню"
          >
            <Menu className="w-5 h-5" />
          </button>

          <Link to="/dashboard">
            <img
              src="/revisitr/logo.png"
              alt="Revisitr"
              decoding="async"
              className="h-6 sm:h-7 w-auto"
            />
          </Link>
        </div>

        {/* Center: nav links */}
        <nav className="hidden lg:flex items-center gap-6 flex-1 pl-6">
          <Link
            to="/dashboard"
            className={cn(
              'text-sm transition-colors',
              isDashboard
                ? 'font-bold text-neutral-900 no-underline'
                : 'font-medium text-neutral-600 hover:text-neutral-900',
            )}
          >
            Панель управления
          </Link>
          <a
            href="#"
            className="text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
          >
            Тарифы
          </a>
          <a
            href="#"
            className="text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
          >
            Поддержка
          </a>
          <a
            href="#"
            className="text-sm text-[#EF3219] font-medium hover:text-[#FF5C47] transition-colors"
          >
            Маркетинг &laquo;под ключ&raquo;
          </a>
          <a
            href="#"
            className="text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
          >
            Контакты
          </a>
        </nav>

        {/* Right: theme toggle + user */}
        <div className="flex items-center gap-3 md:gap-5 ml-auto">
          {/* Theme toggle */}
          <button
            onClick={toggleTheme}
            type="button"
            className="theme-toggle"
            title="Aurora тема"
            aria-label="Переключить тему"
          />

          {/* User dropdown */}
          <div
            ref={menuRef}
            className="relative flex items-center gap-2"
          >
            <button
              type="button"
              onClick={() => setMenuOpen(!menuOpen)}
              className="flex items-center gap-2.5 rounded px-2.5 py-1.5 transition-all duration-150 hover:scale-[1.03]"
            >
              <div className="w-9 h-9 rounded-full flex items-center justify-center bg-white border border-neutral-900">
                <User className="w-4 h-4 text-neutral-900" />
              </div>
              <div className="hidden sm:flex flex-col items-start">
                <span className="text-sm font-bold text-neutral-900 leading-tight uppercase">
                  {user?.name || 'GENNADY P.'}
                </span>
                <span className="text-[10px] font-bold text-white bg-neutral-900 rounded px-1.5 py-0.5 leading-none mt-0.5 uppercase tracking-wider">
                  pro
                </span>
              </div>
            </button>

            <div className={cn(
              'absolute right-0 top-full mt-1 w-48 rounded border border-neutral-900 py-1 z-50 bg-white',
              'transition-all duration-150 origin-top-right',
              menuOpen
                ? 'opacity-100 scale-y-100 pointer-events-auto'
                : 'opacity-0 scale-y-95 pointer-events-none',
            )}>
                <Link
                  to="/dashboard/account"
                  onClick={() => setMenuOpen(false)}
                  className="flex items-center gap-2.5 px-4 py-2 text-sm transition-colors text-neutral-700 hover:bg-neutral-50"
                >
                  <Settings className="w-4 h-4" />
                  Настройки аккаунта
                </Link>
                <div className="my-1 mx-4 border-t border-neutral-200" />
                <button
                  type="button"
                  onClick={handleLogout}
                  className="flex items-center gap-2.5 px-4 py-2 text-sm w-full transition-colors text-red-500 hover:bg-red-50"
                >
                  <LogOut className="w-4 h-4" />
                  Выйти
                </button>
              </div>
          </div>
        </div>
      </div>

      {/* Bottom divider */}
      <div className="mt-4 border-b border-neutral-900" />
    </header>
  )
}
