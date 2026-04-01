import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Check, ChevronLeft, Save, Loader2, Coffee, UtensilsCrossed, Store, Wine } from 'lucide-react'
import { useRFMTemplatesQuery, useRFMActiveTemplateQuery, useRFMSetTemplateMutation } from '@/features/rfm/queries'
import type { RFMTemplate } from '@/features/rfm/types'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import type { LucideIcon } from 'lucide-react'

const TEMPLATE_ICONS: Record<string, { icon: LucideIcon; color: string }> = {
  coffeegng: { icon: Coffee, color: '#f59e0b' },
  qsr: { icon: Store, color: '#3b82f6' },
  tsr: { icon: UtensilsCrossed, color: '#8b5cf6' },
  bar: { icon: Wine, color: '#ef4444' },
}

export default function RFMTemplatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const isCustomMode = searchParams.get('mode') === 'custom'

  const { data: templates, isLoading, isError, mutate: retryTemplates } = useRFMTemplatesQuery()
  const { data: activeTemplate } = useRFMActiveTemplateQuery()
  const setTemplateMutation = useRFMSetTemplateMutation()

  const [showCustom, setShowCustom] = useState(isCustomMode)
  const [selectedKey, setSelectedKey] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)

  // Custom template form state
  const [customName, setCustomName] = useState('')
  const [rThresholds, setRThresholds] = useState<[number, number, number, number]>([7, 14, 30, 60])
  const [fThresholds, setFThresholds] = useState<[number, number, number, number]>([8, 5, 3, 2])
  const [validationErrors, setValidationErrors] = useState<string[]>([])

  // Initialize selected from active template
  useEffect(() => {
    if (activeTemplate?.active_template_key) {
      setSelectedKey(activeTemplate.active_template_key)
    }
    if (activeTemplate?.active_template_type === 'custom' && activeTemplate.template) {
      setCustomName(activeTemplate.template.name)
      setRThresholds(activeTemplate.template.r_thresholds)
      setFThresholds(activeTemplate.template.f_thresholds)
    }
  }, [activeTemplate])

  const activeKey = activeTemplate?.active_template_key ?? null
  const hasChanged = selectedKey !== null && selectedKey !== activeKey

  async function handleSaveStandard() {
    if (!selectedKey || !hasChanged) return
    setSaveError(null)
    try {
      await setTemplateMutation.mutateAsync({ template_type: 'standard', template_key: selectedKey })
      navigate('/dashboard/rfm')
    } catch {
      setSaveError('Не удалось сохранить шаблон. Попробуйте ещё раз.')
    }
  }

  function validateCustom(): boolean {
    const errors: string[] = []
    for (let i = 0; i < 4; i++) {
      if (rThresholds[i] < 0) errors.push('R: значение должно быть ≥ 0')
      if (i > 0 && rThresholds[i] <= rThresholds[i - 1]) {
        errors.push('Recency: пороги должны строго возрастать')
        break
      }
    }
    for (let i = 0; i < 4; i++) {
      if (fThresholds[i] < 1) errors.push('F: значение должно быть ≥ 1')
      if (i > 0 && fThresholds[i] >= fThresholds[i - 1]) {
        errors.push('Frequency: пороги должны строго убывать')
        break
      }
    }
    setValidationErrors(errors)
    return errors.length === 0
  }

  async function handleSaveCustom() {
    if (!validateCustom()) return
    setSaveError(null)
    try {
      await setTemplateMutation.mutateAsync({
        template_type: 'custom',
        custom_name: customName || 'Мой шаблон',
        r_thresholds: rThresholds,
        f_thresholds: fThresholds,
      })
      navigate('/dashboard/rfm')
    } catch {
      setSaveError('Не удалось сохранить шаблон. Попробуйте ещё раз.')
    }
  }

  if (isLoading) {
    return (
      <div>
        <div className="shimmer h-8 w-48 rounded mb-6" />
        <div className="grid grid-cols-2 gap-3">
          {[0, 1, 2, 3].map((i) => <CardSkeleton key={i} />)}
        </div>
      </div>
    )
  }

  if (isError) {
    return <ErrorState title="Не удалось загрузить шаблоны" onRetry={() => retryTemplates()} />
  }

  return (
    <div>
      {/* Back + header */}
      <div className="mb-6 animate-in">
        <button
          type="button"
          onClick={() => navigate('/dashboard/rfm')}
          className="flex items-center gap-1 text-sm text-neutral-400 hover:text-neutral-600 transition-colors mb-3"
        >
          <ChevronLeft className="w-4 h-4" />
          RFM-сегменты
        </button>
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          {showCustom ? 'Ручная настройка' : 'Выбор шаблона'}
        </h1>
        <p className="font-mono text-xs text-neutral-300 uppercase tracking-wider mt-1">
          {showCustom ? 'Задайте пороги вручную' : 'Выберите подходящий тип заведения'}
        </p>
      </div>

      {saveError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded animate-in">
          {saveError}
        </div>
      )}

      {/* Tab switch */}
      <div className="flex gap-1 p-1 bg-neutral-100 rounded w-fit mb-6">
        <button
          type="button"
          onClick={() => setShowCustom(false)}
          className={cn(
            'px-4 py-2 rounded text-sm font-medium transition-all duration-150',
            !showCustom
              ? 'bg-white text-neutral-900 shadow-sm'
              : 'text-neutral-500 hover:text-neutral-700',
          )}
        >
          Стандартные
        </button>
        <button
          type="button"
          onClick={() => setShowCustom(true)}
          className={cn(
            'px-4 py-2 rounded text-sm font-medium transition-all duration-150',
            showCustom
              ? 'bg-white text-neutral-900 shadow-sm'
              : 'text-neutral-500 hover:text-neutral-700',
          )}
        >
          Вручную
        </button>
      </div>

      {/* Standard templates */}
      {!showCustom && templates && (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 animate-in">
            {templates.map((t, i) => {
              const isSelected = selectedKey === t.key
              const iconData = TEMPLATE_ICONS[t.key]
              const Icon = iconData?.icon

              return (
                <button
                  key={t.key}
                  type="button"
                  onClick={() => setSelectedKey(t.key)}
                  className={cn(
                    'relative text-left bg-white rounded border border-neutral-900 p-6',
                    'transition-all duration-200',
                    'animate-in',
                    `animate-in-delay-${i + 1}`,
                    isSelected ? 'bg-neutral-50' : 'hover:bg-neutral-50',
                  )}
                >
                  {isSelected && (
                    <div className="absolute top-4 right-4 w-6 h-6 rounded-full bg-accent flex items-center justify-center">
                      <Check className="w-3.5 h-3.5 text-white" />
                    </div>
                  )}

                  <div className="flex items-center gap-3 mb-3">
                    {Icon && (
                      <div className="w-10 h-10 rounded bg-neutral-100 flex items-center justify-center shrink-0">
                        <Icon className="w-5 h-5" style={{ color: iconData.color }} />
                      </div>
                    )}
                    <h3 className="font-semibold text-neutral-900">{t.name}</h3>
                  </div>
                  <p className="text-sm text-neutral-500 leading-relaxed">{t.description}</p>

                  {activeKey === t.key && !hasChanged && (
                    <p className="mt-3 text-xs text-neutral-400">Текущий шаблон</p>
                  )}
                </button>
              )
            })}
          </div>

          {/* Save button */}
          <div className="mt-6">
            <button
              type="button"
              onClick={handleSaveStandard}
              disabled={!hasChanged || setTemplateMutation.isPending}
              className={cn(
                'flex items-center gap-2 py-2.5 px-6 rounded text-sm font-medium',
                'transition-all duration-150',
                hasChanged
                  ? 'bg-accent text-white hover:bg-accent-hover active:bg-accent/80'
                  : 'bg-neutral-200 text-neutral-400 cursor-not-allowed',
                'disabled:opacity-70',
              )}
            >
              {setTemplateMutation.isPending ? (
                <Loader2 className="w-4 h-4 animate-spin" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              Сохранить
            </button>
          </div>
        </>
      )}

      {/* Custom template editor */}
      {showCustom && (
        <div className="animate-in space-y-6">
          <div>
            <label className="block text-sm font-medium text-neutral-700 mb-1.5">
              Название шаблона
            </label>
            <input
              type="text"
              value={customName}
              onChange={(e) => setCustomName(e.target.value)}
              placeholder="Мой шаблон"
              className={cn(
                'w-full max-w-sm px-4 py-2.5 rounded border border-neutral-200 text-sm',
                'focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/20',
                'placeholder:text-neutral-400',
              )}
            />
          </div>

          <div className="bg-white rounded border border-neutral-900 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-200">
              <h3 className="text-sm font-semibold text-neutral-700">
                Recency — давность последнего визита (дни)
              </h3>
              <p className="text-xs text-neutral-400 mt-0.5">Пороги должны строго возрастать</p>
            </div>
            <div className="px-6 py-4 grid grid-cols-4 gap-3">
              {['R5 (лучший)', 'R4', 'R3', 'R2'].map((label, i) => (
                <div key={label}>
                  <label className="block text-xs text-neutral-500 mb-1">{label}</label>
                  <input
                    type="number"
                    min={0}
                    value={rThresholds[i]}
                    onChange={(e) => {
                      const val = [...rThresholds] as [number, number, number, number]
                      val[i] = parseInt(e.target.value) || 0
                      setRThresholds(val)
                      setValidationErrors([])
                    }}
                    className={cn(
                      'w-full px-3 py-2 rounded border border-neutral-200 text-sm text-center font-mono',
                      'focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/20',
                    )}
                  />
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded border border-neutral-900 overflow-hidden">
            <div className="px-6 py-4 border-b border-neutral-200">
              <h3 className="text-sm font-semibold text-neutral-700">
                Frequency — частота визитов (количество)
              </h3>
              <p className="text-xs text-neutral-400 mt-0.5">Пороги должны строго убывать</p>
            </div>
            <div className="px-6 py-4 grid grid-cols-4 gap-3">
              {['F5 (лучший)', 'F4', 'F3', 'F2'].map((label, i) => (
                <div key={label}>
                  <label className="block text-xs text-neutral-500 mb-1">{label}</label>
                  <input
                    type="number"
                    min={1}
                    value={fThresholds[i]}
                    onChange={(e) => {
                      const val = [...fThresholds] as [number, number, number, number]
                      val[i] = parseInt(e.target.value) || 0
                      setFThresholds(val)
                      setValidationErrors([])
                    }}
                    className={cn(
                      'w-full px-3 py-2 rounded border border-neutral-200 text-sm text-center font-mono',
                      'focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/20',
                    )}
                  />
                </div>
              ))}
            </div>
          </div>

          {validationErrors.length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded p-4">
              <ul className="text-sm text-red-700 space-y-1">
                {validationErrors.map((err, i) => (
                  <li key={i}>• {err}</li>
                ))}
              </ul>
            </div>
          )}

          <button
            type="button"
            onClick={handleSaveCustom}
            disabled={setTemplateMutation.isPending}
            className={cn(
              'flex items-center gap-2 py-2.5 px-6 rounded',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent-hover active:bg-accent/80',
              'transition-all duration-150',
              'disabled:opacity-50',
            )}
          >
            {setTemplateMutation.isPending ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            Сохранить шаблон
          </button>
        </div>
      )}
    </div>
  )
}
