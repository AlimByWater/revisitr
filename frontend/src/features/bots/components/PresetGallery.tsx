import { cn } from '@/lib/utils'
import { Check, RotateCcw } from 'lucide-react'
import type { ModulePreset, BotModuleSettings } from '../types'

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
          <p className="mt-0.5 text-sm text-neutral-500">
            Выберите формат, затем переходите к локальным правкам текста и категорий.
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

      <div className="grid grid-cols-1 gap-3 lg:grid-cols-3">
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
                'relative flex h-full flex-col items-start gap-3 rounded-2xl border p-4 text-left',
                'transition-all hover:shadow-sm disabled:opacity-50 disabled:cursor-not-allowed',
                isActive
                  ? 'border-accent bg-accent/5 shadow-sm'
                  : 'border-surface-border bg-white hover:border-neutral-300',
              )}
            >
              {isActive && (
                <div className="absolute right-3 top-3 flex h-6 w-6 items-center justify-center rounded-full bg-accent text-white">
                  <Check className="h-3 w-3" />
                </div>
              )}
              <PresetThumbnail presetKey={preset.preset_key} />
              <div className="space-y-1">
                <div className="text-base font-semibold text-neutral-900">{preset.name}</div>
                <div className="text-sm leading-relaxed text-neutral-500">{preset.description}</div>
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}

function PresetThumbnail({ presetKey }: { presetKey: string }) {
  if (presetKey === 'tabs') {
    return (
      <div className="w-full rounded-xl border border-surface-border bg-neutral-50 p-3">
        <div className="mb-3 grid grid-cols-3 gap-1.5">
          <div className="h-6 rounded-md bg-accent/20" />
          <div className="h-6 rounded-md border border-surface-border bg-white" />
          <div className="h-6 rounded-md border border-surface-border bg-white" />
        </div>
        <div className="space-y-1.5">
          <div className="h-2.5 rounded bg-neutral-300" />
          <div className="h-2.5 rounded bg-neutral-200" />
          <div className="h-2.5 w-4/5 rounded bg-neutral-200" />
        </div>
      </div>
    )
  }

  if (presetKey === 'carousel') {
    return (
      <div className="w-full rounded-xl border border-surface-border bg-neutral-50 p-3">
        <div className="mb-3 h-16 rounded-lg bg-neutral-300" />
        <div className="space-y-1.5">
          <div className="h-2.5 rounded bg-neutral-300" />
          <div className="h-2.5 w-3/4 rounded bg-neutral-200" />
        </div>
        <div className="mt-3 flex items-center justify-between rounded-lg border border-surface-border bg-white px-2 py-1.5">
          <span className="text-[10px] text-neutral-400">←</span>
          <span className="text-[10px] text-neutral-400">1/5</span>
          <span className="text-[10px] text-neutral-400">→</span>
        </div>
      </div>
    )
  }

  return (
    <div className="w-full rounded-xl border border-surface-border bg-neutral-50 p-3">
      <div className="space-y-1.5">
        <div className="h-2.5 rounded bg-neutral-300" />
        <div className="h-2.5 rounded bg-neutral-200" />
        <div className="h-2.5 rounded bg-neutral-200" />
        <div className="h-2.5 w-5/6 rounded bg-neutral-200" />
      </div>
    </div>
  )
}
