import { useState, useRef, useEffect } from 'react'
import { Smile } from 'lucide-react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useEmojiPacksQuery } from '../queries'
import type { EmojiItem } from '../types'

interface EmojiPickerProps {
  onSelect: (item: EmojiItem) => void
  selected?: string
  triggerClassName?: string
}

export function EmojiPicker({ onSelect, selected, triggerClassName }: EmojiPickerProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)
  const { data: packs = [], isLoading } = useEmojiPacksQuery()

  useEffect(() => {
    function handleClick(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    function handleKey(e: KeyboardEvent) {
      if (e.key === 'Escape') setOpen(false)
    }
    document.addEventListener('mousedown', handleClick)
    document.addEventListener('keydown', handleKey)
    return () => {
      document.removeEventListener('mousedown', handleClick)
      document.removeEventListener('keydown', handleKey)
    }
  }, [])

  const allItems = packs.flatMap((p) => p.items ?? [])

  return (
    <div ref={ref} className="relative inline-block">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          'flex items-center justify-center w-8 h-8 rounded border border-neutral-200 bg-white hover:bg-neutral-50 transition-colors',
          triggerClassName,
        )}
        aria-label="Выбрать эмодзи"
      >
        <Smile className="w-4 h-4 text-neutral-500" />
      </button>

      <div
        role="dialog"
        aria-label="Выбор эмодзи"
        className={cn(
          'absolute top-full left-0 mt-1 z-50 bg-white border border-neutral-900 rounded shadow-md w-64 max-h-72 overflow-y-auto',
          'transition-all duration-150 origin-top',
          open
            ? 'opacity-100 scale-y-100 pointer-events-auto'
            : 'opacity-0 scale-y-95 pointer-events-none',
        )}
      >
        {isLoading && (
          <div className="p-4 text-sm text-neutral-400 text-center">Загрузка...</div>
        )}

        {!isLoading && packs.length === 0 && (
          <div className="p-4 text-sm text-neutral-500 text-center">
            Нет паков.{' '}
            <Link
              to="/dashboard/emoji-packs"
              className="text-accent hover:text-accent-hover font-medium"
              onClick={() => setOpen(false)}
            >
              Создать →
            </Link>
          </div>
        )}

        {!isLoading && packs.map((pack) => {
          const items = pack.items ?? []
          if (items.length === 0) return null
          return (
            <div key={pack.id}>
              <div className="px-3 pt-3 pb-1 text-xs font-medium uppercase tracking-[0.18em] text-neutral-400">
                {pack.name}
              </div>
              <div className="grid grid-cols-4 gap-1 px-2 pb-2">
                {items.map((item) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => { onSelect(item); setOpen(false) }}
                    className={cn(
                      'w-12 h-12 rounded overflow-hidden border-2 transition-colors',
                      selected === item.image_url
                        ? 'border-accent'
                        : 'border-transparent hover:border-neutral-300',
                    )}
                    title={item.name}
                  >
                    <img
                      src={item.image_url}
                      alt={item.name}
                      className="w-full h-full object-cover"
                    />
                  </button>
                ))}
              </div>
            </div>
          )
        })}

        {!isLoading && packs.length > 0 && allItems.length === 0 && (
          <div className="p-4 text-sm text-neutral-400 text-center">Нет эмодзи в паках</div>
        )}
      </div>
    </div>
  )
}
