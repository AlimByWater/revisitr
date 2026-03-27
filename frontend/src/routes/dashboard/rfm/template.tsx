import { useState, useEffect } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { Check, ChevronLeft, Save, Loader2 } from 'lucide-react'
import { useRFMTemplatesQuery, useRFMActiveTemplateQuery, useRFMSetTemplateMutation } from '@/features/rfm/queries'
import type { RFMTemplate } from '@/features/rfm/types'
import { ErrorState } from '@/components/common/ErrorState'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'

export default function RFMTemplatePage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const isCustomMode = searchParams.get('mode') === 'custom'

  const { data: templates, isLoading, isError, mutate: retryTemplates } = useRFMTemplatesQuery()
  const { data: activeTemplate } = useRFMActiveTemplateQuery()
  const setTemplateMutation = useRFMSetTemplateMutation()

  const [showCustom, setShowCustom] = useState(isCustomMode)
  const [confirmKey, setConfirmKey] = useState<string | null>(null)
  const [saveError, setSaveError] = useState<string | null>(null)

  // Custom template form state
  const [customName, setCustomName] = useState('')
  const [rThresholds, setRThresholds] = useState<[number, number, number, number]>([7, 14, 30, 60])
  const [fThresholds, setFThresholds] = useState<[number, number, number, number]>([8, 5, 3, 2])
  const [validationErrors, setValidationErrors] = useState<string[]>([])

  // Initialize custom form from active template if it's custom
  useEffect(() => {
    if (activeTemplate?.active_template_type === 'custom' && activeTemplate.template) {
      setCustomName(activeTemplate.template.name)
      setRThresholds(activeTemplate.template.r_thresholds)
      setFThresholds(activeTemplate.template.f_thresholds)
    }
  }, [activeTemplate])

  async function handleSelectStandard(template: RFMTemplate) {
    if (confirmKey === template.key) {
      setSaveError(null)
      try {
        await setTemplateMutation.mutateAsync({ template_type: 'standard', template_key: template.key })
        setConfirmKey(null)
        navigate('/dashboard/rfm')
      } catch {
        setSaveError('Не удалось сохранить шаблон. Попробуйте ещё раз.')
        setConfirmKey(null)
      }
    } else {
      setConfirmKey(template.key)
    }
  }

  function validateCustom(): boolean {
    const errors: string[] = []
    // R: strictly ascending, all >= 0
    for (let i = 0; i < 4; i++) {
      if (rThresholds[i] < 0) errors.push(`R${5 - i}: значение должно быть ≥ 0`)
      if (i > 0 && rThresholds[i] <= rThresholds[i - 1]) {
        errors.push('Recency: пороги должны строго возрастать')
        break
      }
    }
    // F: strictly descending, all >= 1
    for (let i = 0; i < 4; i++) {
      if (fThresholds[i] < 1) errors.push(`F${5 - i}: значение должно быть ≥ 1`)
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
      <div className="max-w-3xl">
        <div className="shimmer h-8 w-48 rounded mb-6" />
        <div className="grid grid-cols-2 gap-3">
          {[0, 1, 2, 3].map((i) => <CardSkeleton key={i} />)}
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="max-w-3xl">
        <ErrorState title="Не удалось загрузить шаблоны" onRetry={() => retryTemplates()} />
      </div>
    )
  }

  return (
    <div className="max-w-3xl">
      {/* Back + header */}
      <div className="mb-6 animate-in">
        <button
          type="button"
          onClick={() => navigate('/dashboard/rfm')}
          className="flex items-center gap-1 text-sm text-neutral-400 hover:text-neutral-600 transition-colors mb-3"
        >
          <ChevronLeft className="w-4 h-4" />
          RFM-сегментация
        </button>
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          {showCustom ? 'Ручная настройка' : 'Выбор шаблона'}
        </h1>
      </div>

      {/* Save error */}
      {saveError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded-lg animate-in">
          {saveError}
        </div>
      )}

      {/* Tab switch */}
      <div className="flex gap-2 mb-6">
        <button
          type="button"
          onClick={() => setShowCustom(false)}
          className={cn(
            'px-4 py-2 rounded-lg text-sm font-medium transition-all',
            !showCustom
              ? 'bg-neutral-900 text-white'
              : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200',
          )}
        >
          Стандартные
        </button>
        <button
          type="button"
          onClick={() => setShowCustom(true)}
          className={cn(
            'px-4 py-2 rounded-lg text-sm font-medium transition-all',
            showCustom
              ? 'bg-neutral-900 text-white'
              : 'bg-neutral-100 text-neutral-600 hover:bg-neutral-200',
          )}
        >
          Вручную
        </button>
      </div>

      {/* Standard templates */}
      {!showCustom && templates && (
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 animate-in">
          {templates.map((t, i) => {
            const isActive =
              activeTemplate?.active_template_type === 'standard' &&
              activeTemplate?.active_template_key === t.key
            const isConfirming = confirmKey === t.key

            return (
              <button
                key={t.key}
                type="button"
                onClick={() => handleSelectStandard(t)}
                className={cn(
                  'relative text-left bg-white rounded-2xl border p-5',
                  'transition-all duration-200',
                  'animate-in',
                  `animate-in-delay-${i + 1}`,
                  isActive
                    ? 'border-accent/50 ring-1 ring-accent/20'
                    : isConfirming
                      ? 'border-accent/30 bg-accent/5'
                      : 'border-surface-border hover:border-neutral-300 hover:shadow-md',
                )}
              >
                {isActive && (
                  <div className="absolute top-3 right-3 w-6 h-6 rounded-full bg-accent flex items-center justify-center">
                    <Check className="w-3.5 h-3.5 text-white" />
                  </div>
                )}

                <h3 className="font-semibold text-neutral-900 mb-1">{t.name}</h3>
                <p className="text-sm text-neutral-500 mb-4">{t.description}</p>

                <div className="space-y-1.5 text-xs text-neutral-400">
                  <div className="flex items-center gap-2">
                    <span className="w-16 font-medium text-neutral-500">Recency</span>
                    <span className="font-mono tabular-nums">{t.r_thresholds.join(' / ')} дней</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="w-16 font-medium text-neutral-500">Frequency</span>
                    <span className="font-mono tabular-nums">{t.f_thresholds.join(' / ')} визитов</span>
                  </div>
                </div>

                {isConfirming && !isActive && (
                  <p className="mt-3 text-xs text-accent font-medium">
                    Нажмите ещё раз для подтверждения
                  </p>
                )}
              </button>
            )
          })}
        </div>
      )}

      {/* Custom template editor */}
      {showCustom && (
        <div className="animate-in space-y-6">
          {/* Template name */}
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
                'w-full px-4 py-2.5 rounded-lg border border-neutral-200 text-sm',
                'focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/20',
                'placeholder:text-neutral-400',
              )}
            />
          </div>

          {/* Thresholds table */}
          <div className="bg-white rounded-2xl border border-surface-border overflow-hidden">
            <div className="px-6 py-3 border-b border-surface-border">
              <h3 className="text-sm font-semibold text-neutral-700">
                Recency — давность последнего визита (дни)
              </h3>
              <p className="text-xs text-neutral-400 mt-0.5">
                Пороги должны строго возрастать (R5 ≤ R4 ≤ R3 ≤ R2)
              </p>
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
                      'w-full px-3 py-2 rounded-lg border border-neutral-200 text-sm text-center font-mono',
                      'focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/20',
                    )}
                  />
                </div>
              ))}
            </div>
          </div>

          <div className="bg-white rounded-2xl border border-surface-border overflow-hidden">
            <div className="px-6 py-3 border-b border-surface-border">
              <h3 className="text-sm font-semibold text-neutral-700">
                Frequency — частота визитов (количество)
              </h3>
              <p className="text-xs text-neutral-400 mt-0.5">
                Пороги должны строго убывать (F5 ≥ F4 ≥ F3 ≥ F2)
              </p>
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
                      'w-full px-3 py-2 rounded-lg border border-neutral-200 text-sm text-center font-mono',
                      'focus:outline-none focus:border-accent focus:ring-1 focus:ring-accent/20',
                    )}
                  />
                </div>
              ))}
            </div>
          </div>

          {/* Validation errors */}
          {validationErrors.length > 0 && (
            <div className="bg-red-50 border border-red-200 rounded-xl p-4">
              <ul className="text-sm text-red-700 space-y-1">
                {validationErrors.map((err, i) => (
                  <li key={i}>• {err}</li>
                ))}
              </ul>
            </div>
          )}

          {/* Save button */}
          <button
            type="button"
            onClick={handleSaveCustom}
            disabled={setTemplateMutation.isPending}
            className={cn(
              'flex items-center gap-2 py-2.5 px-6 rounded-lg',
              'bg-accent text-white text-sm font-medium',
              'hover:bg-accent-hover active:bg-accent/80',
              'transition-all duration-150',
              'shadow-sm shadow-accent/20',
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
