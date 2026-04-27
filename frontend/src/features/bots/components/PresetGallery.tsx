import { cn } from '@/lib/utils'
import { Check, RotateCcw } from 'lucide-react'
import type { ModulePreset, BotModuleSettings } from '../types'

const presetIcons: Record<string, string> = {
  tabs: '📑',
  list: '📋',
  carousel: '🎠',
}

interface PresetGalleryProps {
  presets: ModulePreset[]
  currentSettings: BotModuleSettings | null
  onSelect: (presetKey: string) => void
  onReset: () => void
  isSelecting: boolean
  isResetting: boolean
}

export function PresetGallery({
  presets,
  currentSettings,
  onSelect,
  onReset,
  isSelecting,
  isResetting,
}: PresetGalleryProps) {
  const activeKey = currentSettings?.preset_key || ''
  const isCustomized = currentSettings?.customized || false

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-semibold text-neutral-900">Шаблон отображения</h3>
          <p className="mt-0.5 text-xs text-neutral-500">
            Выберите как меню будет выглядеть в боте
          </p>
        </div>
        {isCustomized && (
          <button
            type="button"
            onClick={onReset}
            disabled={isResetting}
            className={cn(
              'inline-flex items-center gap-1.5 rounded-lg px-3 py-1.5 text-xs font-medium',
              'border border-surface-border text-neutral-600 hover:bg-neutral-50',
              'transition-colors disabled:opacity-50',
            )}
          >
            <RotateCcw className="h-3 w-3" />
            {isResetting ? 'Сброс...' : 'Сбросить'}
          </button>
        )}
      </div>

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
        {presets.map((preset) => {
          const isActive = preset.preset_key === activeKey
          return (
            <button
              key={preset.preset_key}
              type="button"
              onClick={() => {
                if (!isActive) onSelect(preset.preset_key)
              }}
              disabled={isSelecting || isResetting}
              className={cn(
                'relative flex flex-col items-center gap-2 rounded-xl border-2 p-4 text-center',
                'transition-all hover:shadow-sm disabled:opacity-50 disabled:cursor-not-allowed',
                isActive
                  ? 'border-accent bg-accent/5 shadow-sm'
                  : 'border-surface-border bg-white hover:border-neutral-300',
              )}
            >
              {isActive && (
                <div className="absolute right-2 top-2 flex h-5 w-5 items-center justify-center rounded-full bg-accent text-white">
                  <Check className="h-3 w-3" />
                </div>
              )}
              <span className="text-2xl">{presetIcons[preset.preset_key] || '📄'}</span>
              <span className="text-sm font-medium text-neutral-900">{preset.name}</span>
              <span className="text-xs text-neutral-500 leading-relaxed">{preset.description}</span>
            </button>
          )
        })}
      </div>
    </div>
  )
}
