import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { cn } from '@/lib/utils'
import { botsApi } from '@/features/bots/api'
import { useBotQuery } from '@/features/bots/queries'
import { normalizeModuleConfigs } from '@/features/bots/settings'
import type { ModuleConfigs } from '@/features/bots/types'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'

const inputClassName = cn(
  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
  'text-sm placeholder:text-neutral-400 bg-white',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors disabled:opacity-50 disabled:cursor-not-allowed',
)

interface FeedbackDraft {
  prompt_message: string
  success_message: string
}

function useSaveAction(updater: () => Promise<void>) {
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    if (!saveSuccess) return
    const t = setTimeout(() => setSaveSuccess(false), 3000)
    return () => clearTimeout(t)
  }, [saveSuccess])

  const save = async () => {
    setIsSaving(true)
    setSaveError(null)
    setSaveSuccess(false)
    try {
      await updater()
      setSaveSuccess(true)
    } catch {
      setSaveError('Не удалось сохранить изменения. Попробуйте ещё раз.')
    } finally {
      setIsSaving(false)
    }
  }

  return { isSaving, saveError, saveSuccess, save }
}

function SaveFooter({
  isSaving,
  saveError,
  saveSuccess,
  onSave,
}: {
  isSaving: boolean
  saveError: string | null
  saveSuccess: boolean
  onSave: () => void
}) {
  return (
    <div className="mt-6 flex flex-col gap-3 border-t border-surface-border pt-5 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-h-5 text-sm">
        {saveError && <span className="text-red-600">{saveError}</span>}
        {saveSuccess && <span className="text-green-600">Сохранено</span>}
      </div>
      <button
        type="button"
        onClick={onSave}
        disabled={isSaving}
        className={cn(
          'inline-flex min-h-11 items-center justify-center rounded-lg px-5 text-sm font-medium',
          'bg-accent text-white hover:bg-accent/90 transition-colors',
          'disabled:cursor-not-allowed disabled:opacity-50',
        )}
      >
        {isSaving ? 'Сохранение...' : 'Сохранить изменения'}
      </button>
    </div>
  )
}

function buildModuleConfigs(configs: ModuleConfigs | undefined, draft: FeedbackDraft): ModuleConfigs {
  const normalized = normalizeModuleConfigs(configs)
  return {
    ...normalized,
    feedback: {
      prompt_message: draft.prompt_message,
      success_message: draft.success_message,
    },
  }
}

export default function BotFeedbackSettingsPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)
  const { data: bot, isLoading, isError, mutate } = useBotQuery(Number.isNaN(id) ? 0 : id)
  const [draft, setDraft] = useState<FeedbackDraft | null>(null)

  useEffect(() => {
    if (!bot) return
    const configs = normalizeModuleConfigs(bot.settings.module_configs)
    setDraft({
      prompt_message: configs.feedback?.prompt_message ?? '',
      success_message: configs.feedback?.success_message ?? '',
    })
  }, [bot])

  const { isSaving, saveError, saveSuccess, save } = useSaveAction(async () => {
    if (!bot || !draft) return
    await botsApi.updateSettings(id, {
      module_configs: buildModuleConfigs(bot.settings.module_configs, draft),
    })
    mutate()
  })

  if (isLoading || !draft) {
    return (
      <div className="max-w-4xl">
        <div className="mb-6 h-4 w-32 shimmer rounded" />
        <CardSkeleton />
      </div>
    )
  }

  if (Number.isNaN(id) || isError || !bot) {
    return (
      <div className="max-w-4xl">
        <Link
          to="/dashboard/bots"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          Назад к ботам
        </Link>
        <ErrorState title="Не удалось загрузить настройки" message="Проверьте подключение и попробуйте снова." />
      </div>
    )
  }

  return (
    <div className="max-w-4xl rounded-2xl border border-surface-border bg-white p-6 shadow-sm">
      <div className="mb-5">
        <Link
          to={`/dashboard/bots/${id}?tab=modules`}
          className="mb-4 inline-flex min-h-11 items-center gap-1.5 rounded-lg text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          К модулям
        </Link>
        <h1 className="text-lg font-semibold text-neutral-900">
          <span className="mb-0.5 block font-mono text-[10px] font-normal uppercase tracking-widest text-neutral-400">
            Настройки модуля
          </span>
          Связаться
        </h1>
        <p className="mt-1 max-w-2xl text-sm text-neutral-500">
          Текст запроса к гостю и сообщение после отправки обратной связи.
        </p>
      </div>

      <div className="space-y-4">
        <label className="space-y-2">
          <span className="text-sm font-medium text-neutral-700">Первое сообщение</span>
          <textarea
            rows={3}
            value={draft.prompt_message}
            onChange={(event) => setDraft((current) => (current ? { ...current, prompt_message: event.target.value } : current))}
            className={inputClassName}
            placeholder="Напишите ваш вопрос:"
          />
        </label>

        <label className="space-y-2">
          <span className="text-sm font-medium text-neutral-700">Сообщение после отправки</span>
          <textarea
            rows={3}
            value={draft.success_message}
            onChange={(event) => setDraft((current) => (current ? { ...current, success_message: event.target.value } : current))}
            className={inputClassName}
            placeholder="Ваше сообщение отправлено."
          />
        </label>
      </div>

      <SaveFooter isSaving={isSaving} saveError={saveError} saveSuccess={saveSuccess} onSave={save} />
    </div>
  )
}
