import { Link } from '@tanstack/react-router'
import { User } from 'lucide-react'

export function Header() {
  return (
    <header className="h-16 border-b border-surface-border bg-white/80 backdrop-blur-sm flex items-center justify-between px-6 sticky top-0 z-30">
      <div className="flex items-center gap-8">
        <Link to="/dashboard" className="text-xl font-bold tracking-tight">
          revi<span className="text-accent">s</span>itr
        </Link>

        <nav className="hidden md:flex items-center gap-6">
          <Link
            to="/dashboard"
            className="text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
            activeProps={{ className: 'text-neutral-900' }}
          >
            Панель управления
          </Link>
        </nav>
      </div>

      <div className="flex items-center gap-6">
        <a
          href="#"
          className="text-sm font-medium text-neutral-500 hover:text-accent transition-colors"
        >
          Тарифы
        </a>
        <a
          href="#"
          className="text-sm font-medium text-neutral-500 hover:text-accent transition-colors"
        >
          Поддержка
        </a>

        <div className="flex items-center gap-3 pl-4 border-l border-surface-border">
          <div className="w-8 h-8 rounded-full bg-neutral-200 flex items-center justify-center">
            <User className="w-4 h-4 text-neutral-500" />
          </div>
          <span className="text-sm font-medium text-neutral-700 hidden lg:block">
            Администратор
          </span>
        </div>
      </div>
    </header>
  )
}
