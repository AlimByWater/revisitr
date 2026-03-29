import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useCreatePromotionMutation } from '@/features/promotions/queries'
import { usePreviewAudienceMutation } from '@/features/campaigns/queries'
import { ClientFilterBuilder } from '@/components/filters/ClientFilterBuilder'
import { ArrowLeft, Plus, X, Check } from 'lucide-react'
import type { SegmentFilter } from '@/features/segments/types'
import type {
  CreatePromotionRequest,
  PromotionTrigger,
  PromotionAction,
} from '@/features/promotions/types'

// ── Constants ────────────────────────────────────────────────────────────────

const STEPS = [
  { num: 1, label: 'Основные параметры' },
  { num: 2, label: 'Аудитория' },
  { num: 3, label: 'Триггеры' },
  { num: 4, label: 'Действия' },
] as const

const TRIGGER_TYPES = [
  { value: 'purchase', label: 'Факт покупки' },
  { value: 'purchase_product', label: 'Покупка конкретного товара' },
  { value: 'purchase_min_items', label: 'Мин. кол-во позиций в чеке' },
  { value: 'receipt_sum', label: 'Сумма чека' },
  { value: 'event', label: 'Событие' },
] as const

const EVENT_TYPES = [
  { value: 'birthday', label: 'День рождения' },
  { value: 'registration', label: 'Регистрация' },
  { value: 'activation', label: 'Активация' },
  { value: 'last_purchase', label: 'Последняя покупка' },
] as const

const ACTION_TYPES = [
  { value: 'discount', label: 'Скидка' },
  { value: 'bonus', label: 'Бонусы' },
  { value: 'data_update', label: 'Обновление данных' },
  { value: 'campaign', label: 'Рассылка' },
] as const

// ── Styles ───────────────────────────────────────────────────────────────────

const inputClass = cn(
  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
  'text-sm placeholder:text-neutral-400',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
  'disabled:opacity-50 disabled:cursor-not-allowed',
)

const labelClass = 'block text-sm font-medium text-neutral-700 mb-1.5'

const cardClass = 'bg-white rounded-2xl shadow-sm border border-surface-border p-6'

const itemCardClass = 'bg-neutral-50 rounded-xl p-4 border border-surface-border relative'

const btnPrimaryClass = cn(
  'flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium',
  'bg-accent text-white',
  'hover:bg-accent/90 active:bg-accent/80',
  'transition-colors',
  'disabled:opacity-50 disabled:cursor-not-allowed',
  'focus:outline-none focus:ring-2 focus:ring-accent/20',
)

const btnSecondaryClass = cn(
  'px-5 py-2.5 rounded-lg text-sm font-medium',
  'border border-surface-border text-neutral-700',
  'hover:bg-neutral-50 transition-colors',
)

// ── Component ────────────────────────────────────────────────────────────────

export default function CreatePromotionPage() {
  const navigate = useNavigate()
  const createMutation = useCreatePromotionMutation()
  const previewMutation = usePreviewAudienceMutation()

  // Wizard step
  const [step, setStep] = useState(1)

  // Step 1: Basic settings
  const [name, setName] = useState('')
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [unlimited, setUnlimited] = useState(true)
  const [usageLimit, setUsageLimit] = useState<number | ''>('')
  const [combinable, setCombinable] = useState(false)
  const [combinableWithLoyalty, setCombinableWithLoyalty] = useState(false)

  // Step 2: Audience filter
  const [filter, setFilter] = useState<SegmentFilter>({})

  // Step 3: Triggers
  const [triggers, setTriggers] = useState<PromotionTrigger[]>([])

  // Step 4: Actions
  const [actions, setActions] = useState<PromotionAction[]>([])

  // ── Validation ─────────────────────────────────────────────────────────────

  function isStepValid(s: number): boolean {
    switch (s) {
      case 1:
        return name.trim().length > 0
      case 2:
        return true // filter is optional
      case 3:
        return true // triggers are optional
      case 4:
        return actions.length > 0
      default:
        return false
    }
  }

  // ── Trigger helpers ────────────────────────────────────────────────────────

  function addTrigger() {
    setTriggers((prev) => [...prev, { type: 'purchase' }])
  }

  function updateTrigger(index: number, updated: PromotionTrigger) {
    setTriggers((prev) => prev.map((t, i) => (i === index ? updated : t)))
  }

  function removeTrigger(index: number) {
    setTriggers((prev) => prev.filter((_, i) => i !== index))
  }

  // ── Action helpers ─────────────────────────────────────────────────────────

  function addAction() {
    setActions((prev) => [...prev, { type: 'discount' }])
  }

  function updateAction(index: number, updated: PromotionAction) {
    setActions((prev) => prev.map((a, i) => (i === index ? updated : a)))
  }

  function removeAction(index: number) {
    setActions((prev) => prev.filter((_, i) => i !== index))
  }

  // ── Submit ─────────────────────────────────────────────────────────────────

  function handleSubmit() {
    if (!isStepValid(4)) return

    const data: CreatePromotionRequest = {
      name: name.trim(),
      type:
        actions[0]?.type === 'discount'
          ? 'discount'
          : actions[0]?.type === 'bonus'
            ? 'bonus'
            : actions[0]?.type === 'data_update'
              ? 'tag_update'
              : 'campaign',
      conditions: {},
      result: {},
      starts_at: startsAt ? new Date(startsAt).toISOString() : undefined,
      ends_at: endsAt ? new Date(endsAt).toISOString() : undefined,
      usage_limit: unlimited ? undefined : (usageLimit as number) || undefined,
      combinable,
      // Extended fields — passed via conditions/result or top-level if API supports
      ...(Object.keys(filter).length > 0 && { filter }),
      ...(triggers.length > 0 && { triggers }),
      ...(actions.length > 0 && { actions }),
      ...(combinableWithLoyalty && { combinable_with_loyalty: combinableWithLoyalty }),
    } as CreatePromotionRequest

    createMutation.mutate(data, {
      onSuccess: () => navigate('/dashboard/promotions'),
    })
  }

  // ── Navigation ─────────────────────────────────────────────────────────────

  function goNext() {
    if (step < 4 && isStepValid(step)) setStep(step + 1)
    if (step === 4) handleSubmit()
  }

  function goBack() {
    if (step > 1) setStep(step - 1)
  }

  // ── Render ─────────────────────────────────────────────────────────────────

  return (
    <div className="max-w-2xl">
      {/* Header */}
      <div className="flex items-center gap-4 mb-6">
        <button
          onClick={() => navigate('/dashboard/promotions')}
          type="button"
          className="p-2 rounded-lg hover:bg-neutral-100 transition-colors"
          aria-label="Назад к акциям"
        >
          <ArrowLeft className="w-5 h-5 text-neutral-500" />
        </button>
        <h1 className="font-serif text-2xl font-bold text-neutral-900 tracking-tight">
          Создать акцию
        </h1>
      </div>

      {/* Step indicator */}
      <nav aria-label="Шаги создания акции" className="mb-8">
        <ol className="flex items-center gap-2">
          {STEPS.map((s, i) => {
            const isActive = s.num === step
            const isCompleted = s.num < step
            return (
              <li key={s.num} className="flex items-center gap-2 flex-1">
                <button
                  type="button"
                  onClick={() => {
                    // Allow going back to completed steps
                    if (s.num < step) setStep(s.num)
                  }}
                  disabled={s.num > step}
                  className={cn(
                    'flex items-center gap-2 w-full rounded-lg px-3 py-2.5 text-left transition-colors',
                    isActive && 'bg-accent/10',
                    isCompleted && 'cursor-pointer hover:bg-neutral-50',
                    s.num > step && 'cursor-default',
                  )}
                  aria-current={isActive ? 'step' : undefined}
                >
                  <span
                    className={cn(
                      'flex-shrink-0 w-7 h-7 rounded-full flex items-center justify-center text-xs font-semibold transition-colors',
                      isActive && 'bg-accent text-white',
                      isCompleted && 'bg-accent/20 text-accent',
                      !isActive && !isCompleted && 'bg-neutral-100 text-neutral-400',
                    )}
                  >
                    {isCompleted ? (
                      <Check className="w-3.5 h-3.5" />
                    ) : (
                      s.num
                    )}
                  </span>
                  <span
                    className={cn(
                      'text-xs font-medium hidden sm:block',
                      isActive && 'text-accent',
                      isCompleted && 'text-neutral-600',
                      !isActive && !isCompleted && 'text-neutral-400',
                    )}
                  >
                    {s.label}
                  </span>
                </button>
                {i < STEPS.length - 1 && (
                  <div
                    className={cn(
                      'hidden sm:block h-px flex-shrink-0 w-4',
                      isCompleted ? 'bg-accent/30' : 'bg-neutral-200',
                    )}
                  />
                )}
              </li>
            )
          })}
        </ol>
      </nav>

      {/* Step content */}
      <div className="space-y-6">
        {step === 1 && (
          <StepBasicSettings
            name={name}
            setName={setName}
            startsAt={startsAt}
            setStartsAt={setStartsAt}
            endsAt={endsAt}
            setEndsAt={setEndsAt}
            unlimited={unlimited}
            setUnlimited={setUnlimited}
            usageLimit={usageLimit}
            setUsageLimit={setUsageLimit}
            combinable={combinable}
            setCombinable={setCombinable}
            combinableWithLoyalty={combinableWithLoyalty}
            setCombinableWithLoyalty={setCombinableWithLoyalty}
          />
        )}

        {step === 2 && (
          <StepAudienceFilter
            filter={filter}
            setFilter={setFilter}
            previewCount={previewMutation.isSuccess ? previewMutation.data : null}
            onPreview={() => previewMutation.mutate(filter)}
            isPreviewing={previewMutation.isPending}
          />
        )}

        {step === 3 && (
          <StepTriggers
            triggers={triggers}
            onAdd={addTrigger}
            onUpdate={updateTrigger}
            onRemove={removeTrigger}
          />
        )}

        {step === 4 && (
          <StepActions
            actions={actions}
            onAdd={addAction}
            onUpdate={updateAction}
            onRemove={removeAction}
          />
        )}

        {/* Error message */}
        {createMutation.isError && (
          <p className="text-sm text-red-600" role="alert">
            Не удалось создать акцию. Попробуйте ещё раз.
          </p>
        )}

        {/* Bottom navigation */}
        <div className="flex items-center justify-between pt-2">
          <button
            type="button"
            onClick={() => navigate('/dashboard/promotions')}
            className="text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
          >
            Отмена
          </button>

          <div className="flex items-center gap-3">
            {step > 1 && (
              <button type="button" onClick={goBack} className={btnSecondaryClass}>
                Назад
              </button>
            )}
            <button
              type="button"
              onClick={goNext}
              disabled={!isStepValid(step) || (step === 4 && createMutation.isPending)}
              className={btnPrimaryClass}
            >
              {step === 4
                ? createMutation.isPending
                  ? 'Создание...'
                  : 'Создать акцию'
                : 'Далее'}
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// ── Step 1: Basic Settings ───────────────────────────────────────────────────

interface StepBasicSettingsProps {
  name: string
  setName: (v: string) => void
  startsAt: string
  setStartsAt: (v: string) => void
  endsAt: string
  setEndsAt: (v: string) => void
  unlimited: boolean
  setUnlimited: (v: boolean) => void
  usageLimit: number | ''
  setUsageLimit: (v: number | '') => void
  combinable: boolean
  setCombinable: (v: boolean) => void
  combinableWithLoyalty: boolean
  setCombinableWithLoyalty: (v: boolean) => void
}

function StepBasicSettings({
  name,
  setName,
  startsAt,
  setStartsAt,
  endsAt,
  setEndsAt,
  unlimited,
  setUnlimited,
  usageLimit,
  setUsageLimit,
  combinable,
  setCombinable,
  combinableWithLoyalty,
  setCombinableWithLoyalty,
}: StepBasicSettingsProps) {
  return (
    <div className={cn(cardClass, 'space-y-5')}>
      <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-1">
        Основные параметры
      </p>

      {/* Name */}
      <div>
        <label htmlFor="promo-name" className={labelClass}>
          Название акции <span className="text-red-400">*</span>
        </label>
        <input
          id="promo-name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Например: Скидка 20% на завтраки"
          className={inputClass}
          autoFocus
        />
      </div>

      {/* Period */}
      <fieldset>
        <legend className={labelClass}>Период действия</legend>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label htmlFor="promo-starts" className="block text-xs text-neutral-500 mb-1">
              Начало
            </label>
            <input
              id="promo-starts"
              type="date"
              value={startsAt}
              onChange={(e) => setStartsAt(e.target.value)}
              className={inputClass}
            />
          </div>
          <div>
            <label htmlFor="promo-ends" className="block text-xs text-neutral-500 mb-1">
              Окончание
            </label>
            <input
              id="promo-ends"
              type="date"
              value={endsAt}
              onChange={(e) => setEndsAt(e.target.value)}
              className={inputClass}
            />
          </div>
        </div>
      </fieldset>

      {/* Usage limit */}
      <div className="space-y-3">
        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={unlimited}
            onChange={(e) => setUnlimited(e.target.checked)}
            className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
          />
          <span className="text-sm text-neutral-700">Без ограничения количества использований</span>
        </label>

        {!unlimited && (
          <div>
            <label htmlFor="promo-usage-limit" className={labelClass}>
              Лимит использований
            </label>
            <input
              id="promo-usage-limit"
              type="number"
              min={1}
              value={usageLimit}
              onChange={(e) => setUsageLimit(e.target.value ? Number(e.target.value) : '')}
              placeholder="100"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}
      </div>

      {/* Combinable checkboxes */}
      <div className="space-y-3 pt-2 border-t border-surface-border">
        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={combinable}
            onChange={(e) => setCombinable(e.target.checked)}
            className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
          />
          <span className="text-sm text-neutral-700">Совместима с другими акциями</span>
        </label>

        <label className="flex items-center gap-3 cursor-pointer">
          <input
            type="checkbox"
            checked={combinableWithLoyalty}
            onChange={(e) => setCombinableWithLoyalty(e.target.checked)}
            className="w-4 h-4 rounded border-neutral-300 text-accent focus:ring-accent/20"
          />
          <span className="text-sm text-neutral-700">Совместима с программой лояльности</span>
        </label>
      </div>
    </div>
  )
}

// ── Step 2: Audience Filter ──────────────────────────────────────────────────

interface StepAudienceFilterProps {
  filter: SegmentFilter
  setFilter: (f: SegmentFilter) => void
  previewCount: number | null
  onPreview: () => void
  isPreviewing: boolean
}

function StepAudienceFilter({
  filter,
  setFilter,
  previewCount,
  onPreview,
  isPreviewing,
}: StepAudienceFilterProps) {
  return (
    <div className={cn(cardClass, 'space-y-5')}>
      <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-1">
        Фильтр аудитории
      </p>
      <p className="text-sm text-neutral-500">
        Настройте фильтр, чтобы акция распространялась только на определённых клиентов.
        Оставьте пустым для применения ко всем.
      </p>
      <ClientFilterBuilder
        value={filter}
        onChange={setFilter}
        previewCount={previewCount}
        onPreview={onPreview}
        isPreviewing={isPreviewing}
      />
    </div>
  )
}

// ── Step 3: Triggers ─────────────────────────────────────────────────────────

interface StepTriggersProps {
  triggers: PromotionTrigger[]
  onAdd: () => void
  onUpdate: (index: number, trigger: PromotionTrigger) => void
  onRemove: (index: number) => void
}

function StepTriggers({ triggers, onAdd, onUpdate, onRemove }: StepTriggersProps) {
  return (
    <div className={cn(cardClass, 'space-y-5')}>
      <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-1">
        Триггеры
      </p>
      <p className="text-sm text-neutral-500">
        Укажите условия, при которых акция срабатывает. Можно добавить несколько триггеров.
      </p>

      {triggers.length > 0 && (
        <div className="space-y-3">
          {triggers.map((trigger, index) => (
            <TriggerCard
              key={index}
              trigger={trigger}
              onChange={(t) => onUpdate(index, t)}
              onRemove={() => onRemove(index)}
            />
          ))}
        </div>
      )}

      <button type="button" onClick={onAdd} className={cn(btnSecondaryClass, 'flex items-center gap-2')}>
        <Plus className="w-4 h-4" />
        Добавить триггер
      </button>
    </div>
  )
}

interface TriggerCardProps {
  trigger: PromotionTrigger
  onChange: (t: PromotionTrigger) => void
  onRemove: () => void
}

function TriggerCard({ trigger, onChange, onRemove }: TriggerCardProps) {
  return (
    <div className={itemCardClass}>
      <button
        type="button"
        onClick={onRemove}
        className="absolute top-3 right-3 p-1 rounded-md hover:bg-neutral-200 transition-colors"
        aria-label="Удалить триггер"
      >
        <X className="w-4 h-4 text-neutral-400" />
      </button>

      <div className="space-y-4 pr-8">
        {/* Trigger type */}
        <div>
          <label className={labelClass}>Тип триггера</label>
          <select
            value={trigger.type}
            onChange={(e) => onChange({ type: e.target.value })}
            className={inputClass}
          >
            {TRIGGER_TYPES.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* Conditional fields */}
        {trigger.type === 'purchase_product' && (
          <div>
            <label className={labelClass}>ID товара</label>
            <input
              type="number"
              min={1}
              value={trigger.product_id ?? ''}
              onChange={(e) =>
                onChange({ ...trigger, product_id: e.target.value ? Number(e.target.value) : undefined })
              }
              placeholder="12345"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}

        {trigger.type === 'purchase_min_items' && (
          <div>
            <label className={labelClass}>Мин. кол-во позиций</label>
            <input
              type="number"
              min={1}
              value={trigger.min_items ?? ''}
              onChange={(e) =>
                onChange({ ...trigger, min_items: e.target.value ? Number(e.target.value) : undefined })
              }
              placeholder="3"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}

        {trigger.type === 'receipt_sum' && (
          <div>
            <label className={labelClass}>Минимальная сумма чека</label>
            <input
              type="number"
              min={1}
              value={trigger.min_amount ?? ''}
              onChange={(e) =>
                onChange({ ...trigger, min_amount: e.target.value ? Number(e.target.value) : undefined })
              }
              placeholder="1000"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}

        {trigger.type === 'event' && (
          <div>
            <label className={labelClass}>Тип события</label>
            <select
              value={trigger.event_type ?? ''}
              onChange={(e) => onChange({ ...trigger, event_type: e.target.value || undefined })}
              className={inputClass}
            >
              <option value="">Выберите событие</option>
              {EVENT_TYPES.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        )}
      </div>
    </div>
  )
}

// ── Step 4: Actions ──────────────────────────────────────────────────────────

interface StepActionsProps {
  actions: PromotionAction[]
  onAdd: () => void
  onUpdate: (index: number, action: PromotionAction) => void
  onRemove: (index: number) => void
}

function StepActions({ actions, onAdd, onUpdate, onRemove }: StepActionsProps) {
  return (
    <div className={cn(cardClass, 'space-y-5')}>
      <p className="font-mono text-[10px] uppercase tracking-widest text-neutral-400 mb-1">
        Действия
      </p>
      <p className="text-sm text-neutral-500">
        Укажите, что произойдёт при срабатывании акции. Нужно хотя бы одно действие.
      </p>

      {actions.length > 0 && (
        <div className="space-y-3">
          {actions.map((action, index) => (
            <ActionCard
              key={index}
              action={action}
              onChange={(a) => onUpdate(index, a)}
              onRemove={() => onRemove(index)}
            />
          ))}
        </div>
      )}

      <button type="button" onClick={onAdd} className={cn(btnSecondaryClass, 'flex items-center gap-2')}>
        <Plus className="w-4 h-4" />
        Добавить действие
      </button>
    </div>
  )
}

interface ActionCardProps {
  action: PromotionAction
  onChange: (a: PromotionAction) => void
  onRemove: () => void
}

function ActionCard({ action, onChange, onRemove }: ActionCardProps) {
  return (
    <div className={itemCardClass}>
      <button
        type="button"
        onClick={onRemove}
        className="absolute top-3 right-3 p-1 rounded-md hover:bg-neutral-200 transition-colors"
        aria-label="Удалить действие"
      >
        <X className="w-4 h-4 text-neutral-400" />
      </button>

      <div className="space-y-4 pr-8">
        {/* Action type */}
        <div>
          <label className={labelClass}>Тип действия</label>
          <select
            value={action.type}
            onChange={(e) => onChange({ type: e.target.value })}
            className={inputClass}
          >
            {ACTION_TYPES.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </div>

        {/* Conditional fields */}
        {action.type === 'discount' && (
          <div>
            <label className={labelClass}>Процент скидки</label>
            <input
              type="number"
              min={1}
              max={100}
              value={action.discount_percent ?? ''}
              onChange={(e) =>
                onChange({ ...action, discount_percent: e.target.value ? Number(e.target.value) : undefined })
              }
              placeholder="20"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}

        {action.type === 'bonus' && (
          <div>
            <label className={labelClass}>Количество бонусов</label>
            <input
              type="number"
              min={1}
              value={action.bonus_amount ?? ''}
              onChange={(e) =>
                onChange({ ...action, bonus_amount: e.target.value ? Number(e.target.value) : undefined })
              }
              placeholder="500"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}

        {action.type === 'data_update' && (
          <div className="space-y-4">
            <div>
              <label className={labelClass}>Добавить тег</label>
              <input
                type="text"
                value={action.tag_add ?? ''}
                onChange={(e) => onChange({ ...action, tag_add: e.target.value || undefined })}
                placeholder="vip_client"
                className={cn(inputClass, 'max-w-xs')}
              />
            </div>
            <div>
              <label className={labelClass}>ID уровня</label>
              <input
                type="number"
                min={1}
                value={action.level_id ?? ''}
                onChange={(e) =>
                  onChange({ ...action, level_id: e.target.value ? Number(e.target.value) : undefined })
                }
                placeholder="2"
                className={cn(inputClass, 'max-w-xs')}
              />
            </div>
          </div>
        )}

        {action.type === 'campaign' && (
          <div>
            <label className={labelClass}>ID рассылки</label>
            <input
              type="number"
              min={1}
              value={action.campaign_id ?? ''}
              onChange={(e) =>
                onChange({ ...action, campaign_id: e.target.value ? Number(e.target.value) : undefined })
              }
              placeholder="1"
              className={cn(inputClass, 'max-w-xs')}
            />
          </div>
        )}
      </div>
    </div>
  )
}
