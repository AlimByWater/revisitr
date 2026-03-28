import { useState, useEffect } from 'react'
import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
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

type NavEntry = NavItem | 'separator'

const navigation: NavEntry[] = [
  {
    label: 'Дашборд',
    icon: LayoutDashboard,
    href: '/dashboard',
  },
  'separator',
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
  'separator',
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
    href: '/dashboard/pos',
  },
  {
    label: 'Интеграции',
    icon: Workflow,
    href: '/dashboard/integrations',
  },
]

function NavGroup({ item }: { item: NavItem }) {
  const location = useLocation()
  const currentPath = location.pathname

  const isActive = item.href
    ? currentPath === item.href
    : item.children?.some((child) => currentPath.startsWith(child.href))

  const [expanded, setExpanded] = useState(false)

  useEffect(() => {
    if (isActive && item.children) setExpanded(true)
  }, [isActive, item.children])

  const Icon = item.icon

  if (!item.children) {
    return (
      <Link
        to={item.href!}
        className={cn(
          'flex items-center gap-3 px-3 py-2 text-sm transition-colors',
          isActive
            ? 'font-bold text-neutral-900'
            : 'text-neutral-600 hover:text-neutral-900',
        )}
      >
        <Icon className="w-[18px] h-[18px] shrink-0" />
        <span>{item.label}</span>
      </Link>
    )
  }

  return (
    <div>
      <button
        onClick={() => setExpanded(!expanded)}
        className={cn(
          'w-full flex items-center gap-3 px-3 py-2 text-sm transition-colors',
          isActive
            ? 'font-bold text-neutral-900'
            : 'text-neutral-600 hover:text-neutral-900',
        )}
        type="button"
        aria-expanded={expanded}
      >
        <Icon className="w-[18px] h-[18px] shrink-0" />
        <span className="flex-1 text-left">{item.label}</span>
        <ChevronDown
          className={cn(
            'w-4 h-4 text-neutral-400 transition-transform duration-200',
            expanded && 'rotate-180',
          )}
        />
      </button>

      <div
        className={cn(
          'ml-[30px] overflow-hidden transition-all duration-200',
          expanded ? 'max-h-60 opacity-100 mb-1' : 'max-h-0 opacity-0',
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
                'block px-3 py-1.5 text-sm transition-colors',
                isChildActive
                  ? 'font-medium text-neutral-900'
                  : 'text-neutral-400 hover:text-neutral-700 transition-colors',
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
  const { theme } = useTheme()

  if (theme === 'aurora') return null

  return (
    <aside className="hidden lg:block w-[220px] shrink-0">
      <nav
        className="border border-neutral-900 rounded px-2 py-3 bg-white"
        aria-label="Основная навигация"
      >
        {navigation.map((entry, i) => {
          if (entry === 'separator') {
            return <div key={`sep-${i}`} className="my-2 ml-3 mr-2 border-t border-neutral-200" />
          }
          return <NavGroup key={entry.label} item={entry} />
        })}
      </nav>
    </aside>
  )
}
