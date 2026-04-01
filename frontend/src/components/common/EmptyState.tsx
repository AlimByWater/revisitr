import { cn } from '@/lib/utils'
import { Plus, type LucideIcon } from 'lucide-react'

interface EmptyStateProps {
  icon: LucideIcon
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
  variant?: string
  className?: string
}

export function EmptyState({
  icon: Icon,
  title,
  description,
  actionLabel,
  onAction,
  className,
}: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center justify-center py-24 text-center',
        className,
      )}
    >
      <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
        <Icon className="w-8 h-8 text-neutral-400" />
      </div>

      <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5 tracking-tight">
        {title}
      </h3>
      <p className="text-sm text-neutral-400 max-w-xs leading-relaxed mb-4">
        {description}
      </p>

      {actionLabel && onAction && (
        <button
          onClick={onAction}
          type="button"
          className={cn(
            'inline-flex items-center gap-2 py-2.5 px-5 rounded',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          {actionLabel}
        </button>
      )}
    </div>
  )
}
