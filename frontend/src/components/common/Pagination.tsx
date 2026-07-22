import { ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react'
import { cn } from '@/lib/utils'

interface PaginationProps {
  page: number
  pageCount: number
  onChange: (page: number) => void
  total?: number
  itemsLabel?: string
  className?: string
}

export function Pagination({ page, pageCount, onChange, total, itemsLabel, className }: PaginationProps) {
  const canPrev = page > 1
  const canNext = page < pageCount

  return (
    <div className={cn('flex flex-wrap items-center justify-between gap-3 mt-4', className)}>
      <p className="text-sm text-neutral-500">
        {total !== undefined && itemsLabel
          ? `Всего ${total.toLocaleString('ru-RU')} ${itemsLabel}`
          : `Стр. ${page} из ${pageCount}`}
      </p>
      <div className="flex items-center gap-1">
        <PageButton onClick={() => onChange(1)} disabled={!canPrev} ariaLabel="Первая страница">
          <ChevronsLeft className="w-4 h-4" />
        </PageButton>
        <PageButton onClick={() => onChange(page - 1)} disabled={!canPrev} ariaLabel="Предыдущая страница">
          <ChevronLeft className="w-4 h-4" />
        </PageButton>
        <span className="px-3 py-2 text-sm text-neutral-700 tabular-nums">
          {page} / {pageCount || 1}
        </span>
        <PageButton onClick={() => onChange(page + 1)} disabled={!canNext} ariaLabel="Следующая страница">
          <ChevronRight className="w-4 h-4" />
        </PageButton>
        <PageButton onClick={() => onChange(pageCount)} disabled={!canNext} ariaLabel="Последняя страница">
          <ChevronsRight className="w-4 h-4" />
        </PageButton>
      </div>
    </div>
  )
}

function PageButton({
  onClick,
  disabled,
  ariaLabel,
  children,
}: {
  onClick: () => void
  disabled: boolean
  ariaLabel: string
  children: React.ReactNode
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={ariaLabel}
      className={cn(
        'inline-flex items-center justify-center p-2 min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0 rounded text-neutral-500 hover:bg-neutral-100 transition-colors',
        'disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:bg-transparent',
      )}
    >
      {children}
    </button>
  )
}
