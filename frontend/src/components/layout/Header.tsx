import { Link, useNavigate } from 'react-router-dom'
import { User, Menu, LogOut, Sun, Sparkles } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/contexts/ThemeContext'

interface HeaderProps {
  onMenuToggle?: () => void
}

export function Header({ onMenuToggle }: HeaderProps) {
  const navigate = useNavigate()
  const logout = useAuthStore((s) => s.logout)
  const { theme, toggleTheme } = useTheme()

  const handleLogout = () => {
    logout()
    navigate('/auth/login')
  }

  const isAurora = theme === 'aurora'

  return (
    <header
      className={cn(
        'h-16 border-b flex items-center justify-between px-4 md:px-6 sticky top-0 z-30 transition-all duration-500',
        isAurora ? 'header-glass' : 'bg-white/80 backdrop-blur-sm border-surface-border',
      )}
    >
      <div className="flex items-center gap-4 md:gap-8">
        {/* Mobile burger */}
        <button
          onClick={onMenuToggle}
          type="button"
          className={cn(
            'lg:hidden w-9 h-9 rounded-lg flex items-center justify-center transition-colors',
            isAurora
              ? 'text-white/60 hover:text-white hover:bg-white/10'
              : 'text-neutral-600 hover:text-neutral-900 hover:bg-neutral-100',
          )}
          aria-label="Открыть меню"
        >
          <Menu className="w-5 h-5" />
        </button>

        <Link to="/dashboard" className="text-xl font-bold tracking-tight group/logo select-none">
          <span className={isAurora ? 'text-white/90' : undefined}>
            revi
          </span>
          <span
            className={cn(
              'transition-all duration-300',
              isAurora
                ? 'text-violet-400 group-hover/logo:drop-shadow-[0_0_8px_rgba(139,92,246,0.6)]'
                : 'text-accent group-hover/logo:drop-shadow-[0_0_8px_rgba(232,93,58,0.55)]',
            )}
          >
            s
          </span>
          <span className={isAurora ? 'text-white/90' : undefined}>
            itr
          </span>
        </Link>

        <nav className="hidden md:flex items-center gap-6">
          <Link
            to="/dashboard"
            className={cn(
              'text-sm font-medium transition-colors',
              isAurora
                ? 'text-white/50 hover:text-white/90'
                : 'text-neutral-600 hover:text-neutral-900',
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
          title={isAurora ? 'Светлая тема' : 'Aurora тема'}
          aria-label="Переключить тему"
        />

        <a
          href="#"
          className={cn(
            'hidden md:block text-sm font-medium transition-colors',
            isAurora
              ? 'text-white/40 hover:text-violet-400'
              : 'text-neutral-500 hover:text-accent',
          )}
        >
          Тарифы
        </a>
        <a
          href="#"
          className={cn(
            'hidden md:block text-sm font-medium transition-colors',
            isAurora
              ? 'text-white/40 hover:text-violet-400'
              : 'text-neutral-500 hover:text-accent',
          )}
        >
          Поддержка
        </a>

        <div className={cn(
          'flex items-center gap-2 md:gap-3 pl-3 md:pl-4 border-l',
          isAurora ? 'border-white/10' : 'border-surface-border',
        )}>
          <div className={cn(
            'w-8 h-8 rounded-full flex items-center justify-center',
            isAurora ? 'bg-white/10' : 'bg-neutral-200',
          )}>
            <User className={cn('w-4 h-4', isAurora ? 'text-white/60' : 'text-neutral-500')} />
          </div>
          <span className={cn(
            'text-sm font-medium hidden lg:block',
            isAurora ? 'text-white/70' : 'text-neutral-700',
          )}>
            Администратор
          </span>
          <button
            onClick={handleLogout}
            type="button"
            className={cn(
              'w-8 h-8 rounded-lg flex items-center justify-center transition-colors',
              isAurora
                ? 'text-white/30 hover:text-red-400 hover:bg-red-500/10'
                : 'text-neutral-400 hover:text-red-500 hover:bg-red-50',
            )}
            title="Выйти"
            aria-label="Выйти из аккаунта"
          >
            <LogOut className="w-4 h-4" />
          </button>
        </div>
      </div>
    </header>
  )
}
