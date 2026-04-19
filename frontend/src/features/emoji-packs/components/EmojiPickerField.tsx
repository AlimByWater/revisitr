import { X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { EmojiPicker } from './EmojiPicker'
import type { EmojiItem } from '../types'

interface EmojiPickerFieldProps {
  value?: string
  onChange: (imageUrl: string) => void
  onClear?: () => void
  className?: string
}

export function EmojiPickerField({ value, onChange, onClear, className }: EmojiPickerFieldProps) {
  const handleSelect = (item: EmojiItem) => {
    onChange(item.image_url)
  }

  return (
    <div className={cn('flex items-center gap-2', className)}>
      <div className="w-8 h-8 rounded border border-neutral-200 bg-neutral-50 flex items-center justify-center overflow-hidden shrink-0">
        {value ? (
          <img src={value} alt="Выбранная иконка" className="w-full h-full object-cover" />
        ) : (
          <span className="text-xs text-neutral-400">—</span>
        )}
      </div>

      <EmojiPicker onSelect={handleSelect} selected={value} />

      {value && onClear && (
        <button
          type="button"
          onClick={onClear}
          className="flex items-center justify-center w-6 h-6 rounded text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors"
          aria-label="Убрать иконку"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      )}

      {!value && (
        <span className="text-sm text-neutral-400">Выберите иконку</span>
      )}
    </div>
  )
}
