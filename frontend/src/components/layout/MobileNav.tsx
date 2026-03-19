import { useState, useEffect, useCallback } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  LayoutDashboard,
  Users,
  Heart,
  Mail,
  Bot,
  Store,
  X,
  LogOut,
  type LucideIcon,
} from 'lucide-react'
import { useAuthStore } from '@/stores/auth'

interface MobileNavItem {
  label: string
  icon: LucideIcon
  href: string
}

const mobileNav: MobileNavItem[] = [
  { label: 'Дашборд', icon: LayoutDashboard, href: '/dashboard' },
  { label: 'Клиенты', icon: Users, href: '/dashboard/clients' },
  { label: 'Лояльность', icon: Heart, href: '/dashboard/loyalty' },
  { label: 'Рассылки', icon: Mail, href: '/dashboard/campaigns' },
  { label: 'Боты', icon: Bot, href: '/dashboard/bots' },
  { label: 'Точки продаж', icon: Store, href: '/dashboard/pos' },
]

interface MobileNavProps {
  isOpen: boolean
  onClose: () => void
}

export function MobileNav({ isOpen, onClose }: MobileNavProps) {
  const [isClosing, setIsClosing] = useState(false)
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)

  const handleClose = useCallback(() => {
    setIsClosing(true)
    setTimeout(() => {
      setIsClosing(false)
      onClose()
    }, 250)
  }, [onClose])

  // Close on route change
  useEffect(() => {
    if (isOpen) handleClose()
  }, [location.pathname]) // eslint-disable-line react-hooks/exhaustive-deps

  // Lock body scroll when open
  useEffect(() => {
    if (isOpen) {
      document.body.style.overflow = 'hidden'
      return () => { document.body.style.overflow = '' }
    }
  }, [isOpen])

  // Close on escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isOpen) handleClose()
    }
    window.addEventListener('keydown', handleKey)
    return () => window.removeEventListener('keydown', handleKey)
  }, [isOpen, handleClose])

  if (!isOpen && !isClosing) return null

  return (
    <div className="fixed inset-0 z-50 lg:hidden" role="dialog" aria-modal="true">
      {/* Backdrop */}
      <div
        className={cn(
          'absolute inset-0 bg-black/40 backdrop-blur-sm',
          isClosing ? 'backdrop-exit' : 'backdrop-enter',
        )}
        onClick={handleClose}
        aria-hidden="true"
      />

      {/* Drawer */}
      <aside
        className={cn(
          'absolute inset-y-0 left-0 w-[280px] sidebar-glass flex flex-col',
          isClosing ? 'drawer-exit' : 'drawer-enter',
        )}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-5">
          <span className="text-xl font-bold text-white tracking-tight select-none group/logo">
            revi<span className="text-accent transition-all duration-300 group-hover/logo:drop-shadow-[0_0_10px_rgba(232,93,58,0.65)]">s</span>itr
          </span>
          <button
            onClick={handleClose}
            type="button"
            className="w-8 h-8 rounded-lg flex items-center justify-center text-sidebar-muted hover:text-white hover:bg-sidebar-hover transition-colors"
            aria-label="Закрыть меню"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-3 py-2 space-y-0.5 overflow-y-auto" aria-label="Мобильная навигация">
          {mobileNav.map((item, i) => {
            const Icon = item.icon
            const isActive = location.pathname === item.href ||
              (item.href !== '/dashboard' && location.pathname.startsWith(item.href))
            return (
              <Link
                key={item.href}
                to={item.href}
                className={cn(
                  'flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all duration-150',
                  isActive
                    ? 'bg-sidebar-active text-white'
                    : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
                )}
                style={{ animationDelay: `${i * 30}ms` }}
              >
                <Icon className="w-5 h-5 shrink-0" />
                <span>{item.label}</span>
                {isActive && (
                  <div className="ml-auto w-1.5 h-1.5 rounded-full bg-accent animate-dot-in" />
                )}
              </Link>
            )
          })}
        </nav>

        {/* Footer */}
        <div className="p-4 border-t border-white/10">
          <button
            onClick={() => {
              logout()
              handleClose()
            }}
            type="button"
            className="w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium text-sidebar-muted hover:text-white hover:bg-sidebar-hover transition-colors"
          >
            <LogOut className="w-5 h-5" />
            Выйти
          </button>
        </div>
      </aside>
    </div>
  )
}
