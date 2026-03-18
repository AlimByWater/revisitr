import { cn } from '@/lib/utils'
import { RefreshCw } from 'lucide-react'

interface ErrorStateProps {
  title?: string
  message?: string
  onRetry?: () => void
  className?: string
}

export function ErrorState({
  title = 'Что-то пошло не так',
  message = 'Не удалось загрузить данные. Проверьте подключение и попробуйте снова.',
  onRetry,
  className,
}: ErrorStateProps) {
  return (
    <div className={cn('flex flex-col items-center justify-center py-16 px-6', className)}>
      {/* Glitch-style error icon */}
      <div className="relative mb-6">
        <svg width="64" height="64" viewBox="0 0 64 64" fill="none" className="text-neutral-300">
          <rect x="8" y="12" width="48" height="40" rx="4" stroke="currentColor" strokeWidth="2" />
          <path d="M8 22H56" stroke="currentColor" strokeWidth="2" />
          <circle cx="16" cy="17" r="1.5" fill="currentColor" />
          <circle cx="22" cy="17" r="1.5" fill="currentColor" />
          <circle cx="28" cy="17" r="1.5" fill="currentColor" />
          {/* Broken content lines */}
          <path d="M16 32H36" stroke="currentColor" strokeWidth="2" strokeLinecap="round" opacity="0.5" />
          <path d="M16 38H48" stroke="currentColor" strokeWidth="2" strokeLinecap="round" opacity="0.3" />
          <path d="M16 44H28" stroke="currentColor" strokeWidth="2" strokeLinecap="round" opacity="0.2" />
        </svg>
        {/* Red accent slash */}
        <svg
          width="24" height="24" viewBox="0 0 24 24" fill="none"
          className="absolute -bottom-1 -right-1"
        >
          <circle cx="12" cy="12" r="12" fill="#E85D3A" />
          <path d="M8 8L16 16M16 8L8 16" stroke="white" strokeWidth="2" strokeLinecap="round" />
        </svg>
      </div>

      <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5 text-center tracking-tight">
        {title}
      </h3>
      <p className="text-sm text-neutral-400 max-w-sm text-center mb-6 leading-relaxed">
        {message}
      </p>

      {onRetry && (
        <button
          onClick={onRetry}
          type="button"
          className={cn(
            'inline-flex items-center gap-2 px-5 py-2.5 rounded-lg',
            'border border-neutral-200 text-sm font-medium text-neutral-700',
            'hover:bg-neutral-50 hover:border-neutral-300',
            'active:bg-neutral-100',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
          )}
        >
          <RefreshCw className="w-4 h-4" />
          Попробовать снова
        </button>
      )}
    </div>
  )
}
