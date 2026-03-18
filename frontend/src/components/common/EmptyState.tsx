import { cn } from '@/lib/utils'
import { Plus, type LucideIcon } from 'lucide-react'

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
  variant?: 'default' | 'bots' | 'loyalty' | 'clients' | 'campaigns' | 'pos'
  className?: string
}

const patterns: Record<string, React.ReactNode> = {
  bots: (
    <svg width="160" height="120" viewBox="0 0 160 120" fill="none" className="mx-auto">
      {/* Bot face */}
      <rect x="50" y="20" width="60" height="50" rx="12" stroke="#d4d4d4" strokeWidth="1.5" />
      <circle cx="70" cy="45" r="5" fill="#E85D3A" opacity="0.2" />
      <circle cx="90" cy="45" r="5" fill="#E85D3A" opacity="0.2" />
      <circle cx="70" cy="45" r="2" fill="#E85D3A" />
      <circle cx="90" cy="45" r="2" fill="#E85D3A" />
      <path d="M73 55C73 55 80 60 87 55" stroke="#d4d4d4" strokeWidth="1.5" strokeLinecap="round" />
      {/* Antenna */}
      <line x1="80" y1="20" x2="80" y2="10" stroke="#d4d4d4" strokeWidth="1.5" />
      <circle cx="80" cy="8" r="3" fill="#E85D3A" opacity="0.3" />
      {/* Body */}
      <rect x="55" y="75" width="50" height="25" rx="6" stroke="#d4d4d4" strokeWidth="1.5" />
      <line x1="50" y1="70" x2="55" y2="85" stroke="#d4d4d4" strokeWidth="1.5" />
      <line x1="110" y1="70" x2="105" y2="85" stroke="#d4d4d4" strokeWidth="1.5" />
      {/* Signal waves */}
      <path d="M120 30C125 25 125 15 120 10" stroke="#E85D3A" strokeWidth="1" opacity="0.3" strokeLinecap="round" />
      <path d="M127 33C135 25 135 12 127 5" stroke="#E85D3A" strokeWidth="1" opacity="0.15" strokeLinecap="round" />
    </svg>
  ),
  loyalty: (
    <svg width="160" height="120" viewBox="0 0 160 120" fill="none" className="mx-auto">
      {/* Heart */}
      <path
        d="M80 95L30 50C20 35 25 15 45 15C55 15 65 25 80 40C95 25 105 15 115 15C135 15 140 35 130 50L80 95Z"
        stroke="#d4d4d4" strokeWidth="1.5" fill="none"
      />
      {/* Inner heart glow */}
      <path
        d="M80 80L48 52C42 43 45 30 55 30C62 30 70 37 80 48C90 37 98 30 105 30C115 30 118 43 112 52L80 80Z"
        fill="#E85D3A" opacity="0.08"
      />
      {/* Star badges */}
      <circle cx="35" cy="80" r="8" stroke="#d4d4d4" strokeWidth="1" />
      <text x="35" y="84" textAnchor="middle" fill="#d4d4d4" fontSize="10" fontWeight="600">★</text>
      <circle cx="80" cy="105" r="8" stroke="#E85D3A" strokeWidth="1" opacity="0.3" />
      <text x="80" y="109" textAnchor="middle" fill="#E85D3A" fontSize="10" opacity="0.5" fontWeight="600">★</text>
      <circle cx="125" cy="80" r="8" stroke="#d4d4d4" strokeWidth="1" />
      <text x="125" y="84" textAnchor="middle" fill="#d4d4d4" fontSize="10" fontWeight="600">★</text>
    </svg>
  ),
  clients: (
    <svg width="160" height="120" viewBox="0 0 160 120" fill="none" className="mx-auto">
      {/* Person silhouettes staggered */}
      <circle cx="55" cy="35" r="12" stroke="#d4d4d4" strokeWidth="1.5" />
      <path d="M35 70C35 55 45 48 55 48C65 48 75 55 75 70" stroke="#d4d4d4" strokeWidth="1.5" fill="none" />
      <circle cx="80" cy="28" r="14" stroke="#d4d4d4" strokeWidth="1.5" />
      <path d="M56 68C56 50 66 42 80 42C94 42 104 50 104 68" stroke="#d4d4d4" strokeWidth="1.5" fill="none" />
      <circle cx="105" cy="35" r="12" stroke="#d4d4d4" strokeWidth="1.5" />
      <path d="M85 70C85 55 95 48 105 48C115 48 125 55 125 70" stroke="#d4d4d4" strokeWidth="1.5" fill="none" />
      {/* Accent dot */}
      <circle cx="80" cy="85" r="4" fill="#E85D3A" opacity="0.2" />
      <circle cx="80" cy="85" r="1.5" fill="#E85D3A" />
      {/* Dotted connection line */}
      <line x1="40" y1="85" x2="120" y2="85" stroke="#d4d4d4" strokeWidth="1" strokeDasharray="3 3" />
    </svg>
  ),
  campaigns: (
    <svg width="160" height="120" viewBox="0 0 160 120" fill="none" className="mx-auto">
      {/* Envelope */}
      <rect x="30" y="30" width="100" height="65" rx="6" stroke="#d4d4d4" strokeWidth="1.5" />
      <path d="M30 36L80 68L130 36" stroke="#d4d4d4" strokeWidth="1.5" />
      {/* Paper peeking out */}
      <rect x="45" y="15" width="70" height="40" rx="3" fill="white" stroke="#d4d4d4" strokeWidth="1" />
      <line x1="55" y1="25" x2="95" y2="25" stroke="#e5e5e5" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="55" y1="32" x2="105" y2="32" stroke="#e5e5e5" strokeWidth="1.5" strokeLinecap="round" />
      <line x1="55" y1="39" x2="80" y2="39" stroke="#e5e5e5" strokeWidth="1.5" strokeLinecap="round" />
      {/* Send arrow */}
      <circle cx="125" cy="25" r="12" fill="#E85D3A" opacity="0.1" />
      <path d="M120 25L128 25M128 25L124 21M128 25L124 29" stroke="#E85D3A" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
  pos: (
    <svg width="160" height="120" viewBox="0 0 160 120" fill="none" className="mx-auto">
      {/* Building */}
      <rect x="45" y="25" width="70" height="70" rx="4" stroke="#d4d4d4" strokeWidth="1.5" />
      {/* Door */}
      <rect x="70" y="70" width="20" height="25" rx="2" stroke="#d4d4d4" strokeWidth="1.5" />
      <circle cx="86" cy="83" r="1.5" fill="#d4d4d4" />
      {/* Windows */}
      <rect x="55" y="35" width="14" height="12" rx="1.5" fill="#E85D3A" opacity="0.08" stroke="#d4d4d4" strokeWidth="1" />
      <rect x="91" y="35" width="14" height="12" rx="1.5" fill="#E85D3A" opacity="0.08" stroke="#d4d4d4" strokeWidth="1" />
      <rect x="55" y="55" width="14" height="12" rx="1.5" fill="#E85D3A" opacity="0.05" stroke="#d4d4d4" strokeWidth="1" />
      <rect x="91" y="55" width="14" height="12" rx="1.5" fill="#E85D3A" opacity="0.05" stroke="#d4d4d4" strokeWidth="1" />
      {/* Roof accent */}
      <path d="M40 25L80 8L120 25" stroke="#d4d4d4" strokeWidth="1.5" />
      {/* Pin marker */}
      <circle cx="130" cy="18" r="8" fill="#E85D3A" opacity="0.15" />
      <circle cx="130" cy="18" r="3" fill="#E85D3A" opacity="0.4" />
    </svg>
  ),
  default: (
    <svg width="160" height="120" viewBox="0 0 160 120" fill="none" className="mx-auto">
      <rect x="40" y="20" width="80" height="80" rx="16" stroke="#d4d4d4" strokeWidth="1.5" />
      <circle cx="80" cy="55" r="12" stroke="#d4d4d4" strokeWidth="1.5" />
      <path d="M75 55L78 58L85 51" stroke="#E85D3A" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  ),
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  actionLabel,
  onAction,
  variant = 'default',
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        'bg-white rounded-2xl border border-surface-border',
        'py-16 px-8 text-center',
        'animate-in',
        className,
      )}
    >
      <div className="mb-6 opacity-80">
        {patterns[variant] || patterns.default}
      </div>

      <div className="w-12 h-12 rounded-2xl bg-neutral-50 border border-surface-border flex items-center justify-center mx-auto mb-5">
        <Icon className="w-6 h-6 text-neutral-400" />
      </div>

      <h2 className="font-serif text-xl font-bold text-neutral-800 mb-2 tracking-tight">
        {title}
      </h2>
      <p className="text-sm text-neutral-400 max-w-sm mx-auto leading-relaxed mb-8">
        {description}
      </p>

      {actionLabel && onAction && (
        <button
          onClick={onAction}
          type="button"
          className={cn(
            'inline-flex items-center gap-2 py-2.5 px-5 rounded-lg',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
            'shadow-sm shadow-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          {actionLabel}
        </button>
      )}
    </div>
  )
}
