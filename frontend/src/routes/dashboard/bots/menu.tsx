import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBotQuery, useBotModuleSettingsQuery, useModulePresetsQuery, useResetPresetMutation, useSelectPresetMutation, useUpdateCustomizationsMutation } from '@/features/bots/queries'
import type { MenuPresetCustomizations } from '@/features/bots/types'
import { MenuPresetCustomizer, createDefaultMenuPresetCustomizations, isMenuPresetCustomizationDirty, normalizeMenuPresetCustomizations, sanitizeMenuPresetCustomizations } from '@/features/bots/components/MenuPresetCustomizer'
import { PresetGallery } from '@/features/bots/components/PresetGallery'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { useMenusQuery } from '@/features/menus/queries'
import { menusApi } from '@/features/menus/api'
import type { Menu } from '@/features/menus/types'

function pickBotMenu(menus: Menu[], boundPosIds: number[]): Menu | null {
  if (menus.length === 0) return null

  if (boundPosIds.length > 0) {
    const boundMenu = menus.find((menu) =>
      (menu.bindings ?? []).some((binding) => binding.is_active && boundPosIds.includes(binding.pos_id)),
    )
    if (boundMenu) return boundMenu
  }

  return menus.find((menu) => (menu.categories ?? []).length > 0) ?? menus[0]
}

export default function BotMenuSettingsPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)

  const { data: bot, isLoading: botLoading, isError: botError } = useBotQuery(Number.isNaN(id) ? 0 : id)
  const { data: presets = [], isLoading: presetsLoading } = useModulePresetsQuery('menu')
  const { data: menus = [], isLoading: menusLoading } = useMenusQuery()
  const {
    data: moduleSettings,
    isLoading: settingsLoading,
    mutate: mutateSettings,
  } = useBotModuleSettingsQuery(Number.isNaN(id) ? 0 : id, 'menu')

  const selectPreset = useSelectPresetMutation(id, 'menu')
  const updateCustomizations = useUpdateCustomizationsMutation(id, 'menu')
  const resetPreset = useResetPresetMutation(id, 'menu')

  const [boundPosIds, setBoundPosIds] = useState<number[]>([])
  const [posBindingsLoaded, setPosBindingsLoaded] = useState(false)
  const [selectedPresetKey, setSelectedPresetKey] = useState('')
  const [draft, setDraft] = useState<MenuPresetCustomizations>({})
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    if (!saveSuccess) return
    const timeout = setTimeout(() => setSaveSuccess(false), 3000)
    return () => clearTimeout(timeout)
  }, [saveSuccess])

  useEffect(() => {
    if (!id || Number.isNaN(id) || id <= 0) return
    let mounted = true
    menusApi
      .getBotPOSLocations(id)
      .then((response) => {
        if (!mounted) return
        setBoundPosIds(response.pos_ids ?? [])
        setPosBindingsLoaded(true)
      })
      .catch(() => {
        if (!mounted) return
        setBoundPosIds([])
        setPosBindingsLoaded(true)
      })
    return () => {
      mounted = false
    }
  }, [id])

  const activeMenu = useMemo(
    () => pickBotMenu(menus, boundPosIds),
    [boundPosIds, menus],
  )

  const normalizedFromServer = useMemo(
    () => normalizeMenuPresetCustomizations(moduleSettings?.customizations, activeMenu),
    [activeMenu, moduleSettings?.customizations],
  )

  useEffect(() => {
    const fallbackPresetKey = moduleSettings?.preset_key || presets[0]?.preset_key || 'tabs'
    setSelectedPresetKey(fallbackPresetKey)
    setDraft(normalizedFromServer)
  }, [moduleSettings?.preset_key, normalizedFromServer, presets])

  const hasDraftChanges = useMemo(() => {
    const presetChanged = selectedPresetKey !== (moduleSettings?.preset_key || presets[0]?.preset_key || 'tabs')
    const customizationsChanged =
      JSON.stringify(sanitizeMenuPresetCustomizations(draft, activeMenu)) !==
      JSON.stringify(sanitizeMenuPresetCustomizations(normalizedFromServer, activeMenu))
    return presetChanged || customizationsChanged
  }, [activeMenu, draft, moduleSettings?.preset_key, normalizedFromServer, presets, selectedPresetKey])

  if (botLoading || presetsLoading || settingsLoading || menusLoading || !posBindingsLoaded) {
    return <CardSkeleton />
  }
  if (botError || !bot) return <ErrorState title="Бот не найден" />

  const handleLocalReset = () => {
    setDraft(createDefaultMenuPresetCustomizations(activeMenu))
    setSaveError(null)
    setSaveSuccess(false)
  }

  const handleSelectPreset = (presetKey: string) => {
    setSelectedPresetKey(presetKey)
    setDraft(createDefaultMenuPresetCustomizations(activeMenu))
    setSaveError(null)
    setSaveSuccess(false)
  }

  const handleSave = async () => {
    setIsSaving(true)
    setSaveError(null)
    setSaveSuccess(false)

    try {
      const currentPresetKey = moduleSettings?.preset_key || ''
      if (selectedPresetKey && selectedPresetKey !== currentPresetKey) {
        await selectPreset.mutateAsync(selectedPresetKey)
      }

      const sanitized = sanitizeMenuPresetCustomizations(draft, activeMenu)
      const hasCustomizations = isMenuPresetCustomizationDirty(draft, activeMenu)

      if (hasCustomizations) {
        await updateCustomizations.mutateAsync(sanitized)
      } else if (moduleSettings?.customized && selectedPresetKey === currentPresetKey) {
        await resetPreset.mutateAsync(undefined as never)
      }

      await mutateSettings()
      setSaveSuccess(true)
    } catch {
      setSaveError('Не удалось сохранить пресет меню. Попробуйте ещё раз.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="mx-auto max-w-5xl space-y-6 py-6">
      <div className="flex items-center gap-3">
        <Link
          to={`/dashboard/bots/${id}?tab=modules`}
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-surface-border text-neutral-500 transition-colors hover:text-neutral-700"
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
          currentSettings={
            moduleSettings
              ? {
                  ...moduleSettings,
                  preset_key: selectedPresetKey || moduleSettings.preset_key,
                }
              : null
          }
          onSelect={handleSelectPreset}
          onReset={handleLocalReset}
          isSelecting={isSaving}
          isResetting={isSaving}
        />
      </div>

      <div className="rounded-xl border border-surface-border bg-white p-5">
        <div className="mb-4">
          <h3 className="text-sm font-semibold text-neutral-900">Кастомизация</h3>
          <p className="mt-1 text-xs text-neutral-500">
            Меняйте табы и тексты. Логика переходов остаётся фиксированной внутри пресета.
          </p>
        </div>

        <MenuPresetCustomizer
          menu={activeMenu}
          value={draft}
          onChange={setDraft}
        />

        <div className="mt-6 flex flex-col gap-3 border-t border-surface-border pt-5 sm:flex-row sm:items-center sm:justify-between">
          <div className="min-h-5 text-sm">
            {saveError && <span className="text-red-600">{saveError}</span>}
            {saveSuccess && <span className="text-green-600">Пресет сохранён</span>}
          </div>

          <div className="flex flex-col gap-2 sm:flex-row">
            <button
              type="button"
              onClick={handleLocalReset}
              className="inline-flex min-h-11 items-center justify-center rounded-lg border border-surface-border px-4 text-sm font-medium text-neutral-600 transition-colors hover:bg-neutral-50"
            >
              Сбросить локально
            </button>
            <button
              type="button"
              onClick={handleSave}
              disabled={isSaving || !hasDraftChanges}
              className={cn(
                'inline-flex min-h-11 items-center justify-center rounded-lg px-5 text-sm font-medium text-white transition-colors',
                'bg-accent hover:bg-accent/90 disabled:cursor-not-allowed disabled:opacity-50',
              )}
            >
              {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
            </button>
          </div>
        </div>
      </div>

      <div className="rounded-xl border border-surface-border bg-white p-5">
        <h3 className="text-sm font-semibold text-neutral-900">Управление меню</h3>
        <p className="mt-1 text-xs text-neutral-500">
          Категории, позиции, цены и привязка к точкам продаж редактируются в отдельном разделе.
        </p>
        <Link
          to={`/dashboard/menus?botId=${id}`}
          className="mt-3 inline-flex items-center gap-1.5 text-sm font-medium text-accent transition-colors hover:text-accent/80"
        >
          Открыть управление меню →
        </Link>
      </div>
    </div>
  )
}
