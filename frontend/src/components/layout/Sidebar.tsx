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
  CreditCard,
  ChevronDown,
  Smile,
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
      { label: 'RFM', href: '/dashboard/clients/segments' },
    ],
  },
  {
    label: 'Клиенты',
    icon: Users,
    children: [
      { label: 'Клиенты', href: '/dashboard/clients' },
      { label: 'RFM-сегменты', href: '/dashboard/rfm' },
      { label: 'Мои сегменты', href: '/dashboard/clients/custom-segments' },
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
      { label: 'Создать рассылку', href: '/dashboard/campaigns/create' },
      { label: 'Все рассылки', href: '/dashboard/campaigns' },
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
  {
    label: 'Биллинг',
    icon: CreditCard,
    children: [
      { label: 'Тариф', href: '/dashboard/billing' },
      { label: 'Счета', href: '/dashboard/billing/invoices' },
    ],
  },
  {
    label: 'Эмодзи',
    icon: Smile,
    href: '/dashboard/emoji-packs',
  },
]

// Group indices for default theme ordering:
// Group 1 (top): Дашборд (index 0)
// Group 2 (business): Аналитика(1), Клиенты(2), Лояльность(3), Рассылки(4), Акции(5)
// Group 3 (config): Мои боты(6), Точки продаж(7), Интеграции(8), Биллинг(9), Эмодзи(10)
const groupBusiness = [1, 2, 3, 4, 5]
const groupConfig = [6, 7, 8, 9, 10]

// Collect ALL child hrefs across all nav groups for global best-match logic
const allNavHrefs = navigation.flatMap(item =>
  item.children?.map(c => c.href) ?? (item.href ? [item.href] : [])
)

/** Check if `href` is the best (longest) match for `path` among all nav hrefs */
function isBestNavMatch(href: string, path: string): boolean {
  if (path !== href && !path.startsWith(href + '/')) return false
  // Ensure no other href is a longer/better match
  return !allNavHrefs.some(other =>
    other !== href &&
    other.length > href.length &&
    (path === other || path.startsWith(other + '/'))
  )
}

function NavGroup({ item, badges }: { item: NavItem; badges: Record<string, number> }) {
  const location = useLocation()
  const currentPath = location.pathname

  const isActive = item.href
    ? isBestNavMatch(item.href, currentPath)
    : item.children?.some((child) => isBestNavMatch(child.href, currentPath))

  // Collapsed by default, auto-expand only when child is active
  const [expanded, setExpanded] = useState(false)

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
          'flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm transition-all duration-200',
          isActive
            ? 'font-bold text-neutral-900'
            : 'text-neutral-600 hover:text-neutral-900 hover:scale-[1.02] transition-all duration-150',
        )}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1">{item.label}</span>
      </Link>
    )
  }

  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-3 px-4 py-2.5 rounded-lg text-sm transition-all duration-200',
          isActive
            ? 'font-bold text-neutral-900'
            : 'text-neutral-600 hover:text-neutral-900 hover:scale-[1.02] transition-all duration-150',
        )}
        type="button"
        aria-expanded={expanded}
      >
        <Icon className="w-5 h-5 shrink-0" />
        <span className="flex-1 text-left">{item.label}</span>
        {badge !== undefined && badge > 0 && (
          <span className="text-[10px] font-medium tabular-nums px-1.5 py-0.5 rounded-sm min-w-[18px] text-center bg-neutral-100 text-neutral-600 border border-neutral-300">
            {badge}
          </span>
        )}
        <ChevronDown
          className={cn(
            'w-4 h-4 transition-transform duration-200 text-neutral-500',
            expanded && 'rotate-180',
          )}
        />
      </button>

      <div
        className={cn(
          'ml-8 space-y-0.5 overflow-hidden transition-all duration-200',
          expanded ? 'max-h-40 opacity-100 mb-1' : 'max-h-0 opacity-0',
        )}
      >
        {item.children.map((child) => {
          const isChildActive = isBestNavMatch(child.href, currentPath)
          return (
            <Link
              key={child.href}
              to={child.href}
              className={cn(
                'block px-4 py-2 rounded-lg text-sm transition-all duration-200',
                isChildActive
                  ? 'font-medium text-neutral-900'
                  : 'text-neutral-400 hover:text-neutral-700',
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
    <aside className="hidden lg:block w-[220px] shrink-0">
      <nav
        className="border border-neutral-900 rounded px-2 py-3 bg-white"
        aria-label="Основная навигация"
      >
        {/* Дашборд */}
        <NavGroup item={navigation[0]} badges={badges} />

        {/* Separator */}
        <div className="my-2 ml-3 mr-2 border-t border-neutral-200" />

        {/* Аналитика, Клиенты, Лояльность, Рассылки, Акции */}
        <div className="space-y-0.5">
          {groupBusiness.map((i) => (
            <NavGroup key={navigation[i].label} item={navigation[i]} badges={badges} />
          ))}
        </div>

        {/* Separator */}
        <div className="my-2 ml-3 mr-2 border-t border-neutral-200" />

        {/* Мои боты, Маркетплейс, Точки продаж, Меню, Интеграции, Биллинг */}
        <div className="space-y-0.5">
          {groupConfig.map((i) => (
            <NavGroup key={navigation[i].label} item={navigation[i]} badges={badges} />
          ))}
        </div>
      </nav>
    </aside>
  )
}
