import { useState, useEffect, useCallback } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { botsApi } from '@/features/bots/api'
import type { FormField, CreateManagedBotRequest } from '@/features/bots/types'
import {
  ArrowLeft, ArrowRight, Bot, Check, ExternalLink,
  Loader2, Plus, Trash2, GripVertical,
} from 'lucide-react'

const MODULES = [
  { id: 'loyalty', label: 'Лояльность', desc: 'Бонусная программа и уровни' },
  { id: 'menu', label: 'Меню', desc: 'Каталог блюд и напитков' },
  { id: 'marketplace', label: 'Маркетплейс', desc: 'Товары и мерч' },
  { id: 'feedback', label: 'Отзывы', desc: 'Сбор обратной связи' },
  { id: 'booking', label: 'Бронирование', desc: 'Резерв столиков' },
]

const DEFAULT_FORM_FIELDS: FormField[] = [
  { name: 'first_name', label: 'Имя', type: 'text', required: true },
  { name: 'phone', label: 'Телефон', type: 'phone', required: true },
]

const inputClass = cn(
  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
  'text-sm placeholder:text-neutral-400',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
  'disabled:opacity-50 disabled:cursor-not-allowed',
)

export default function CreateBotPage() {
  const navigate = useNavigate()
  const [step, setStep] = useState(0)

  // Step 1 — basic info
  const [name, setName] = useState('')
  const [username, setUsername] = useState('')
  const [description, setDescription] = useState('')

  // Step 2 — welcome + regform
  const [welcomeMessage, setWelcomeMessage] = useState('')
  const [formFields, setFormFields] = useState<FormField[]>(DEFAULT_FORM_FIELDS)

  // Step 3 — modules
  const [modules, setModules] = useState<string[]>(['loyalty'])

  // Step 4 — creation state
  const [deepLink, setDeepLink] = useState('')
  const [botId, setBotId] = useState<number | null>(null)
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState('')
  const [botStatus, setBotStatus] = useState<string>('pending')

  // Fallback
  const [showFallback, setShowFallback] = useState(false)
  const [fallbackToken, setFallbackToken] = useState('')
  const [fallbackCreating, setFallbackCreating] = useState(false)

  // Validation
  const isStep1Valid = name.trim().length > 0 &&
    username.trim().length >= 5 &&
    username.toLowerCase().endsWith('bot')

  const usernameError = username.length > 0 && !username.toLowerCase().endsWith('bot')
    ? 'Username должен заканчиваться на "bot"'
    : username.length > 0 && username.length < 5
      ? 'Минимум 5 символов'
      : ''

  // Step 4: submit wizard + poll status
  const handleCreateManaged = useCallback(async () => {
    setCreating(true)
    setError('')
    try {
      const data: CreateManagedBotRequest = {
        name: name.trim(),
        username: username.trim(),
        description: description.trim(),
        welcome_message: welcomeMessage.trim() || undefined,
        registration_form: formFields,
        modules,
      }
      const res = await botsApi.createManaged(data)
      setDeepLink(res.deep_link)
      setBotId(res.bot_id)
      setBotStatus(res.status)
    } catch {
      setError('Ошибка создания бота. Попробуйте снова.')
    } finally {
      setCreating(false)
    }
  }, [name, username, description, welcomeMessage, formFields, modules])

  // Poll bot status after creation
  useEffect(() => {
    if (!botId || botStatus !== 'pending') return
    const interval = setInterval(async () => {
      try {
        const res = await botsApi.getBotStatus(botId)
        setBotStatus(res.status)
        if (res.status === 'active') {
          clearInterval(interval)
        }
      } catch { /* ignore */ }
    }, 3000)
    const timeout = setTimeout(() => clearInterval(interval), 5 * 60 * 1000)
    return () => { clearInterval(interval); clearTimeout(timeout) }
  }, [botId, botStatus])

  // Redirect on success
  useEffect(() => {
    if (botStatus === 'active' && botId) {
      const timer = setTimeout(() => navigate(`/dashboard/bots/${botId}`), 1500)
      return () => clearTimeout(timer)
    }
  }, [botStatus, botId, navigate])

  // Fallback: create with token
  const handleFallbackCreate = async () => {
    setFallbackCreating(true)
    setError('')
    try {
      const bot = await botsApi.create({ name: name.trim(), token: fallbackToken.trim() })
      navigate(`/dashboard/bots/${bot.id}`)
    } catch {
      setError('Не удалось создать бота. Проверьте токен.')
    } finally {
      setFallbackCreating(false)
    }
  }

  // Form field helpers
  const addFormField = () => {
    setFormFields([...formFields, { name: `field_${Date.now()}`, label: '', type: 'text', required: false }])
  }
  const removeFormField = (idx: number) => {
    setFormFields(formFields.filter((_, i) => i !== idx))
  }
  const updateFormField = (idx: number, patch: Partial<FormField>) => {
    setFormFields(formFields.map((f, i) => i === idx ? { ...f, ...patch } : f))
  }

  const toggleModule = (id: string) => {
    setModules(prev => prev.includes(id) ? prev.filter(m => m !== id) : [...prev, id])
  }

  const steps = ['Основное', 'Приветствие', 'Модули', 'Создание']

  return (
    <div className="max-w-2xl mx-auto animate-in">
      {/* Header */}
      <div className="flex items-center gap-3 mb-6">
        <Link
          to="/dashboard/bots"
          className="p-2 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <h1 className="font-serif text-2xl font-bold text-neutral-900">Создать бота</h1>
      </div>

      {/* Progress */}
      <div className="flex items-center gap-2 mb-8">
        {steps.map((label, i) => (
          <div key={label} className="flex-1">
            <div
              className={cn(
                'h-1.5 rounded-full transition-all duration-500',
                i < step ? 'bg-accent' : i === step ? 'bg-accent/50' : 'bg-neutral-200',
              )}
            />
            <p className={cn(
              'text-[10px] mt-1 text-center font-medium uppercase tracking-wider',
              i <= step ? 'text-neutral-600' : 'text-neutral-300',
            )}>
              {label}
            </p>
          </div>
        ))}
      </div>

      {/* Step content */}
      <div className="bg-white rounded-xl border border-surface-border p-6">
        {/* Step 1: Basic info */}
        {step === 0 && (
          <div className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                Имя бота <span className="text-red-400">*</span>
              </label>
              <input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Мой ресторан"
                maxLength={100}
                autoFocus
                className={inputClass}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                Username <span className="text-red-400">*</span>
              </label>
              <div className="relative">
                <span className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-400 text-sm">@</span>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value.replace(/[^a-zA-Z0-9_]/g, ''))}
                  placeholder="myrestaurant_bot"
                  maxLength={32}
                  className={cn(inputClass, 'pl-8')}
                />
              </div>
              {usernameError && <p className="mt-1 text-xs text-red-500">{usernameError}</p>}
              <p className="mt-1 text-xs text-neutral-400">5–32 символа, латиница + цифры + _, должен заканчиваться на «bot»</p>
            </div>
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                Описание
              </label>
              <textarea
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Описание бота (видно в профиле)"
                maxLength={512}
                rows={3}
                className={cn(inputClass, 'resize-none')}
              />
            </div>
            <div className="flex justify-end pt-2">
              <button
                onClick={() => setStep(1)}
                disabled={!isStep1Valid}
                className={cn(
                  'flex items-center gap-2 py-2.5 px-5 rounded-lg',
                  'bg-accent text-white text-sm font-medium',
                  'hover:bg-accent-hover transition-colors',
                  'disabled:opacity-40 disabled:cursor-not-allowed',
                )}
              >
                Далее <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        )}

        {/* Step 2: Welcome + Registration */}
        {step === 1 && (
          <div className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-neutral-700 mb-1.5">
                Welcome-сообщение
              </label>
              <textarea
                value={welcomeMessage}
                onChange={(e) => setWelcomeMessage(e.target.value)}
                placeholder="Сообщение при первом запуске бота клиентом"
                rows={4}
                className={cn(inputClass, 'resize-none')}
              />
            </div>

            <div>
              <div className="flex items-center justify-between mb-3">
                <label className="text-sm font-medium text-neutral-700">Поля регистрации</label>
                <button
                  onClick={addFormField}
                  className="flex items-center gap-1 text-xs text-accent hover:text-accent-hover transition-colors"
                >
                  <Plus className="w-3.5 h-3.5" /> Добавить
                </button>
              </div>
              <div className="space-y-2">
                {formFields.map((field, idx) => (
                  <div key={idx} className="flex items-center gap-2 p-3 rounded-lg bg-neutral-50 border border-neutral-100">
                    <GripVertical className="w-4 h-4 text-neutral-300 shrink-0" />
                    <input
                      type="text"
                      value={field.label}
                      onChange={(e) => updateFormField(idx, { label: e.target.value, name: e.target.value.toLowerCase().replace(/\s+/g, '_') })}
                      placeholder="Название поля"
                      className="flex-1 px-3 py-1.5 text-sm rounded border border-neutral-200 focus:outline-none focus:border-accent"
                    />
                    <select
                      value={field.type}
                      onChange={(e) => updateFormField(idx, { type: e.target.value })}
                      className="px-2 py-1.5 text-xs rounded border border-neutral-200"
                    >
                      <option value="text">Текст</option>
                      <option value="phone">Телефон</option>
                      <option value="date">Дата</option>
                      <option value="email">Email</option>
                    </select>
                    <label className="flex items-center gap-1 text-xs text-neutral-500 whitespace-nowrap">
                      <input
                        type="checkbox"
                        checked={field.required}
                        onChange={(e) => updateFormField(idx, { required: e.target.checked })}
                        className="rounded"
                      />
                      Обяз.
                    </label>
                    <button
                      onClick={() => removeFormField(idx)}
                      className="p-1 text-neutral-400 hover:text-red-500 transition-colors"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </button>
                  </div>
                ))}
              </div>
            </div>

            <div className="flex justify-between pt-2">
              <button
                onClick={() => setStep(0)}
                className="flex items-center gap-2 py-2.5 px-4 rounded-lg border border-surface-border text-sm text-neutral-600 hover:bg-neutral-50 transition-colors"
              >
                <ArrowLeft className="w-4 h-4" /> Назад
              </button>
              <div className="flex gap-2">
                <button
                  onClick={() => setStep(2)}
                  className="py-2.5 px-4 rounded-lg text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
                >
                  Пропустить
                </button>
                <button
                  onClick={() => setStep(2)}
                  className="flex items-center gap-2 py-2.5 px-5 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-hover transition-colors"
                >
                  Далее <ArrowRight className="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
        )}

        {/* Step 3: Modules */}
        {step === 2 && (
          <div className="space-y-4">
            <p className="text-sm text-neutral-500 mb-4">Выберите модули для бота. Можно изменить позже.</p>
            {MODULES.map((mod) => (
              <label
                key={mod.id}
                className={cn(
                  'flex items-center gap-4 p-4 rounded-lg border cursor-pointer transition-all',
                  modules.includes(mod.id)
                    ? 'border-accent bg-accent/5'
                    : 'border-neutral-200 hover:border-neutral-300',
                )}
              >
                <input
                  type="checkbox"
                  checked={modules.includes(mod.id)}
                  onChange={() => toggleModule(mod.id)}
                  className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
                />
                <div>
                  <p className="text-sm font-medium text-neutral-900">{mod.label}</p>
                  <p className="text-xs text-neutral-400">{mod.desc}</p>
                </div>
              </label>
            ))}

            <div className="flex justify-between pt-4">
              <button
                onClick={() => setStep(1)}
                className="flex items-center gap-2 py-2.5 px-4 rounded-lg border border-surface-border text-sm text-neutral-600 hover:bg-neutral-50 transition-colors"
              >
                <ArrowLeft className="w-4 h-4" /> Назад
              </button>
              <button
                onClick={() => { setStep(3); handleCreateManaged() }}
                className="flex items-center gap-2 py-2.5 px-5 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-hover transition-colors"
              >
                <Bot className="w-4 h-4" /> Создать бота
              </button>
            </div>
          </div>
        )}

        {/* Step 4: Creation */}
        {step === 3 && (
          <div className="text-center py-4">
            {creating ? (
              <div className="space-y-4">
                <Loader2 className="w-10 h-10 text-accent animate-spin mx-auto" />
                <p className="text-neutral-600">Создаём бота...</p>
              </div>
            ) : error ? (
              <div className="space-y-4">
                <p className="text-red-600 text-sm">{error}</p>
                <button
                  onClick={handleCreateManaged}
                  className="py-2.5 px-5 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-hover transition-colors"
                >
                  Повторить
                </button>
              </div>
            ) : botStatus === 'active' ? (
              <div className="space-y-4">
                <div className="w-14 h-14 rounded-full bg-green-100 flex items-center justify-center mx-auto">
                  <Check className="w-7 h-7 text-green-600" />
                </div>
                <div>
                  <p className="text-lg font-semibold text-neutral-900">Бот создан!</p>
                  <p className="text-sm text-neutral-500 mt-1">Переходим на страницу бота...</p>
                </div>
              </div>
            ) : (
              <div className="space-y-6">
                <div>
                  <Bot className="w-12 h-12 text-accent/60 mx-auto mb-4" />
                  <p className="text-neutral-900 font-medium mb-2">
                    Перейдите в Telegram для подтверждения
                  </p>
                  <p className="text-sm text-neutral-500">
                    Нажмите кнопку ниже, подтвердите создание бота в Telegram и вернитесь сюда.
                  </p>
                </div>

                <a
                  href={deepLink}
                  target="_blank"
                  rel="noopener noreferrer"
                  className={cn(
                    'inline-flex items-center gap-2 py-3 px-6 rounded-lg',
                    'bg-[#2AABEE] text-white text-sm font-medium',
                    'hover:bg-[#229ED9] transition-colors',
                  )}
                >
                  <ExternalLink className="w-4 h-4" />
                  Создать в Telegram
                </a>

                <div className="flex items-center gap-2 justify-center text-neutral-400">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span className="text-xs">Ожидаем подтверждения...</span>
                </div>
              </div>
            )}

            {/* Fallback */}
            {!creating && botStatus === 'pending' && (
              <div className="mt-8 pt-6 border-t border-neutral-100">
                {!showFallback ? (
                  <button
                    onClick={() => setShowFallback(true)}
                    className="text-xs text-neutral-400 hover:text-neutral-600 transition-colors"
                  >
                    У меня уже есть бот
                  </button>
                ) : (
                  <div className="max-w-sm mx-auto space-y-3">
                    <input
                      type="text"
                      value={fallbackToken}
                      onChange={(e) => setFallbackToken(e.target.value)}
                      placeholder="Вставьте токен от @BotFather"
                      className={inputClass}
                    />
                    <button
                      onClick={handleFallbackCreate}
                      disabled={!fallbackToken.trim() || fallbackCreating}
                      className={cn(
                        'w-full py-2.5 px-4 rounded-lg',
                        'bg-neutral-900 text-white text-sm font-medium',
                        'hover:bg-neutral-800 transition-colors',
                        'disabled:opacity-40 disabled:cursor-not-allowed',
                      )}
                    >
                      {fallbackCreating ? 'Создание...' : 'Подключить бота'}
                    </button>
                  </div>
                )}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  )
}
