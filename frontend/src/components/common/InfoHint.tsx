import type { ReactNode } from 'react'
import { Info } from 'lucide-react'
import { cn } from '@/lib/utils'

interface InfoHintProps {
  content: ReactNode
  label?: string
  className?: string
}

export function InfoHint({
  content,
  label = 'Показать подробности',
  className,
}: InfoHintProps) {
  return (
    <details className={cn('group relative inline-flex', className)}>
      <summary
        aria-label={label}
        className={cn(
          'flex h-5 w-5 cursor-pointer list-none items-center justify-center rounded-full',
          'text-neutral-400 transition-colors hover:bg-neutral-100 hover:text-neutral-700',
          'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:text-neutral-700',
        )}
      >
        <Info className="h-4 w-4" />
      </summary>
      <div
        className={cn(
          'absolute left-0 top-full z-20 mt-2 w-64 rounded-xl border border-surface-border bg-white p-3 shadow-lg',
          'text-sm leading-relaxed text-neutral-600',
        )}
      >
        {content}
      </div>
    </details>
  )
}
