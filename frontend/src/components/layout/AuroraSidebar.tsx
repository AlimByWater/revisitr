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
  UtensilsCrossed,
  ShoppingBag,
  Workflow,
  CreditCard,
  ChevronRight,
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
  { label: 'Дашборд', icon: LayoutDashboard, href: '/dashboard' },
  {
    label: 'Аналитика', icon: TrendingUp,
    children: [
      { label: 'Продажи', href: '/dashboard/analytics/sales' },
      { label: 'Лояльность', href: '/dashboard/analytics/loyalty' },
      { label: 'Рассылки', href: '/dashboard/analytics/mailings' },
    ],
  },
  {
    label: 'Клиенты', icon: Users,
    children: [
      { label: 'Клиенты', href: '/dashboard/clients' },
      { label: 'Сегментация', href: '/dashboard/clients/segments' },
      { label: 'Прогнозы', href: '/dashboard/clients/predictions' },
    ],
  },
  {
    label: 'Лояльность', icon: Heart,
    children: [
      { label: 'Мои программы', href: '/dashboard/loyalty' },
      { label: 'Wallet', href: '/dashboard/loyalty/wallet' },
    ],
  },
  {
    label: 'Рассылки', icon: Mail,
    children: [
      { label: 'Все рассылки', href: '/dashboard/campaigns' },
      { label: 'Создать', href: '/dashboard/campaigns/create' },
      { label: 'Шаблоны', href: '/dashboard/campaigns/templates' },
      { label: 'Авто-сценарии', href: '/dashboard/campaigns/scenarios' },
    ],
  },
  {
    label: 'Акции', icon: Tag,
    children: [
      { label: 'Мои акции', href: '/dashboard/promotions' },
      { label: 'Промокоды', href: '/dashboard/promotions/codes' },
      { label: 'Архив', href: '/dashboard/promotions/archive' },
    ],
  },
  {
    label: 'Боты', icon: Bot, badgeKey: 'bots',
    children: [{ label: 'Список ботов', href: '/dashboard/bots' }],
  },
  { label: 'Маркетплейс', icon: ShoppingBag, href: '/dashboard/marketplace' },
  { label: 'Точки продаж', icon: Store, href: '/dashboard/pos' },
  { label: 'Меню', icon: UtensilsCrossed, href: '/dashboard/menus' },
  { label: 'Интеграции', icon: Workflow, href: '/dashboard/integrations' },
  {
    label: 'Биллинг', icon: CreditCard,
    children: [
      { label: 'Тариф', href: '/dashboard/billing' },
      { label: 'Счета', href: '/dashboard/billing/invoices' },
    ],
  },
]

function AuroraNavItem({ item, badges, expanded }: { item: NavItem; badges: Record<string, number>; expanded: boolean }) {
  const location = useLocation()
  const currentPath = location.pathname
  const [open, setOpen] = useState(false)

  const isActive = item.href
    ? currentPath === item.href
    : item.children?.some((c) => currentPath.startsWith(c.href))

  useEffect(() => {
    if (isActive && item.children) setOpen(true)
  }, [isActive, item.children])

  const Icon = item.icon
  const badge = item.badgeKey ? badges[item.badgeKey] : undefined

  if (!item.children) {
    return (
      <Link
        to={item.href!}
        className={cn(
          'group relative flex items-center gap-3 rounded-xl transition-all duration-200',
          expanded ? 'px-3 py-2.5' : 'px-0 py-2.5 justify-center',
          isActive
            ? 'bg-violet-500/15 text-white'
            : 'text-white/35 hover:text-white/80 hover:bg-white/[0.05]',
        )}
        title={!expanded ? item.label : undefined}
      >
        <Icon className={cn('shrink-0', expanded ? 'w-[18px] h-[18px]' : 'w-5 h-5')} />
        {expanded && <span className="text-sm font-medium truncate">{item.label}</span>}
        {isActive && (
          <div className="absolute right-0 top-1/2 -translate-y-1/2 w-[3px] h-4 rounded-l-full bg-violet-400" />
        )}
      </Link>
    )
  }

  return (
    <div>
      <button
        onClick={() => expanded && setOpen(!open)}
        className={cn(
          'group relative w-full flex items-center gap-3 rounded-xl transition-all duration-200',
          expanded ? 'px-3 py-2.5' : 'px-0 py-2.5 justify-center',
          isActive
            ? 'text-white/90'
            : 'text-white/35 hover:text-white/80 hover:bg-white/[0.05]',
        )}
        type="button"
        title={!expanded ? item.label : undefined}
      >
        <Icon className={cn('shrink-0', expanded ? 'w-[18px] h-[18px]' : 'w-5 h-5')} />
        {expanded && (
          <>
            <span className="flex-1 text-left text-sm font-medium truncate">{item.label}</span>
            {badge !== undefined && badge > 0 && (
              <span className="text-[10px] font-bold tabular-nums bg-violet-500/20 text-violet-300 px-1.5 py-0.5 rounded-md">
                {badge}
              </span>
            )}
            <ChevronRight
              className={cn('w-3.5 h-3.5 transition-transform duration-200', open && 'rotate-90')}
            />
          </>
        )}
        {isActive && !expanded && (
          <div className="absolute right-0 top-1/2 -translate-y-1/2 w-[3px] h-4 rounded-l-full bg-violet-400" />
        )}
      </button>

      {expanded && (
        <div
          className={cn(
            'ml-7 space-y-0.5 overflow-hidden transition-all duration-200',
            open ? 'mt-1 max-h-48 opacity-100' : 'max-h-0 opacity-0',
          )}
        >
          {item.children.map((child) => {
            const isChildActive =
              currentPath === child.href ||
              (child.href !== '/dashboard' && currentPath.startsWith(child.href + '/'))
            return (
              <Link
                key={child.href}
                to={child.href}
                className={cn(
                  'block px-3 py-1.5 rounded-lg text-[13px] transition-all duration-150',
                  isChildActive
                    ? 'text-violet-300 bg-violet-500/10'
                    : 'text-white/30 hover:text-white/70 hover:bg-white/[0.04]',
                )}
              >
                {child.label}
              </Link>
            )
          })}
        </div>
      )}
    </div>
  )
}

export function AuroraSidebar() {
  const [expanded, setExpanded] = useState(false)
  const { data: bots } = useBotsQuery()
  const badges: Record<string, number> = { bots: bots?.length ?? 0 }

  return (
    <aside
      onMouseEnter={() => setExpanded(true)}
      onMouseLeave={() => setExpanded(false)}
      className={cn(
        'shrink-0 flex-col h-screen sticky top-0 hidden lg:flex z-40',
        'border-r border-white/[0.06] transition-all duration-300 ease-[cubic-bezier(0.25,0.1,0.25,1)]',
        'bg-[rgba(255,255,255,0.02)] backdrop-blur-xl',
        expanded ? 'w-[260px]' : 'w-[68px]',
      )}
    >
      {/* Logo */}
      <div className={cn('flex items-center h-16 shrink-0', expanded ? 'px-5' : 'justify-center')}>
        {expanded ? (
          <div className="flex items-center gap-2">
            <span className="text-xl font-bold text-white/90 tracking-tight select-none">
              revi<span className="text-violet-400">s</span>itr
            </span>
            <span className="text-[9px] font-bold text-violet-300 bg-violet-500/15 px-1.5 py-0.5 rounded uppercase tracking-wider">
              PRO
            </span>
          </div>
        ) : (
          <div className="w-8 h-8 rounded-lg bg-violet-500/20 flex items-center justify-center">
            <span className="text-sm font-black text-violet-400 select-none">R</span>
          </div>
        )}
      </div>

      {/* Nav */}
      <nav className={cn('flex-1 overflow-y-auto overflow-x-hidden py-2', expanded ? 'px-3' : 'px-2.5')}>
        <div className="space-y-0.5">
          <AuroraNavItem item={navigation[0]} badges={badges} expanded={expanded} />
        </div>

        <div className={cn('mt-4 pt-4 border-t border-white/[0.04] space-y-0.5')}>
          {navigation.slice(1, 6).map((item) => (
            <AuroraNavItem key={item.label} item={item} badges={badges} expanded={expanded} />
          ))}
        </div>

        <div className="mt-4 pt-4 border-t border-white/[0.04] space-y-0.5">
          {navigation.slice(6).map((item) => (
            <AuroraNavItem key={item.label} item={item} badges={badges} expanded={expanded} />
          ))}
        </div>
      </nav>

      {/* Footer */}
      {expanded && (
        <div className="p-4 border-t border-white/[0.04]">
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
            className="block w-full text-[11px] text-white/15 hover:text-white/35 transition-colors text-center mb-2"
          >
            Пройти настройку заново
          </button>
          <p className="text-[10px] font-mono text-white/15 text-center uppercase tracking-wider">
            &copy; 2026 Revisitr
          </p>
        </div>
      )}
    </aside>
  )
}
