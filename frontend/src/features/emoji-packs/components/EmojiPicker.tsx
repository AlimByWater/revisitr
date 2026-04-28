import { useState, useRef, useEffect } from 'react'
import { Smile } from 'lucide-react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useEmojiPacksQuery } from '../queries'
import { standardEmojiGroups } from '../standardEmoji'
import type { EmojiItem } from '../types'

type TabKey = 'standard' | 'custom'

interface EmojiPickerProps {
  onSelect: (item: EmojiItem) => void
  onSelectStandard?: (emoji: string) => void
  selected?: string
  triggerClassName?: string
  children?: React.ReactNode
}

export function EmojiPicker({ onSelect, onSelectStandard, selected, triggerClassName, children }: EmojiPickerProps) {
  const [open, setOpen] = useState(false)
  const [tab, setTab] = useState<TabKey>('standard')
  const [search, setSearch] = useState('')
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

  const handleSelectStandard = (emoji: string) => {
    if (onSelectStandard) {
      onSelectStandard(emoji)
    } else {
      onSelect({
        id: 0,
        pack_id: 0,
        name: emoji,
        image_url: '',
        emoji,
        sort_order: 0,
        created_at: '',
      })
    }
    setOpen(false)
  }

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
        {children || <Smile className="w-4 h-4 text-neutral-500" />}
      </button>

      <div
        role="dialog"
        aria-label="Выбор эмодзи"
        className={cn(
          'absolute top-full left-0 mt-1 z-50 bg-white border border-neutral-900 rounded shadow-md w-72 max-h-80 flex flex-col',
          'transition-all duration-150 origin-top',
          open
            ? 'opacity-100 scale-y-100 pointer-events-auto'
            : 'opacity-0 scale-y-95 pointer-events-none',
        )}
      >
        <div className="flex border-b border-neutral-200 shrink-0">
          <button
            type="button"
            onClick={() => setTab('standard')}
            className={cn(
              'flex-1 px-3 py-2 text-xs font-medium transition-colors',
              tab === 'standard'
                ? 'text-neutral-900 border-b-2 border-neutral-900'
                : 'text-neutral-400 hover:text-neutral-600',
            )}
          >
            Стандартные
          </button>
          <button
            type="button"
            onClick={() => setTab('custom')}
            className={cn(
              'flex-1 px-3 py-2 text-xs font-medium transition-colors',
              tab === 'custom'
                ? 'text-neutral-900 border-b-2 border-neutral-900'
                : 'text-neutral-400 hover:text-neutral-600',
            )}
          >
            Кастомные
          </button>
        </div>

        <div className="px-2 pt-2 shrink-0">
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Поиск..."
            className="w-full rounded border border-neutral-200 bg-neutral-50 px-2.5 py-1.5 text-xs placeholder:text-neutral-400 focus:outline-none focus:ring-1 focus:ring-accent/30 focus:border-accent"
          />
        </div>

        <div className="overflow-y-auto flex-1 min-h-0">
          {tab === 'standard' && (
            <>
              {standardEmojiGroups.map((group) => {
                const filtered = search
                  ? group.emoji.filter((e) => e.includes(search))
                  : group.emoji
                if (filtered.length === 0) return null
                return (
                  <div key={group.name}>
                    <div className="px-3 pt-3 pb-1 text-xs font-medium uppercase tracking-[0.18em] text-neutral-400">
                      {group.name}
                    </div>
                    <div className="grid grid-cols-8 gap-0.5 px-2 pb-2">
                      {filtered.map((emoji, i) => (
                        <button
                          key={`${emoji}-${i}`}
                          type="button"
                          onClick={() => handleSelectStandard(emoji)}
                          className="w-8 h-8 rounded flex items-center justify-center text-lg hover:bg-neutral-100 transition-colors"
                          title={emoji}
                        >
                          {emoji}
                        </button>
                      ))}
                    </div>
                  </div>
                )
              })}
              {search && standardEmojiGroups.every((g) => g.emoji.every((e) => !e.includes(search))) && (
                <div className="p-4 text-sm text-neutral-400 text-center">Ничего не найдено</div>
              )}
            </>
          )}

          {tab === 'custom' && (
            <>
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
                const items = (pack.items ?? []).filter((item) =>
                  !search || item.name.toLowerCase().includes(search.toLowerCase()),
                )
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

              {!isLoading && packs.length > 0 && packs.every((p) => {
                const items = (p.items ?? []).filter((item) =>
                  !search || item.name.toLowerCase().includes(search.toLowerCase()),
                )
                return items.length === 0
              }) && (
                <div className="p-4 text-sm text-neutral-400 text-center">
                  {search ? 'Ничего не найдено' : 'Нет эмодзи в паках'}
                </div>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}
