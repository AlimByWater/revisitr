import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBotQuery } from '@/features/bots/queries'
import {
  useModulePresetsQuery,
  useBotModuleSettingsQuery,
  useSelectPresetMutation,
  useResetPresetMutation,
} from '@/features/bots/queries'
import { PresetGallery } from '@/features/bots/components/PresetGallery'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'

export default function BotMenuSettingsPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)

  const { data: bot, isLoading: botLoading, isError: botError } = useBotQuery(Number.isNaN(id) ? 0 : id)
  const { data: presets = [], isLoading: presetsLoading } = useModulePresetsQuery('menu')
  const {
    data: moduleSettings,
    isLoading: settingsLoading,
    mutate: mutateSettings,
  } = useBotModuleSettingsQuery(Number.isNaN(id) ? 0 : id, 'menu')

  const selectPreset = useSelectPresetMutation(id, 'menu')
  const resetPreset = useResetPresetMutation(id, 'menu')

  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    if (!saveSuccess) return
    const t = setTimeout(() => setSaveSuccess(false), 3000)
    return () => clearTimeout(t)
  }, [saveSuccess])

  if (botLoading || presetsLoading || settingsLoading) return <CardSkeleton />
  if (botError || !bot) return <ErrorState title="Бот не найден" />

  const handleSelect = async (presetKey: string) => {
    try {
      await selectPreset.mutateAsync(presetKey)
      await mutateSettings()
      setSaveSuccess(true)
    } catch {
      // handled by SWR
    }
  }

  const handleReset = async () => {
    try {
      await resetPreset.mutateAsync(undefined as never)
      await mutateSettings()
      setSaveSuccess(true)
    } catch {
      // handled by SWR
    }
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6 py-6">
      <div className="flex items-center gap-3">
        <Link
          to={`/dashboard/bots/${id}`}
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-surface-border text-neutral-500 hover:text-neutral-700 transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-lg font-semibold text-neutral-900">
            Настройки модуля «Меню»
          </h1>
          <p className="text-sm text-neutral-500">{bot.name}</p>
        </div>
      </div>

      <div className="rounded-xl border border-surface-border bg-white p-5">
        <PresetGallery
          presets={presets}
          currentSettings={moduleSettings ?? null}
          onSelect={handleSelect}
          onReset={handleReset}
          isSelecting={selectPreset.isMutating}
          isResetting={resetPreset.isMutating}
        />
      </div>

      <div className="rounded-xl border border-surface-border bg-white p-5">
        <h3 className="text-sm font-semibold text-neutral-900">Управление меню</h3>
        <p className="mt-1 text-xs text-neutral-500">
          Категории, позиции, цены и привязка к точкам продаж
        </p>
        <Link
          to={`/dashboard/menus?botId=${id}`}
          className="mt-3 inline-flex items-center gap-1.5 text-sm font-medium text-accent hover:text-accent/80 transition-colors"
        >
          Открыть управление меню →
        </Link>
      </div>

      {saveSuccess && (
        <div className="text-sm text-green-600">
          Шаблон сохранён. Изменения вступят в силу при следующем взаимодействии пользователя с ботом.
        </div>
      )}
    </div>
  )
}
