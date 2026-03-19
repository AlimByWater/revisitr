import { Link, useNavigate, useLocation } from 'react-router-dom'
import { User, LogOut, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/contexts/ThemeContext'

const breadcrumbMap: Record<string, string> = {
  dashboard: 'Дашборд',
  analytics: 'Аналитика',
  sales: 'Продажи',
  loyalty: 'Лояльность',
  mailings: 'Рассылки',
  clients: 'Клиенты',
  segments: 'Сегментация',
  bots: 'Боты',
  pos: 'Точки продаж',
  campaigns: 'Рассылки',
  create: 'Создать',
  scenarios: 'Сценарии',
  promotions: 'Акции',
  codes: 'Промокоды',
  archive: 'Архив',
  integrations: 'Интеграции',
}

export function AuroraHeader() {
  const navigate = useNavigate()
  const location = useLocation()
  const logout = useAuthStore((s) => s.logout)
  const { toggleTheme } = useTheme()

  const handleLogout = () => {
    logout()
    navigate('/auth/login')
  }

  // Build breadcrumbs from path
  const parts = location.pathname.replace('/dashboard', '').split('/').filter(Boolean)
  const crumbs = parts.map((part, i) => ({
    label: breadcrumbMap[part] || part,
    path: '/dashboard/' + parts.slice(0, i + 1).join('/'),
    isLast: i === parts.length - 1,
  }))

  return (
    <header className="h-14 flex items-center justify-between px-6 sticky top-0 z-30 border-b border-white/[0.04] bg-[rgba(15,11,26,0.5)] backdrop-blur-xl">
      <div className="flex items-center gap-2 text-sm">
        <Link to="/dashboard" className="text-white/30 hover:text-white/60 transition-colors font-medium">
          Дашборд
        </Link>
        {crumbs.map((c) => (
          <span key={c.path} className="flex items-center gap-2">
            <ChevronRight className="w-3 h-3 text-white/15" />
            {c.isLast ? (
              <span className="text-white/80 font-medium">{c.label}</span>
            ) : (
              <Link to={c.path} className="text-white/30 hover:text-white/60 transition-colors font-medium">
                {c.label}
              </Link>
            )}
          </span>
        ))}
      </div>

      <div className="flex items-center gap-4">
        <button
          onClick={toggleTheme}
          type="button"
          className="theme-toggle"
          title="Светлая тема"
          aria-label="Переключить тему"
        />

        <div className="flex items-center gap-2 pl-4 border-l border-white/[0.06]">
          <div className="w-7 h-7 rounded-full bg-gradient-to-br from-violet-500/30 to-blue-500/30 flex items-center justify-center ring-1 ring-white/10">
            <User className="w-3.5 h-3.5 text-white/60" />
          </div>
          <span className="text-[13px] font-medium text-white/50 hidden lg:block">
            Админ
          </span>
          <button
            onClick={handleLogout}
            type="button"
            className="w-7 h-7 rounded-lg flex items-center justify-center text-white/20 hover:text-red-400 hover:bg-red-500/10 transition-colors"
            title="Выйти"
          >
            <LogOut className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>
    </header>
  )
}
