import { Suspense, lazy, useCallback, useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, CircleAlert, CircleCheckBig, LayoutTemplate, ListTree, UtensilsCrossed } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBotQuery, useBotModuleSettingsQuery, useModulePresetsQuery, useResetPresetMutation, useSelectPresetMutation, useUpdateCustomizationsMutation } from '@/features/bots/queries'
import type { MenuPresetCustomizations } from '@/features/bots/types'
import { MenuPresetCustomizer, createDefaultMenuPresetCustomizations, isMenuPresetCustomizationDirty, normalizeMenuPresetCustomizations, sanitizeMenuPresetCustomizations } from '@/features/bots/components/MenuPresetCustomizer'
import { PresetGallery } from '@/features/bots/components/PresetGallery'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'
import { InfoHint } from '@/components/common/InfoHint'
import { useMenusQuery, useMenuQuery, useUpdateMenuMutation } from '@/features/menus/queries'
import { menusApi } from '@/features/menus/api'
import type { Menu } from '@/features/menus/types'
import type { InlineButton, MessageContent, PreviewScreen } from '@/features/telegram-preview/types'
import { InteractivePreview } from '@/features/telegram-preview/components/InteractivePreview'
import { campaignsApi } from '@/features/campaigns/api'
import {
  buildCategoryContent,
  buildMenuInitialContent,
  buildMenuListContent,
  buildItemCardContent,
  buildMenuPreviewState,
  buildMenuTabsContent,
} from './menuPreview'

const MessageContentEditor = lazy(() =>
  import('@/features/telegram-preview/components/MessageContentEditor').then((mod) => ({
    default: mod.MessageContentEditor,
  })),
)

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

interface MenuDisplaySettingsPanelProps {
  botId: number
  activeMenuId?: number
  embedded?: boolean
}

export default function BotMenuSettingsPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)

  return <MenuDisplaySettingsPanel botId={id} />
}

export function MenuDisplaySettingsPanel({
  botId,
  activeMenuId,
  embedded = false,
}: MenuDisplaySettingsPanelProps) {
  const id = botId

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

  const activeMenuStub = useMemo(
    () => (activeMenuId ? null : pickBotMenu(menus, boundPosIds)),
    [activeMenuId, boundPosIds, menus],
  )

  const { data: fullMenu, mutate: mutateFullMenu } = useMenuQuery(activeMenuId ?? activeMenuStub?.id ?? 0)
  const activeMenu = fullMenu ?? activeMenuStub
  const updateMenu = useUpdateMenuMutation()

  const [introContent, setIntroContent] = useState<MessageContent | null>(null)

  useEffect(() => {
    if (activeMenu?.intro_content) {
      setIntroContent(activeMenu.intro_content)
    }
  }, [activeMenu?.intro_content])

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

  const selectedPreset = useMemo(
    () => presets.find((preset) => preset.preset_key === selectedPresetKey) ?? presets[0] ?? null,
    [presets, selectedPresetKey],
  )

  const categoryCount = draft.categories?.length ?? activeMenu?.categories?.length ?? 0
  const activeMenuName = activeMenu?.name || 'Меню не выбрано'
  const menuEditorHref = activeMenu
    ? `/dashboard/menus/${activeMenu.id}?botId=${id}&tab=content`
    : `/dashboard/menus?botId=${id}`

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

      if (activeMenu && introContent) {
        await updateMenu.mutateAsync({
          id: activeMenu.id,
          data: { intro_content: introContent },
        })
        await mutateFullMenu()
      }

      await mutateSettings()
      setSaveSuccess(true)
    } catch {
      setSaveError('Не удалось сохранить. Попробуйте ещё раз.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className={cn(embedded ? 'space-y-6' : 'mx-auto max-w-7xl space-y-6 py-6')}>
      {!embedded && (
        <header className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
          <div className="flex items-start gap-3">
            <Link
              to={`/dashboard/bots/${id}?tab=modules`}
              className="mt-1 flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-surface-border bg-white text-neutral-500 transition-colors hover:text-neutral-700"
            >
              <ArrowLeft className="h-4 w-4" />
            </Link>
            <div className="space-y-1">
              <div className="flex flex-wrap items-center gap-2">
                <h1 className="text-2xl font-semibold tracking-tight text-neutral-900">
                  Настройки модуля «Меню»
                </h1>
                <span
                  className={cn(
                    'inline-flex min-h-7 items-center rounded-full border px-2.5 text-xs font-medium',
                    hasDraftChanges
                      ? 'border-amber-200 bg-amber-50 text-amber-700'
                      : 'border-emerald-200 bg-emerald-50 text-emerald-700',
                  )}
                >
                  {hasDraftChanges ? 'Есть несохранённые изменения' : 'Изменения сохранены'}
                </span>
              </div>
              <p className="text-sm text-neutral-500">{bot.name}</p>
            </div>
          </div>
        </header>
      )}

      <div className="sticky top-4 z-10 rounded-2xl border border-surface-border bg-white/95 p-4 shadow-sm backdrop-blur">
        <div className="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div className="grid gap-3 sm:grid-cols-3">
            <div className="rounded-xl bg-neutral-50 px-3 py-2">
              <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Шаблон</div>
              <div className="mt-1 flex items-center gap-2 text-sm font-medium text-neutral-900">
                <LayoutTemplate className="h-4 w-4 text-neutral-400" />
                <span>{selectedPreset?.name || 'Не выбран'}</span>
              </div>
            </div>
            {activeMenu ? (
              <Link
                to={menuEditorHref}
                className="rounded-xl bg-neutral-50 px-3 py-2 transition-colors hover:bg-neutral-100"
              >
                <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Меню</div>
                <div className="mt-1 flex items-center gap-2 text-sm font-medium text-neutral-900">
                  <UtensilsCrossed className="h-4 w-4 text-neutral-400" />
                  <span>{activeMenuName}</span>
                </div>
              </Link>
            ) : (
              <div className="rounded-xl bg-neutral-50 px-3 py-2">
                <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Меню</div>
                <div className="mt-1 flex items-center gap-2 text-sm font-medium text-neutral-900">
                  <UtensilsCrossed className="h-4 w-4 text-neutral-400" />
                  <span>{activeMenuName}</span>
                </div>
              </div>
            )}
            <div className="rounded-xl bg-neutral-50 px-3 py-2">
              <div className="text-[11px] font-medium uppercase tracking-[0.18em] text-neutral-400">Категории</div>
              <div className="mt-1 flex items-center gap-2 text-sm font-medium text-neutral-900">
                <ListTree className="h-4 w-4 text-neutral-400" />
                <span>{categoryCount}</span>
              </div>
            </div>
          </div>

          <div className="flex flex-col gap-3 lg:min-w-[22rem]">
            <div className="min-h-5 text-sm">
              {saveError && (
                <span className="inline-flex items-center gap-1.5 text-red-600">
                  <CircleAlert className="h-4 w-4" />
                  {saveError}
                </span>
              )}
              {saveSuccess && (
                <span className="inline-flex items-center gap-1.5 text-emerald-600">
                  <CircleCheckBig className="h-4 w-4" />
                  Пресет сохранён
                </span>
              )}
            </div>

            <div className="flex flex-col gap-2 sm:flex-row sm:justify-end">
              <button
                type="button"
                onClick={handleLocalReset}
                className="inline-flex min-h-11 items-center justify-center rounded-xl border border-surface-border px-4 text-sm font-medium text-neutral-600 transition-colors hover:bg-neutral-50"
              >
                Сбросить локально
              </button>
              <button
                type="button"
                onClick={handleSave}
                disabled={isSaving || !hasDraftChanges}
                className={cn(
                  'inline-flex min-h-11 items-center justify-center rounded-xl px-5 text-sm font-medium text-white transition-colors',
                  'bg-accent hover:bg-accent/90 disabled:cursor-not-allowed disabled:opacity-50',
                )}
              >
                {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
              </button>
            </div>
          </div>
        </div>
      </div>

      <div className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_22rem]">
        <main className="space-y-6">
          <section className="rounded-2xl border border-surface-border bg-white p-6">
            <div className="mb-5 flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
              <div className="max-w-2xl">
                <div className="flex items-center gap-2">
                  <h2 className="text-lg font-semibold text-neutral-900">Способ показа меню</h2>
                  <InfoHint content="Выберите формат: табы, список или карусель. Детали текста и категорий настраиваются ниже." />
                </div>
              </div>
            </div>

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
          </section>

          <section className="rounded-2xl border border-surface-border bg-white p-6">
            <div className="mb-5 max-w-2xl">
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold text-neutral-900">Тексты, кнопки и категории</h2>
                <InfoHint content="Названия и иконки берутся из меню и могут быть переопределены для бота." />
              </div>
              <p className="mt-2 text-sm text-neutral-500">
                Настройте как меню выглядит в боте. Оригиналы берутся из{' '}
                {activeMenu ? (
                  <Link to={menuEditorHref} className="text-accent hover:text-accent/80 font-medium">
                    редактора меню
                  </Link>
                ) : (
                  <span>редактора меню</span>
                )}.
              </p>
            </div>

            <MenuPresetCustomizer
              menu={activeMenu}
              value={draft}
              onChange={setDraft}
            />
          </section>

          <section className="rounded-2xl border border-surface-border bg-white p-6">
            <div className="mb-5 max-w-2xl">
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold text-neutral-900">Приветственное сообщение</h2>
                <InfoHint content="Текст и медиа, которые бот отправит, когда гость нажмёт кнопку «Меню»." />
              </div>
            </div>
            <div className="rounded-xl border border-surface-border bg-neutral-50/70 p-4">
              <Suspense fallback={<div className="h-32 rounded border border-neutral-200 bg-neutral-50/60 animate-pulse" />}>
                <MessageContentEditor
                  value={introContent ?? { parts: [{ type: 'text', text: '', parse_mode: 'Markdown' }] }}
                  onChange={setIntroContent}
                  onUpload={campaignsApi.uploadFile}
                  maxParts={5}
                />
              </Suspense>
            </div>
          </section>
        </main>

        <aside className="space-y-4 xl:sticky xl:top-28 xl:self-start">
          <MenuTelegramPreview
            botName={bot.name}
            draft={draft}
            menu={activeMenu}
            introContent={introContent}
            selectedPresetKey={selectedPresetKey}
          />

          <section className="rounded-2xl border border-surface-border bg-white p-5">
            <h3 className="text-sm font-semibold text-neutral-900">Управление составом меню</h3>
            <p className="mt-2 text-sm text-neutral-500">
              Цены, блюда, фото и привязку к точкам продаж редактируйте отдельно.
            </p>
            <Link
              to={menuEditorHref}
              className="mt-3 inline-flex items-center gap-1.5 text-sm font-medium text-accent transition-colors hover:text-accent/80"
            >
              {activeMenu ? `Редактировать «${activeMenu.name}»` : 'Открыть управление меню'} →
            </Link>
          </section>
        </aside>
      </div>
    </div>
  )
}

function MenuTelegramPreview({
  botName,
  draft,
  menu,
  introContent,
  selectedPresetKey,
}: {
  botName: string
  draft: MenuPresetCustomizations
  menu: Menu | null
  introContent: MessageContent | null
  selectedPresetKey: string
}) {
  const previewState = useMemo(
    () => buildMenuPreviewState({ introContent, menu, draft, selectedPresetKey }),
    [draft, introContent, menu, selectedPresetKey],
  )

  const initialContent = useMemo(
    () => (previewState.isListPreset ? buildMenuListContent(previewState) : buildMenuInitialContent(previewState)),
    [previewState],
  )

  const handleButtonClick = useCallback(
    (button: InlineButton): PreviewScreen | null => {
      if (button.data === 'menu:root') {
        return { content: initialContent, transition: 'replace' }
      }

      const tabMatch = button.data?.match(/^menu:tab:(\d+)$/)
      if (tabMatch) {
        return {
          content: buildMenuTabsContent(previewState, Number(tabMatch[1])),
          transition: 'replace',
        }
      }

      const categoryMatch = button.data?.match(/^menu:cat:(\d+):\d+$/)
      if (categoryMatch) {
        const categoryId = Number(categoryMatch[1])
        const category = previewState.categories.find((c) => c.id === categoryId)
        if (!category) return null
        return {
          content: buildCategoryContent(category),
          transition: 'replace',
        }
      }

      const itemMatch = button.data?.match(/^menu:item:(\d+):(\d+):\d+$/)
      if (itemMatch) {
        const itemId = Number(itemMatch[1])
        const categoryId = Number(itemMatch[2])
        const category = previewState.categories.find((c) => c.id === categoryId)
        if (!category) return null
        const content = buildItemCardContent(category, itemId)
        if (!content) return null
        return {
          content,
          transition: 'replace',
        }
      }

      const navMatch = button.data?.match(/^menu:cardnav:(\d+):(\d+)$/)
      if (navMatch) {
        const itemId = Number(navMatch[1])
        const categoryId = Number(navMatch[2])
        const category = previewState.categories.find((c) => c.id === categoryId)
        if (!category) return null
        const content = buildItemCardContent(category, itemId)
        if (!content) return null
        return {
          content,
          transition: 'replace',
        }
      }

      if (button.data === 'menu:noop') {
        return null
      }

      if (button.data === 'menu:cardclose') {
        return { content: initialContent, transition: 'replace' }
      }

      return null
    },
    [initialContent, previewState],
  )

  return (
    <section className="rounded-2xl border border-surface-border bg-white p-4">
      <h3 className="mb-3 text-sm font-semibold text-neutral-900">Превью в Telegram</h3>
      <InteractivePreview
        initialContent={initialContent}
        botName={botName}
        onButtonClick={handleButtonClick}
      />
    </section>
  )
}
