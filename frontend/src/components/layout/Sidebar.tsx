import { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useBotsQuery } from '@/features/bots/queries'
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
    label: 'Точки продаж',
    icon: Store,
    href: '/dashboard/pos',
  },
  {
    label: 'Интеграции',
    icon: Workflow,
    href: '/dashboard/integrations',
  },
]

function NavGroup({ item, badges }: { item: NavItem; badges: Record<string, number> }) {
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
          'flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-150',
          isActive
            ? 'bg-sidebar-active text-white'
            : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
        )}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1">{item.label}</span>
        {isActive && <div className="w-1.5 h-1.5 rounded-full bg-accent" />}
      </Link>
    )
  }

  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-all duration-150',
          isActive
            ? 'text-white'
            : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
        )}
        type="button"
        aria-expanded={expanded}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1 text-left">{item.label}</span>
        {badge !== undefined && badge > 0 && (
          <span className="text-[10px] font-bold tabular-nums bg-accent/15 text-accent px-1.5 py-0.5 rounded-md min-w-[20px] text-center">
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
                'block px-4 py-2 rounded-lg text-sm transition-all duration-150',
                isChildActive
                  ? 'text-white bg-sidebar-active'
                  : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
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

  const badges: Record<string, number> = {
    bots: bots?.length ?? 0,
  }

  return (
    <aside className="w-sidebar bg-sidebar shrink-0 flex-col h-screen sticky top-0 hidden lg:flex">
      <div className="p-6">
        <div className="flex items-center gap-2">
          <span className="text-2xl font-bold text-white tracking-tight">
            revi<span className="text-accent">s</span>itr
          </span>
          <span className="text-[10px] font-bold text-accent bg-accent/10 px-1.5 py-0.5 rounded uppercase tracking-wider">
            PRO
          </span>
        </div>
      </div>

      <nav
        className="flex-1 px-3 space-y-1 overflow-y-auto"
        aria-label="Основная навигация"
      >
        {navigation.map((item) => (
          <NavGroup key={item.label} item={item} badges={badges} />
        ))}
      </nav>

      <div className="p-4 border-t border-white/10">
        <p className="text-xs text-sidebar-muted text-center">
          &copy; 2026 Revisitr
        </p>
      </div>
    </aside>
  )
}
