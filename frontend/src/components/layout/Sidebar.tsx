import { useState } from 'react'
import { Link, useRouterState } from '@tanstack/react-router'
import { cn } from '@/lib/utils'
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
      { label: 'Мои программы', href: '/dashboard/loyalty/programs' },
      { label: 'Wallet', href: '/dashboard/loyalty/wallet' },
    ],
  },
  {
    label: 'Рассылки',
    icon: Mail,
    children: [
      { label: 'Запуск', href: '/dashboard/mailings/launch' },
      { label: 'Авто-рассылки', href: '/dashboard/mailings/auto' },
      { label: 'Архив', href: '/dashboard/mailings/archive' },
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
    children: [
      { label: 'Список ботов', href: '/dashboard/bots' },
    ],
  },
  {
    label: 'Точки продаж',
    icon: Store,
    href: '/dashboard/locations',
  },
  {
    label: 'Интеграции',
    icon: Workflow,
    href: '/dashboard/integrations',
  },
]

function NavGroup({ item }: { item: NavItem }) {
  const [expanded, setExpanded] = useState(false)
  const routerState = useRouterState()
  const currentPath = routerState.location.pathname

  const isActive = item.href
    ? currentPath === item.href
    : item.children?.some((child) => currentPath.startsWith(child.href))

  const Icon = item.icon

  if (!item.children) {
    return (
      <Link
        to={item.href!}
        className={cn(
          'flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors',
          isActive
            ? 'bg-sidebar-active text-white'
            : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
        )}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span>{item.label}</span>
      </Link>
    )
  }

  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm font-medium transition-colors',
          isActive
            ? 'text-white'
            : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
        )}
        type="button"
        aria-expanded={expanded}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1 text-left">{item.label}</span>
        <ChevronDown
          className={cn(
            'w-4 h-4 transition-transform duration-200',
            expanded && 'rotate-180',
          )}
        />
      </button>

      {expanded && (
        <div className="ml-8 mt-1 space-y-0.5">
          {item.children.map((child) => (
            <Link
              key={child.href}
              to={child.href}
              className={cn(
                'block px-4 py-2 rounded-lg text-sm transition-colors',
                currentPath === child.href
                  ? 'text-white bg-sidebar-active'
                  : 'text-sidebar-muted hover:text-white hover:bg-sidebar-hover',
              )}
            >
              {child.label}
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}

export function Sidebar() {
  return (
    <aside className="w-sidebar bg-sidebar shrink-0 flex flex-col h-screen sticky top-0">
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
          <NavGroup key={item.label} item={item} />
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
