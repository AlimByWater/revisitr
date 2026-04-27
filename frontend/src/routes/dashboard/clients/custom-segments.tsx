import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  useSegmentsQuery,
  useCreateSegmentMutation,
  useDeleteSegmentMutation,
  useUpdateSegmentMutation,
} from '@/features/segments/queries'
import { segmentsApi } from '@/features/segments/api'
import type { SegmentFilter, Segment } from '@/features/segments/types'
import { RULE_OPERATORS } from '@/features/segments/types'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { Plus, Trash2, Users, Filter, X, Eye, Pencil } from 'lucide-react'
import { CustomSelect } from '@/components/common/CustomSelect'
import { Button } from '@/components/common/Button'

// Filterable client attributes from the design doc
const FILTER_ATTRIBUTES = [
  { key: 'gender', label: 'Пол', type: 'select', options: ['Мужской', 'Женский'] },
  { key: 'age_from', label: 'Возраст от', type: 'number' },
  { key: 'age_to', label: 'Возраст до', type: 'number' },
  { key: 'registered_after', label: 'Дата регистрации от', type: 'date' },
  { key: 'registered_before', label: 'Дата регистрации до', type: 'date' },
  { key: 'loyalty_level', label: 'Уровень лояльности', type: 'text' },
  { key: 'rfm_segment', label: 'RFM-сегмент', type: 'select', options: ['new', 'promising', 'regular', 'vip', 'rare_valuable', 'churn_risk', 'lost'] },
  { key: 'min_spend', label: 'Мин. сумма покупок', type: 'number' },
  { key: 'max_spend', label: 'Макс. сумма покупок', type: 'number' },
  { key: 'min_visits', label: 'Мин. кол-во покупок', type: 'number' },
  { key: 'max_visits', label: 'Макс. кол-во покупок', type: 'number' },
  { key: 'min_balance', label: 'Мин. баланс', type: 'number' },
  { key: 'city', label: 'Город', type: 'text' },
  { key: 'os', label: 'Устройство', type: 'select', options: ['android', 'ios', 'web'] },
  { key: 'has_telegram', label: 'Telegram', type: 'boolean' },
  { key: 'tags', label: 'Теги', type: 'text' },
] as const

interface FilterRule {
  attribute: string
  operator: string
  value: string
}

const inputClassName = cn(
  'w-full px-3 py-2 rounded border border-neutral-200 text-sm',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
)

export default function CustomSegmentsPage() {
  const { data: segments, isLoading } = useSegmentsQuery()
  const createSegment = useCreateSegmentMutation()
  const deleteSegment = useDeleteSegmentMutation()
  const updateSegment = useUpdateSegmentMutation()

  const [showBuilder, setShowBuilder] = useState(false)
  const [editingSegment, setEditingSegment] = useState<Segment | null>(null)
  const [name, setName] = useState('')
  const [rules, setRules] = useState<FilterRule[]>([{ attribute: 'gender', operator: 'eq', value: '' }])
  const [previewCount, setPreviewCount] = useState<number | null>(null)
  const [previewing, setPreviewing] = useState(false)

  const customSegments = (segments ?? []).filter((s) => s.type === 'custom')

  const addRule = () => {
    setRules([...rules, { attribute: 'gender', operator: 'eq', value: '' }])
  }

  const removeRule = (index: number) => {
    setRules(rules.filter((_, i) => i !== index))
  }

  const updateRule = (index: number, field: keyof FilterRule, value: string) => {
    setRules(rules.map((r, i) => (i === index ? { ...r, [field]: value } : r)))
  }

  const buildFilter = (): SegmentFilter => {
    const filter: SegmentFilter = {}
    for (const rule of rules) {
      if (!rule.value) continue
      const attr = FILTER_ATTRIBUTES.find((a) => a.key === rule.attribute)
      if (!attr) continue

      if (rule.attribute === 'gender') filter.gender = rule.value
      if (rule.attribute === 'age_from') filter.age_from = Number(rule.value)
      if (rule.attribute === 'age_to') filter.age_to = Number(rule.value)
      if (rule.attribute === 'min_spend') filter.min_spend = Number(rule.value)
      if (rule.attribute === 'max_spend') filter.max_spend = Number(rule.value)
      if (rule.attribute === 'min_visits') filter.min_visits = Number(rule.value)
      if (rule.attribute === 'max_visits') filter.max_visits = Number(rule.value)
      if (rule.attribute === 'rfm_segment') filter.rfm_category = rule.value
      if (rule.attribute === 'tags') filter.tags = rule.value.split(',').map((t) => t.trim()).filter(Boolean)
    }
    return filter
  }

  const handlePreview = async () => {
    setPreviewing(true)
    try {
      const result = await segmentsApi.previewCount(buildFilter())
      setPreviewCount(result.count)
    } catch {
      setPreviewCount(null)
    } finally {
      setPreviewing(false)
    }
  }

  const handleCreate = async () => {
    if (!name.trim()) return
    if (editingSegment) {
      await updateSegment.mutate({
        id: editingSegment.id,
        data: { name: name.trim(), filter: buildFilter() },
      })
    } else {
      await createSegment.mutate({
        name: name.trim(),
        type: 'custom',
        filter: buildFilter(),
        auto_assign: false,
      })
    }
    setShowBuilder(false)
    setEditingSegment(null)
    setName('')
    setRules([{ attribute: 'gender', operator: 'eq', value: '' }])
    setPreviewCount(null)
  }

  const handleEditSegment = (segment: Segment) => {
    setEditingSegment(segment)
    setName(segment.name)
    // Reconstruct rules from filter
    const filter = segment.filter ?? {}
    const reconstructed: FilterRule[] = []
    if (filter.gender) reconstructed.push({ attribute: 'gender', operator: 'eq', value: filter.gender as string })
    if (filter.age_from !== undefined) reconstructed.push({ attribute: 'age_from', operator: 'gte', value: String(filter.age_from) })
    if (filter.age_to !== undefined) reconstructed.push({ attribute: 'age_to', operator: 'lte', value: String(filter.age_to) })
    if (filter.rfm_category) reconstructed.push({ attribute: 'rfm_segment', operator: 'eq', value: filter.rfm_category as string })
    if (filter.tags && Array.isArray(filter.tags)) reconstructed.push({ attribute: 'tags', operator: 'eq', value: (filter.tags as string[]).join(',') })
    setRules(reconstructed.length > 0 ? reconstructed : [{ attribute: 'gender', operator: 'eq', value: '' }])
    setPreviewCount(null)
    setShowBuilder(true)
  }

  const handleDelete = async (id: number) => {
    await deleteSegment.mutate(id)
  }

  if (isLoading) {
    return (
      <div>
        <div className="h-8 w-48 shimmer rounded mb-4" />
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {[0, 1, 2].map((i) => <CardSkeleton key={i} />)}
        </div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-start justify-between mb-6 animate-in">
        <div>
          <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">
            Мои сегменты
          </h1>
          <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">
            Создавайте пользовательские сегменты по любым параметрам клиентов
          </p>
        </div>
        <Button
          variant="primary"
          leftIcon={<Plus className="w-4 h-4" />}
          onClick={() => { setEditingSegment(null); setShowBuilder(!showBuilder) }}
        >
          Создать сегмент
        </Button>
      </div>

      {/* Filter builder */}
      {showBuilder && (
        <div className="bg-white rounded border border-neutral-900 p-6 mb-6 animate-in">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Filter className="w-5 h-5 text-accent" />
              <h2 className="text-lg font-semibold text-neutral-900">
                {editingSegment ? 'Редактировать сегмент' : 'Новый сегмент'}
              </h2>
            </div>
            <button
              type="button"
              onClick={() => setShowBuilder(false)}
              className="text-neutral-400 hover:text-neutral-600 transition-colors"
            >
              <X className="w-5 h-5" />
            </button>
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-neutral-700 mb-1.5">Название</label>
            <input
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Например: VIP-клиенты из Москвы"
              className={cn(inputClassName, 'max-w-md')}
            />
          </div>

          <div className="space-y-3 mb-4">
            <p className="text-sm font-medium text-neutral-700">Фильтры</p>
            {rules.map((rule, index) => {
              const ruleAttr = FILTER_ATTRIBUTES.find((a) => a.key === rule.attribute)
              const isNonNumeric = ruleAttr?.type === 'select' || ruleAttr?.type === 'boolean'
              return (
              <div key={index} className="flex items-center gap-3 p-3 rounded bg-neutral-50">
                <CustomSelect
                  value={rule.attribute}
                  onChange={(v) => {
                    const newAttr = FILTER_ATTRIBUTES.find((a) => a.key === v)
                    const lockOp = newAttr?.type === 'select' || newAttr?.type === 'boolean'
                    setRules(rules.map((r, i) =>
                      i === index
                        ? { ...r, attribute: v, value: '', ...(lockOp ? { operator: 'eq' } : {}) }
                        : r,
                    ))
                  }}
                  options={FILTER_ATTRIBUTES.map((attr) => ({
                    value: attr.key,
                    label: attr.label,
                  }))}
                  width="200px"
                  light
                />

                <CustomSelect
                  value={isNonNumeric ? 'eq' : rule.operator}
                  onChange={(v) => updateRule(index, 'operator', v)}
                  options={isNonNumeric
                    ? [{ value: 'eq', label: '=' }]
                    : Object.entries(RULE_OPERATORS).map(([key, label]) => ({
                        value: key,
                        label: label,
                      }))
                  }
                  disabled={isNonNumeric}
                  width="120px"
                  light
                />

                {(() => {
                  const attr = FILTER_ATTRIBUTES.find((a) => a.key === rule.attribute)
                  if (attr?.type === 'select' && 'options' in attr) {
                    return (
                      <CustomSelect
                        className="flex-1"
                        value={rule.value}
                        onChange={(v) => updateRule(index, 'value', v)}
                        options={attr.options.map((opt) => ({
                          value: opt,
                          label: opt,
                        }))}
                        placeholder="Выберите..."
                        light
                      />
                    )
                  }
                  if (attr?.type === 'boolean') {
                    return (
                      <CustomSelect
                        className="flex-1"
                        value={rule.value}
                        onChange={(v) => updateRule(index, 'value', v)}
                        options={[
                          { value: 'true', label: 'Да' },
                          { value: 'false', label: 'Нет' },
                        ]}
                        placeholder="Выберите..."
                        light
                      />
                    )
                  }
                  return (
                    <input
                      type={attr?.type === 'number' ? 'number' : attr?.type === 'date' ? 'date' : 'text'}
                      value={rule.value}
                      onChange={(e) => updateRule(index, 'value', e.target.value)}
                      placeholder="Значение"
                      className={cn(inputClassName, 'flex-1')}
                    />
                  )
                })()}

                <button
                  type="button"
                  onClick={() => removeRule(index)}
                  disabled={rules.length <= 1}
                  className="p-2 rounded text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors disabled:opacity-30"
                >
                  <Trash2 className="w-4 h-4" />
                </button>
              </div>
              )
            })}

            <button
              type="button"
              onClick={addRule}
              className="flex items-center gap-1.5 text-sm text-accent hover:text-accent/80 font-medium transition-colors"
            >
              <Plus className="w-4 h-4" />
              Добавить фильтр
            </button>
          </div>

          <div className="flex items-center justify-between pt-4 border-t border-neutral-200">
            <div className="flex items-center gap-3">
              <button
                type="button"
                onClick={handlePreview}
                disabled={previewing}
                className={cn(
                  'flex items-center gap-1.5 py-2 px-4 rounded text-sm font-medium',
                  'border border-neutral-200 text-neutral-700',
                  'hover:bg-neutral-50 transition-colors',
                  'disabled:opacity-50',
                )}
              >
                <Eye className="w-4 h-4" />
                {previewing ? 'Подсчёт...' : 'Предпросмотр'}
              </button>
              {previewCount !== null && (
                <span className="text-sm text-neutral-500">
                  Найдено: <strong className="text-neutral-900">{previewCount}</strong> клиентов
                </span>
              )}
            </div>

            <Button
              variant="primary"
              onClick={handleCreate}
              disabled={(editingSegment ? updateSegment.isPending : createSegment.isPending) || !name.trim()}
            >
              {editingSegment
                ? (updateSegment.isPending ? 'Сохранение...' : 'Сохранить')
                : (createSegment.isPending ? 'Создание...' : 'Создать сегмент')}
            </Button>
          </div>
        </div>
      )}

      {/* Segment cards */}
      {customSegments.length === 0 && !showBuilder ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <Users className="w-8 h-8 text-neutral-400 mb-4" />
          <h3 className="font-display text-xl font-bold text-neutral-800 mb-1.5">
            Нет пользовательских сегментов
          </h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed mb-4">
            Создайте сегмент с нужными фильтрами для таргетированных рассылок и аналитики
          </p>
          <Button
            variant="primary"
            leftIcon={<Plus className="w-4 h-4" />}
            onClick={() => { setEditingSegment(null); setShowBuilder(true) }}
          >
            Создать первый сегмент
          </Button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {customSegments.map((segment: Segment) => (
            <SegmentCard
              key={segment.id}
              segment={segment}
              onDelete={handleDelete}
              onEdit={handleEditSegment}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function SegmentCard({
  segment,
  onDelete,
  onEdit,
}: {
  segment: Segment
  onDelete: (id: number) => void
  onEdit: (segment: Segment) => void
}) {
  const filterEntries = Object.entries(segment.filter ?? {}).filter(([, v]) => v !== undefined && v !== null)

  return (
    <div
      className={cn(
        'bg-white rounded border border-neutral-900 p-5',
        'hover:scale-[1.02] transition-transform duration-150 cursor-pointer',
      )}
      onClick={() => onEdit(segment)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && onEdit(segment)}
    >
      <div className="flex items-start justify-between mb-3">
        <h3 className="text-sm font-semibold text-neutral-900">{segment.name}</h3>
        <div className="flex items-center gap-1.5 shrink-0 ml-2">
          <span className="font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border bg-neutral-100 text-neutral-600 border-neutral-300">
            Custom
          </span>
          <button
            type="button"
            onClick={(e) => { e.stopPropagation(); onDelete(segment.id) }}
            className="p-1.5 rounded text-neutral-400 hover:text-red-500 hover:bg-red-50 transition-colors"
          >
            <Trash2 className="w-3.5 h-3.5" />
          </button>
        </div>
      </div>

      {filterEntries.length > 0 && (
        <div className="flex flex-wrap gap-1.5 mb-3">
          {filterEntries.map(([key, value]) => {
            const attr = FILTER_ATTRIBUTES.find((a) => a.key === key)
            return (
              <span
                key={key}
                className="text-[10px] font-medium px-2 py-0.5 rounded border bg-neutral-50 text-neutral-600 border-neutral-200"
              >
                {attr?.label ?? key}: {Array.isArray(value) ? value.join(', ') : String(value)}
              </span>
            )
          })}
        </div>
      )}

      <div className="mt-4 pt-4 border-t border-neutral-200 flex items-center justify-between text-neutral-400 font-mono text-[11px] uppercase tracking-wider tabular-nums">
        <span className="flex items-center gap-1.5">
          <Users className="w-3.5 h-3.5" />
          {(segment.client_count ?? 0).toLocaleString('ru-RU')} клиентов
        </span>
        <span>{new Date(segment.created_at).toLocaleDateString('ru-RU')}</span>
      </div>
    </div>
  )
}
