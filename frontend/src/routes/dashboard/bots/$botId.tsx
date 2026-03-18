import { Link, useParams } from 'react-router-dom'
import { useState, useEffect } from 'react'
import { cn } from '@/lib/utils'
import { useBotQuery } from '@/features/bots/queries'
import { botsApi } from '@/features/bots/api'
import { ArrowLeft, Plus, Trash2 } from 'lucide-react'
import type { BotButton, BotSettings, FormField } from '@/features/bots/types'

const statusConfig = {
  active: { label: 'Активен', className: 'bg-green-100 text-green-700' },
  inactive: { label: 'Неактивен', className: 'bg-neutral-100 text-neutral-500' },
  error: { label: 'Ошибка', className: 'bg-red-100 text-red-700' },
} as const

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

const inputClassName = cn(
  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
  'text-sm placeholder:text-neutral-400',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
  'disabled:opacity-50 disabled:cursor-not-allowed',
)

export default function BotDetailPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)
  const { data: bot, isLoading, isError } = useBotQuery(id)

  const [settings, setSettings] = useState<BotSettings | null>(null)
  const [isSaving, setIsSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)
  const [saveSuccess, setSaveSuccess] = useState(false)

  useEffect(() => {
    if (bot?.settings) {
      setSettings({
        welcome_message: bot.settings.welcome_message ?? '',
        modules: bot.settings.modules ?? [],
        buttons: bot.settings.buttons ?? [],
        registration_form: bot.settings.registration_form ?? [],
      })
    }
  }, [bot])

  const handleSave = async () => {
    if (!settings) return
    setIsSaving(true)
    setSaveError(null)
    setSaveSuccess(false)

    try {
      await botsApi.updateSettings(id, settings)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch {
      setSaveError('Не удалось сохранить настройки. Попробуйте снова.')
    } finally {
      setIsSaving(false)
    }
  }

  if (isLoading) {
    return (
      <div className="max-w-3xl">
        <div className="flex items-center justify-center py-20">
          <div className="w-6 h-6 border-2 border-neutral-300 border-t-neutral-900 rounded-full animate-spin" />
        </div>
      </div>
    )
  }

  if (isError || !bot) {
    return (
      <div className="max-w-3xl">
        <Link
          to="/dashboard/bots"
          className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
        >
          <ArrowLeft className="w-4 h-4" />
          Назад к списку
        </Link>
        <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
          <p className="text-sm text-red-600">
            Не удалось загрузить информацию о боте.
          </p>
        </div>
      </div>
    )
  }

  const status = statusConfig[bot.status]

  // --- Button helpers ---
  const addButton = () => {
    if (!settings) return
    setSettings({
      ...settings,
      buttons: [...settings.buttons, { label: '', type: 'url', value: '' }],
    })
  }

  const updateButton = (index: number, field: keyof BotButton, value: string) => {
    if (!settings) return
    const updated = settings.buttons.map((btn, i) =>
      i === index ? { ...btn, [field]: value } : btn,
    )
    setSettings({ ...settings, buttons: updated })
  }

  const removeButton = (index: number) => {
    if (!settings) return
    setSettings({
      ...settings,
      buttons: settings.buttons.filter((_, i) => i !== index),
    })
  }

  // --- Form field helpers ---
  const addFormField = () => {
    if (!settings) return
    setSettings({
      ...settings,
      registration_form: [
        ...settings.registration_form,
        { name: '', label: '', type: 'text', required: false },
      ],
    })
  }

  const updateFormField = (index: number, field: keyof FormField, value: string | boolean) => {
    if (!settings) return
    const updated = settings.registration_form.map((f, i) =>
      i === index ? { ...f, [field]: value } : f,
    )
    setSettings({ ...settings, registration_form: updated })
  }

  const removeFormField = (index: number) => {
    if (!settings) return
    setSettings({
      ...settings,
      registration_form: settings.registration_form.filter((_, i) => i !== index),
    })
  }

  // --- Module helpers ---
  const availableModules = ['loyalty', 'mailings', 'promotions', 'referral', 'feedback']
  const moduleLabels: Record<string, string> = {
    loyalty: 'Лояльность',
    mailings: 'Рассылки',
    promotions: 'Акции',
    referral: 'Реферальная программа',
    feedback: 'Обратная связь',
  }

  const toggleModule = (module: string) => {
    if (!settings) return
    const modules = settings.modules.includes(module)
      ? settings.modules.filter((m) => m !== module)
      : [...settings.modules, module]
    setSettings({ ...settings, modules })
  }

  return (
    <div className="max-w-3xl">
      <Link
        to="/dashboard/bots"
        className="inline-flex items-center gap-1.5 text-sm text-neutral-500 hover:text-neutral-700 transition-colors mb-6"
      >
        <ArrowLeft className="w-4 h-4" />
        Назад к списку
      </Link>

      {/* Bot info header */}
      <div className="bg-white rounded-2xl shadow-sm border border-surface-border p-6 mb-6">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold text-neutral-900">{bot.name}</h1>
            <p className="text-neutral-400 mt-1">@{bot.username}</p>
          </div>
          <span
            className={cn(
              'text-xs font-medium px-2.5 py-1 rounded-full',
              status.className,
            )}
          >
            {status.label}
          </span>
        </div>
        <p className="text-sm text-neutral-400 mt-4">
          Создан {formatDate(bot.created_at)}
        </p>
      </div>

      {settings && (
        <div className="space-y-6">
          {/* Welcome message & modules */}
          <section className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
            <h2 className="text-lg font-semibold text-neutral-900 mb-5">
              Настройки бота
            </h2>

            <div className="space-y-5">
              <div>
                <label
                  htmlFor="welcome-message"
                  className="block text-sm font-medium text-neutral-700 mb-1.5"
                >
                  Приветственное сообщение
                </label>
                <textarea
                  id="welcome-message"
                  value={settings.welcome_message}
                  onChange={(e) =>
                    setSettings({ ...settings, welcome_message: e.target.value })
                  }
                  placeholder="Привет! Добро пожаловать в нашу программу лояльности..."
                  rows={4}
                  disabled={isSaving}
                  className={cn(inputClassName, 'resize-none')}
                />
              </div>

              <div>
                <p className="block text-sm font-medium text-neutral-700 mb-3">
                  Модули
                </p>
                <div className="flex flex-wrap gap-2">
                  {availableModules.map((module) => {
                    const isActive = settings.modules.includes(module)
                    return (
                      <button
                        key={module}
                        type="button"
                        onClick={() => toggleModule(module)}
                        disabled={isSaving}
                        className={cn(
                          'px-3 py-1.5 rounded-lg text-sm font-medium transition-colors',
                          'disabled:opacity-50 disabled:cursor-not-allowed',
                          isActive
                            ? 'bg-neutral-900 text-white'
                            : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200',
                        )}
                      >
                        {moduleLabels[module] ?? module}
                      </button>
                    )
                  })}
                </div>
              </div>
            </div>
          </section>

          {/* Menu buttons */}
          <section className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
            <div className="flex items-center justify-between mb-5">
              <h2 className="text-lg font-semibold text-neutral-900">
                Кнопки меню
              </h2>
              <button
                type="button"
                onClick={addButton}
                disabled={isSaving}
                className={cn(
                  'flex items-center gap-1.5 text-sm font-medium text-accent',
                  'hover:text-accent/80 transition-colors',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                )}
              >
                <Plus className="w-4 h-4" />
                Добавить
              </button>
            </div>

            {settings.buttons.length === 0 ? (
              <p className="text-sm text-neutral-400 text-center py-4">
                Нет кнопок меню
              </p>
            ) : (
              <div className="space-y-3">
                {settings.buttons.map((button, index) => (
                  <div
                    key={index}
                    className="flex items-start gap-3 p-3 rounded-lg bg-neutral-50"
                  >
                    <div className="flex-1 grid grid-cols-3 gap-3">
                      <input
                        type="text"
                        value={button.label}
                        onChange={(e) => updateButton(index, 'label', e.target.value)}
                        placeholder="Название"
                        disabled={isSaving}
                        className={inputClassName}
                        aria-label={`Название кнопки ${index + 1}`}
                      />
                      <select
                        value={button.type}
                        onChange={(e) => updateButton(index, 'type', e.target.value)}
                        disabled={isSaving}
                        className={inputClassName}
                        aria-label={`Тип кнопки ${index + 1}`}
                      >
                        <option value="url">Ссылка</option>
                        <option value="callback">Callback</option>
                        <option value="webapp">WebApp</option>
                      </select>
                      <input
                        type="text"
                        value={button.value}
                        onChange={(e) => updateButton(index, 'value', e.target.value)}
                        placeholder="Значение"
                        disabled={isSaving}
                        className={inputClassName}
                        aria-label={`Значение кнопки ${index + 1}`}
                      />
                    </div>
                    <button
                      type="button"
                      onClick={() => removeButton(index)}
                      disabled={isSaving}
                      className="p-2 rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                      aria-label={`Удалить кнопку ${index + 1}`}
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Registration form fields */}
          <section className="bg-white rounded-2xl shadow-sm border border-surface-border p-6">
            <div className="flex items-center justify-between mb-5">
              <h2 className="text-lg font-semibold text-neutral-900">
                Анкета регистрации
              </h2>
              <button
                type="button"
                onClick={addFormField}
                disabled={isSaving}
                className={cn(
                  'flex items-center gap-1.5 text-sm font-medium text-accent',
                  'hover:text-accent/80 transition-colors',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                )}
              >
                <Plus className="w-4 h-4" />
                Добавить
              </button>
            </div>

            {settings.registration_form.length === 0 ? (
              <p className="text-sm text-neutral-400 text-center py-4">
                Нет полей анкеты
              </p>
            ) : (
              <div className="space-y-3">
                {settings.registration_form.map((field, index) => (
                  <div
                    key={index}
                    className="flex items-start gap-3 p-3 rounded-lg bg-neutral-50"
                  >
                    <div className="flex-1 grid grid-cols-2 gap-3">
                      <input
                        type="text"
                        value={field.name}
                        onChange={(e) => updateFormField(index, 'name', e.target.value)}
                        placeholder="Идентификатор (name)"
                        disabled={isSaving}
                        className={inputClassName}
                        aria-label={`Идентификатор поля ${index + 1}`}
                      />
                      <input
                        type="text"
                        value={field.label}
                        onChange={(e) => updateFormField(index, 'label', e.target.value)}
                        placeholder="Название для пользователя"
                        disabled={isSaving}
                        className={inputClassName}
                        aria-label={`Название поля ${index + 1}`}
                      />
                      <select
                        value={field.type}
                        onChange={(e) => updateFormField(index, 'type', e.target.value)}
                        disabled={isSaving}
                        className={inputClassName}
                        aria-label={`Тип поля ${index + 1}`}
                      >
                        <option value="text">Текст</option>
                        <option value="email">Email</option>
                        <option value="phone">Телефон</option>
                        <option value="date">Дата</option>
                        <option value="select">Выбор</option>
                      </select>
                      <label className="flex items-center gap-2 px-4 py-2.5">
                        <input
                          type="checkbox"
                          checked={field.required}
                          onChange={(e) =>
                            updateFormField(index, 'required', e.target.checked)
                          }
                          disabled={isSaving}
                          className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
                        />
                        <span className="text-sm text-neutral-700">Обязательное</span>
                      </label>
                    </div>
                    <button
                      type="button"
                      onClick={() => removeFormField(index)}
                      disabled={isSaving}
                      className="p-2 rounded-lg text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-50"
                      aria-label={`Удалить поле ${index + 1}`}
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* Save section */}
          <div className="flex items-center justify-between">
            <div>
              {saveError && (
                <p className="text-sm text-red-600">{saveError}</p>
              )}
              {saveSuccess && (
                <p className="text-sm text-green-600">Настройки сохранены</p>
              )}
            </div>
            <button
              type="button"
              onClick={handleSave}
              disabled={isSaving}
              className={cn(
                'py-2.5 px-6 rounded-lg',
                'bg-neutral-900 text-white text-sm font-medium',
                'hover:bg-neutral-800 active:bg-neutral-950',
                'transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-neutral-900/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {isSaving ? 'Сохранение...' : 'Сохранить'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
