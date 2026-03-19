import { Link, useNavigate } from 'react-router-dom'
import { User, Menu, LogOut } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'

interface HeaderProps {
  onMenuToggle?: () => void
}

export function Header({ onMenuToggle }: HeaderProps) {
  const navigate = useNavigate()
  const logout = useAuthStore((s) => s.logout)

  const handleLogout = () => {
    logout()
    navigate('/auth/login')
  }

  return (
    <header className="h-16 border-b border-surface-border bg-white/80 backdrop-blur-sm flex items-center justify-between px-4 md:px-6 sticky top-0 z-30">
      <div className="flex items-center gap-4 md:gap-8">
        {/* Mobile burger */}
        <button
          onClick={onMenuToggle}
          type="button"
          className="lg:hidden w-9 h-9 rounded-lg flex items-center justify-center text-neutral-600 hover:text-neutral-900 hover:bg-neutral-100 transition-colors"
          aria-label="Открыть меню"
        >
          <Menu className="w-5 h-5" />
        </button>

        <Link to="/dashboard" className="text-xl font-bold tracking-tight group/logo select-none">
          revi<span className="text-accent transition-all duration-300 group-hover/logo:drop-shadow-[0_0_8px_rgba(232,93,58,0.55)]">s</span>itr
        </Link>

        <nav className="hidden md:flex items-center gap-6">
          <Link
            to="/dashboard"
            className="text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
          >
            Панель управления
          </Link>
        </nav>
      </div>

      <div className="flex items-center gap-3 md:gap-6">
        <a
          href="#"
          className="hidden md:block text-sm font-medium text-neutral-500 hover:text-accent transition-colors"
        >
          Тарифы
        </a>
        <a
          href="#"
          className="hidden md:block text-sm font-medium text-neutral-500 hover:text-accent transition-colors"
        >
          Поддержка
        </a>

        <div className="flex items-center gap-2 md:gap-3 pl-3 md:pl-4 border-l border-surface-border">
          <div className="w-8 h-8 rounded-full bg-neutral-200 flex items-center justify-center">
            <User className="w-4 h-4 text-neutral-500" />
          </div>
          <span className="text-sm font-medium text-neutral-700 hidden lg:block">
            Администратор
          </span>
          <button
            onClick={handleLogout}
            type="button"
            className={cn(
              'w-8 h-8 rounded-lg flex items-center justify-center',
              'text-neutral-400 hover:text-red-500 hover:bg-red-50',
              'transition-colors',
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
