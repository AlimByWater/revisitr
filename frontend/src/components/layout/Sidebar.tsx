import { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
import { useTheme } from '@/contexts/ThemeContext'
import {
  LayoutDashboard,
  TrendingUp,
  Users,
  Heart,
  Mail,
  Tag,
  Bot,
  Store,
  Workflow,
  UtensilsCrossed,
  ShoppingBag,
  CreditCard,
  ChevronDown,
  type LucideIcon,
} from 'lucide-react'

interface SubItem {
  label: string
  href: string
}

interface NavItem {
  label: string
  icon: LucideIcon
  href?: string
  children?: SubItem[]
  badgeKey?: string
}

const navigation: NavItem[] = [
  {
    label: 'Дашборд',
    icon: LayoutDashboard,
    href: '/dashboard',
  },
  {
    label: 'Аналитика',
    icon: TrendingUp,
    children: [
      { label: 'Продажи', href: '/dashboard/analytics/sales' },
      { label: 'Лояльность', href: '/dashboard/analytics/loyalty' },
      { label: 'Рассылки', href: '/dashboard/analytics/mailings' },
    ],
  },
  {
    label: 'Клиенты',
    icon: Users,
    children: [
      { label: 'Клиенты', href: '/dashboard/clients' },
      { label: 'Сегментация', href: '/dashboard/clients/segments' },
      { label: 'Прогнозы', href: '/dashboard/clients/predictions' },
    ],
  },
  {
    label: 'Лояльность',
    icon: Heart,
    children: [
      { label: 'Мои программы', href: '/dashboard/loyalty' },
      { label: 'Wallet', href: '/dashboard/loyalty/wallet' },
    ],
  },
  {
    label: 'Рассылки',
    icon: Mail,
    children: [
      { label: 'Все рассылки', href: '/dashboard/campaigns' },
      { label: 'Создать', href: '/dashboard/campaigns/create' },
      { label: 'Шаблоны', href: '/dashboard/campaigns/templates' },
      { label: 'Авто-сценарии', href: '/dashboard/campaigns/scenarios' },
    ],
  },
  {
    label: 'Акции',
    icon: Tag,
    children: [
      { label: 'Мои акции', href: '/dashboard/promotions' },
      { label: 'Промокоды', href: '/dashboard/promotions/codes' },
      { label: 'Архив', href: '/dashboard/promotions/archive' },
    ],
  },
  {
    label: 'Мои боты',
    icon: Bot,
    badgeKey: 'bots',
    children: [
      { label: 'Список ботов', href: '/dashboard/bots' },
    ],
  },
  {
    label: 'Маркетплейс',
    icon: ShoppingBag,
    href: '/dashboard/marketplace',
  },
  {
    label: 'Точки продаж',
    icon: Store,
    href: '/dashboard/pos',
  },
  {
    label: 'Меню',
    icon: UtensilsCrossed,
    href: '/dashboard/menus',
  },
  {
    label: 'Интеграции',
    icon: Workflow,
    href: '/dashboard/integrations',
  },
  {
    label: 'Биллинг',
    icon: CreditCard,
    children: [
      { label: 'Тариф', href: '/dashboard/billing' },
      { label: 'Счета', href: '/dashboard/billing/invoices' },
    ],
  },
]

function NavGroup({ item, badges, isAurora }: { item: NavItem; badges: Record<string, number>; isAurora: boolean }) {
  const location = useLocation()
  const currentPath = location.pathname

  const isActive = item.href
    ? currentPath === item.href
    : item.children?.some((child) => currentPath.startsWith(child.href))

  // Auto-expand active groups
  const [expanded, setExpanded] = useState(isActive ?? false)

  useEffect(() => {
    if (isActive && item.children) setExpanded(true)
  }, [isActive, item.children])

  const Icon = item.icon
  const badge = item.badgeKey ? badges[item.badgeKey] : undefined

  if (!item.children) {
    return (
      <Link
        to={item.href!}
        className={cn(
          'flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200',
          isActive
            ? isAurora
              ? 'bg-[var(--color-accent)]/15 text-white'
              : 'bg-sidebar-active text-white'
            : isAurora
              ? 'text-white/40 hover:text-white/90 hover:bg-white/[0.06] hover:translate-x-0.5'
              : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover hover:translate-x-0.5',
        )}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1">{item.label}</span>
        {isActive && (
          <div
            className={cn(
              'w-1.5 h-1.5 rounded-full animate-dot-in',
              isAurora ? 'bg-violet-400' : 'bg-accent',
            )}
          />
        )}
      </Link>
    )
  }

  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-200',
          isActive
            ? isAurora ? 'text-white/90' : 'text-white'
            : isAurora
              ? 'text-white/40 hover:text-white/90 hover:bg-white/[0.06] hover:translate-x-0.5'
              : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover hover:translate-x-0.5',
        )}
        type="button"
        aria-expanded={expanded}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1 text-left">{item.label}</span>
        {badge !== undefined && badge > 0 && (
          <span className={cn(
            'text-[10px] font-bold tabular-nums px-1.5 py-0.5 rounded-md min-w-[20px] text-center',
            isAurora
              ? 'bg-violet-500/20 text-violet-300'
              : 'bg-accent/15 text-accent',
          )}>
            {badge}
          </span>
        )}
        <ChevronDown
          className={cn(
            'w-4 h-4 transition-transform duration-200',
            expanded && 'rotate-180',
          )}
        />
      </button>

      <div
        className={cn(
          'ml-8 space-y-0.5 overflow-hidden transition-all duration-200',
          expanded ? 'mt-1 max-h-40 opacity-100' : 'max-h-0 opacity-0',
        )}
      >
        {item.children.map((child) => {
          const isChildActive = currentPath === child.href ||
            (child.href !== '/dashboard' && currentPath.startsWith(child.href + '/'))
          return (
            <Link
              key={child.href}
              to={child.href}
              className={cn(
                'block px-4 py-2 rounded-lg text-sm transition-all duration-200',
                isChildActive
                  ? isAurora
                    ? 'text-white bg-[var(--color-accent)]/15'
                    : 'text-white bg-sidebar-active'
                  : isAurora
                    ? 'text-white/35 hover:text-white/90 hover:bg-white/[0.06] hover:translate-x-0.5'
                    : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover hover:translate-x-0.5',
              )}
            >
              {child.label}
            </Link>
          )
        })}
      </div>
    </div>
  )
}

export function Sidebar() {
  const { data: bots } = useBotsQuery()
  const { theme } = useTheme()
  const isAurora = theme === 'aurora'

  const badges: Record<string, number> = {
    bots: bots?.length ?? 0,
  }

  return (
    <aside className="w-sidebar sidebar-glass shrink-0 flex-col h-screen sticky top-0 hidden lg:flex">
      <div className="p-6">
        <div className="flex items-center gap-2 group/logo">
          <span className={cn(
            'text-2xl font-bold tracking-tight select-none',
            isAurora ? 'text-white/90' : 'text-white',
          )}>
            revi
            <span
              className={cn(
                'transition-all duration-300',
                isAurora
                  ? 'text-violet-400 group-hover/logo:drop-shadow-[0_0_10px_rgba(139,92,246,0.65)]'
                  : 'text-accent group-hover/logo:drop-shadow-[0_0_10px_rgba(232,93,58,0.65)]',
              )}
            >
              s
            </span>
            itr
          </span>
          <span className={cn(
            'text-[10px] font-bold px-1.5 py-0.5 rounded uppercase tracking-wider',
            isAurora
              ? 'text-violet-300 bg-violet-500/15'
              : 'text-accent bg-accent/10',
          )}>
            PRO
          </span>
        </div>
      </div>

      <nav
        className="flex-1 px-3 overflow-y-auto py-1"
        aria-label="Основная навигация"
      >
        {/* Primary */}
        <NavGroup item={navigation[0]} badges={badges} isAurora={isAurora} />

        {/* Business data */}
        <div className="mt-4 space-y-0.5">
          {navigation.slice(1, 6).map((item) => (
            <NavGroup key={item.label} item={item} badges={badges} isAurora={isAurora} />
          ))}
        </div>

        {/* Configuration */}
        <div className={cn(
          'mt-4 pt-4 border-t space-y-0.5',
          isAurora ? 'border-white/[0.05]' : 'border-white/[0.07]',
        )}>
          {navigation.slice(6).map((item) => (
            <NavGroup key={item.label} item={item} badges={badges} isAurora={isAurora} />
          ))}
        </div>
      </nav>

      <div className={cn(
        'p-4 border-t',
        isAurora ? 'border-white/[0.05]' : 'border-white/10',
      )}>
        <button
          type="button"
          onClick={async () => {
            const token = localStorage.getItem('token')
            if (!token) return
            try {
              const baseURL = import.meta.env.VITE_API_URL || '/api/v1'
              await fetch(`${baseURL}/onboarding/reset`, {
                method: 'POST',
                headers: { Authorization: `Bearer ${token}` },
              })
              const basePath = import.meta.env.BASE_URL?.replace(/\/$/, '') || ''
              window.location.href = `${basePath}/dashboard/onboarding`
            } catch { /* ignore */ }
          }}
          className="block w-full text-[11px] text-white/20 hover:text-white/40 transition-colors text-center mb-2"
        >
          Пройти настройку заново
        </button>
        <p className="text-[11px] font-mono text-white/20 text-center uppercase tracking-wider">
          &copy; 2026 Revisitr
        </p>
      </div>
    </aside>
  )
}
