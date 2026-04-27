import { Trash2 } from 'lucide-react'
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
      <EmojiPicker
        onSelect={handleSelect}
        selected={value}
        triggerClassName="!w-[38px] !h-[38px] shrink-0"
      />
      {value && onClear && (
        <button
          type="button"
          onClick={onClear}
          className="flex items-center justify-center w-7 h-7 rounded text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors"
          aria-label="Убрать иконку"
          title="Убрать иконку"
        >
          <Trash2 className="w-3.5 h-3.5" />
        </button>
      )}
    </div>
  )
}
