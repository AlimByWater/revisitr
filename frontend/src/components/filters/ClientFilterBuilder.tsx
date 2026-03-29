import { useState, useCallback } from 'react'
import { cn } from '@/lib/utils'
import { ChevronDown, Eye, RotateCcw } from 'lucide-react'
import type { SegmentFilter } from '@/features/segments/types'

const RFM_SEGMENT_LABELS: Record<string, string> = {
  new: 'Новые',
  promising: 'Перспективные',
  regular: 'Регулярные',
  vip: 'VIP / Ядро',
  rare_valuable: 'Редкие, но ценные',
  churn_risk: 'На грани оттока',
  lost: 'Потерянные',
}

const OS_OPTIONS = [
  { value: 'android', label: 'Android' },
  { value: 'ios', label: 'iOS' },
  { value: 'web', label: 'Web' },
]

const GENDER_OPTIONS = [
  { value: 'male', label: 'Мужской' },
  { value: 'female', label: 'Женский' },
]

const inputClassName = cn(
  'w-full px-3 py-2 rounded-lg border border-surface-border text-sm',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
)

export interface ClientFilterBuilderProps {
  value: SegmentFilter
  onChange: (filter: SegmentFilter) => void
  previewCount?: number | null
  onPreview?: () => void
  isPreviewing?: boolean
  hiddenFields?: string[]
  className?: string
}

type FilterKey = keyof SegmentFilter

/** Remove undefined/null/empty values from filter, keeping only meaningful entries */
function cleanFilter(filter: SegmentFilter): SegmentFilter {
  const cleaned: SegmentFilter = {}
  for (const [key, value] of Object.entries(filter)) {
    if (value === undefined || value === null || value === '') continue
    if (Array.isArray(value) && value.length === 0) continue
    ;(cleaned as Record<string, unknown>)[key] = value
  }
  return cleaned
}

/** Count active (non-empty) filter values for a set of keys */
function countActiveFilters(filter: SegmentFilter, keys: FilterKey[]): number {
  let count = 0
  for (const key of keys) {
    const val = filter[key]
    if (val === undefined || val === null || val === '') continue
    if (Array.isArray(val) && val.length === 0) continue
    count++
  }
  return count
}

interface FilterGroupProps {
  title: string
  activeCount: number
  children: React.ReactNode
  defaultOpen?: boolean
}

function FilterGroup({ title, activeCount, children, defaultOpen = false }: FilterGroupProps) {
  const [open, setOpen] = useState(defaultOpen)

  return (
    <div className="border border-surface-border rounded-xl overflow-hidden">
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className={cn(
          'w-full flex items-center justify-between px-4 py-3',
          'text-sm font-medium text-neutral-700 hover:bg-neutral-50 transition-colors',
        )}
      >
        <span className="flex items-center gap-2">
          {title}
          {activeCount > 0 && (
            <span className="inline-flex items-center justify-center min-w-[20px] h-5 px-1.5 rounded-full bg-accent/10 text-accent text-xs font-semibold tabular-nums">
              {activeCount}
            </span>
          )}
        </span>
        <ChevronDown
          className={cn(
            'w-4 h-4 text-neutral-400 transition-transform duration-200',
            open && 'rotate-180',
          )}
        />
      </button>
      {open && (
        <div className="px-4 pb-4 pt-1 space-y-3 border-t border-surface-border/50">
          {children}
        </div>
      )}
    </div>
  )
}

export function ClientFilterBuilder({
  value,
  onChange,
  previewCount,
  onPreview,
  isPreviewing,
  hiddenFields = [],
  className,
}: ClientFilterBuilderProps) {
  const isHidden = useCallback(
    (field: string) => hiddenFields.includes(field),
    [hiddenFields],
  )

  function update(key: FilterKey, val: unknown) {
    const next = { ...value, [key]: val }
    onChange(cleanFilter(next))
  }

  function updateNumber(key: FilterKey, raw: string) {
    if (raw === '') {
      const next = { ...value }
      delete next[key]
      onChange(cleanFilter(next))
    } else {
      update(key, Number(raw))
    }
  }

  function updateString(key: FilterKey, raw: string) {
    if (raw === '') {
      const next = { ...value }
      delete next[key]
      onChange(cleanFilter(next))
    } else {
      update(key, raw)
    }
  }

  function handleReset() {
    // Keep hidden fields in the filter
    const kept: SegmentFilter = {}
    for (const field of hiddenFields) {
      const val = value[field as FilterKey]
      if (val !== undefined) {
        ;(kept as Record<string, unknown>)[field] = val
      }
    }
    onChange(kept)
  }

  const demographyKeys: FilterKey[] = ['search', 'gender', 'age_from', 'age_to', 'city']
  const activityKeys: FilterKey[] = [
    'registered_from',
    'registered_to',
    'min_visits',
    'max_visits',
    'min_spend',
    'max_spend',
  ]
  const loyaltyKeys: FilterKey[] = [
    'level_id',
    'min_balance',
    'max_balance',
    'min_spent_points',
    'max_spent_points',
  ]
  const segmentationKeys: FilterKey[] = ['rfm_category', 'tags', 'os']
  const botKeys: FilterKey[] = ['bot_id']

  const totalActive =
    countActiveFilters(value, demographyKeys) +
    countActiveFilters(value, activityKeys) +
    countActiveFilters(value, loyaltyKeys) +
    countActiveFilters(value, segmentationKeys) +
    (!isHidden('bot_id') ? countActiveFilters(value, botKeys) : 0)

  return (
    <div className={cn('bg-white rounded-2xl border border-surface-border', className)}>
      <div className="p-4 space-y-2">
        {/* Демография */}
        <FilterGroup
          title="Демография"
          activeCount={countActiveFilters(value, demographyKeys.filter((k) => !isHidden(k)))}
          defaultOpen
        >
          {!isHidden('search') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Поиск по имени / телефону
              </label>
              <input
                type="text"
                value={value.search ?? ''}
                onChange={(e) => updateString('search', e.target.value)}
                placeholder="Имя, телефон..."
                className={inputClassName}
              />
            </div>
          )}
          {!isHidden('gender') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">Пол</label>
              <select
                value={value.gender ?? ''}
                onChange={(e) => updateString('gender', e.target.value)}
                className={inputClassName}
              >
                <option value="">Любой</option>
                {GENDER_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
          )}
          {(!isHidden('age_from') || !isHidden('age_to')) && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">Возраст</label>
              <div className="flex items-center gap-2">
                {!isHidden('age_from') && (
                  <input
                    type="number"
                    min={0}
                    value={value.age_from ?? ''}
                    onChange={(e) => updateNumber('age_from', e.target.value)}
                    placeholder="от"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
                <span className="text-neutral-400 text-sm shrink-0">&mdash;</span>
                {!isHidden('age_to') && (
                  <input
                    type="number"
                    min={0}
                    value={value.age_to ?? ''}
                    onChange={(e) => updateNumber('age_to', e.target.value)}
                    placeholder="до"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
              </div>
            </div>
          )}
          {!isHidden('city') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">Город</label>
              <input
                type="text"
                value={value.city ?? ''}
                onChange={(e) => updateString('city', e.target.value)}
                placeholder="Москва"
                className={inputClassName}
              />
            </div>
          )}
        </FilterGroup>

        {/* Активность */}
        <FilterGroup
          title="Активность"
          activeCount={countActiveFilters(value, activityKeys.filter((k) => !isHidden(k)))}
        >
          {(!isHidden('registered_from') || !isHidden('registered_to')) && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Дата регистрации
              </label>
              <div className="flex items-center gap-2">
                {!isHidden('registered_from') && (
                  <input
                    type="date"
                    value={value.registered_from ?? ''}
                    onChange={(e) => updateString('registered_from', e.target.value)}
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
                <span className="text-neutral-400 text-sm shrink-0">&mdash;</span>
                {!isHidden('registered_to') && (
                  <input
                    type="date"
                    value={value.registered_to ?? ''}
                    onChange={(e) => updateString('registered_to', e.target.value)}
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
              </div>
            </div>
          )}
          {(!isHidden('min_visits') || !isHidden('max_visits')) && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Количество визитов
              </label>
              <div className="flex items-center gap-2">
                {!isHidden('min_visits') && (
                  <input
                    type="number"
                    min={0}
                    value={value.min_visits ?? ''}
                    onChange={(e) => updateNumber('min_visits', e.target.value)}
                    placeholder="от"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
                <span className="text-neutral-400 text-sm shrink-0">&mdash;</span>
                {!isHidden('max_visits') && (
                  <input
                    type="number"
                    min={0}
                    value={value.max_visits ?? ''}
                    onChange={(e) => updateNumber('max_visits', e.target.value)}
                    placeholder="до"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
              </div>
            </div>
          )}
          {(!isHidden('min_spend') || !isHidden('max_spend')) && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Сумма покупок
              </label>
              <div className="flex items-center gap-2">
                {!isHidden('min_spend') && (
                  <input
                    type="number"
                    min={0}
                    value={value.min_spend ?? ''}
                    onChange={(e) => updateNumber('min_spend', e.target.value)}
                    placeholder="от"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
                <span className="text-neutral-400 text-sm shrink-0">&mdash;</span>
                {!isHidden('max_spend') && (
                  <input
                    type="number"
                    min={0}
                    value={value.max_spend ?? ''}
                    onChange={(e) => updateNumber('max_spend', e.target.value)}
                    placeholder="до"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
              </div>
            </div>
          )}
        </FilterGroup>

        {/* Лояльность */}
        <FilterGroup
          title="Лояльность"
          activeCount={countActiveFilters(value, loyaltyKeys.filter((k) => !isHidden(k)))}
        >
          {!isHidden('level_id') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">Уровень</label>
              <input
                type="number"
                min={1}
                value={value.level_id ?? ''}
                onChange={(e) => updateNumber('level_id', e.target.value)}
                placeholder="ID уровня"
                className={inputClassName}
              />
            </div>
          )}
          {(!isHidden('min_balance') || !isHidden('max_balance')) && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Баланс баллов
              </label>
              <div className="flex items-center gap-2">
                {!isHidden('min_balance') && (
                  <input
                    type="number"
                    min={0}
                    value={value.min_balance ?? ''}
                    onChange={(e) => updateNumber('min_balance', e.target.value)}
                    placeholder="от"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
                <span className="text-neutral-400 text-sm shrink-0">&mdash;</span>
                {!isHidden('max_balance') && (
                  <input
                    type="number"
                    min={0}
                    value={value.max_balance ?? ''}
                    onChange={(e) => updateNumber('max_balance', e.target.value)}
                    placeholder="до"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
              </div>
            </div>
          )}
          {(!isHidden('min_spent_points') || !isHidden('max_spent_points')) && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Потрачено баллов
              </label>
              <div className="flex items-center gap-2">
                {!isHidden('min_spent_points') && (
                  <input
                    type="number"
                    min={0}
                    value={value.min_spent_points ?? ''}
                    onChange={(e) => updateNumber('min_spent_points', e.target.value)}
                    placeholder="от"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
                <span className="text-neutral-400 text-sm shrink-0">&mdash;</span>
                {!isHidden('max_spent_points') && (
                  <input
                    type="number"
                    min={0}
                    value={value.max_spent_points ?? ''}
                    onChange={(e) => updateNumber('max_spent_points', e.target.value)}
                    placeholder="до"
                    className={cn(inputClassName, 'flex-1')}
                  />
                )}
              </div>
            </div>
          )}
        </FilterGroup>

        {/* Сегментация */}
        <FilterGroup
          title="Сегментация"
          activeCount={countActiveFilters(value, segmentationKeys.filter((k) => !isHidden(k)))}
        >
          {!isHidden('rfm_category') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                RFM-сегмент
              </label>
              <select
                value={value.rfm_category ?? ''}
                onChange={(e) => updateString('rfm_category', e.target.value)}
                className={inputClassName}
              >
                <option value="">Любой</option>
                {Object.entries(RFM_SEGMENT_LABELS).map(([key, label]) => (
                  <option key={key} value={key}>
                    {label}
                  </option>
                ))}
              </select>
            </div>
          )}
          {!isHidden('tags') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">
                Теги (через запятую)
              </label>
              <input
                type="text"
                value={value.tags?.join(', ') ?? ''}
                onChange={(e) => {
                  const raw = e.target.value
                  if (raw.trim() === '') {
                    const next = { ...value }
                    delete next.tags
                    onChange(cleanFilter(next))
                  } else {
                    update(
                      'tags',
                      raw
                        .split(',')
                        .map((t) => t.trim())
                        .filter(Boolean),
                    )
                  }
                }}
                placeholder="vip, студенты..."
                className={inputClassName}
              />
            </div>
          )}
          {!isHidden('os') && (
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">Устройство</label>
              <select
                value={value.os ?? ''}
                onChange={(e) => updateString('os', e.target.value)}
                className={inputClassName}
              >
                <option value="">Любое</option>
                {OS_OPTIONS.map((o) => (
                  <option key={o.value} value={o.value}>
                    {o.label}
                  </option>
                ))}
              </select>
            </div>
          )}
        </FilterGroup>

        {/* Бот */}
        {!isHidden('bot_id') && (
          <FilterGroup
            title="Бот"
            activeCount={countActiveFilters(value, botKeys)}
          >
            <div>
              <label className="block text-xs font-medium text-neutral-500 mb-1">ID бота</label>
              <input
                type="number"
                min={1}
                value={value.bot_id ?? ''}
                onChange={(e) => updateNumber('bot_id', e.target.value)}
                placeholder="ID бота"
                className={inputClassName}
              />
            </div>
          </FilterGroup>
        )}
      </div>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-surface-border flex items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {onPreview && (
            <button
              type="button"
              onClick={onPreview}
              disabled={isPreviewing}
              className={cn(
                'flex items-center gap-1.5 py-2 px-3 rounded-lg text-sm font-medium',
                'border border-surface-border text-neutral-700',
                'hover:bg-neutral-50 transition-colors',
                'disabled:opacity-50',
              )}
            >
              <Eye className="w-4 h-4" />
              {isPreviewing ? 'Подсчёт...' : 'Посчитать'}
            </button>
          )}
          {previewCount != null && (
            <span className="text-sm text-neutral-500">
              Найдено: <strong className="text-neutral-900">{previewCount}</strong> клиентов
            </span>
          )}
        </div>

        <button
          type="button"
          onClick={handleReset}
          disabled={totalActive === 0}
          className={cn(
            'flex items-center gap-1.5 py-2 px-3 rounded-lg text-sm font-medium',
            'text-neutral-500 hover:text-neutral-700 hover:bg-neutral-50 transition-colors',
            'disabled:opacity-30 disabled:cursor-not-allowed',
          )}
        >
          <RotateCcw className="w-3.5 h-3.5" />
          Сбросить фильтры
        </button>
      </div>
    </div>
  )
}
