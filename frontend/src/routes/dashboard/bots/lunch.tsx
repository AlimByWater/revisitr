import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, Check, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBotQuery } from '@/features/bots/queries'
import { normalizeTimeInput } from '@/features/bots/settings'
import { lunchApi } from '@/features/lunch/api'
import { useLunchProgramQuery } from '@/features/lunch/queries'
import type {
  LunchAvailabilitySlot,
  LunchCourse,
  LunchFormat,
  LunchPriceMode,
  LunchProgram,
  SaveLunchCourseRequest,
  SaveLunchFormatRequest,
} from '@/features/lunch/types'
import { ordersApi } from '@/features/orders/api'
import { ORDER_STATUS_LABELS, ORDER_STATUS_STYLES } from '@/features/orders/labels'
import { useOrdersQuery } from '@/features/orders/queries'
import { menusApi } from '@/features/menus/api'
import { useMenuQuery, useMenusQuery } from '@/features/menus/queries'
import type { Menu, MenuCategory } from '@/features/menus/types'
import { CustomSelect } from '@/components/common/CustomSelect'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'

const inputClassName = cn(
  'w-full px-4 py-2.5 rounded border border-neutral-200',
  'text-sm placeholder:text-neutral-400 bg-white',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors disabled:opacity-50 disabled:cursor-not-allowed',
)

const WEEKDAYS = [
  { value: 1, label: 'Пн' },
  { value: 2, label: 'Вт' },
  { value: 3, label: 'Ср' },
  { value: 4, label: 'Чт' },
  { value: 5, label: 'Пт' },
  { value: 6, label: 'Сб' },
  { value: 7, label: 'Вс' },
]

const PRICE_MODES: { value: LunchPriceMode; label: string; description: string }[] = [
  { value: 'fixed', label: 'Фиксированная', description: 'Одна цена на формат, не зависит от выбранных позиций' },
  { value: 'sum_of_items', label: 'Сумма позиций', description: 'Цена заказа — сумма цен выбранных блюд из меню' },
  { value: 'base_plus_surcharge', label: 'База + доплаты', description: 'Базовая цена формата плюс доплаты за отдельные позиции' },
]

const priceFormatter = new Intl.NumberFormat('ru-RU', {
  style: 'currency',
  currency: 'RUB',
  maximumFractionDigits: 0,
})

function formatPrice(value: number): string {
  return priceFormatter.format(value)
}

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

function SectionCard({
  eyebrow,
  title,
  description,
  actions,
  children,
}: {
  eyebrow: string
  title: string
  description?: string
  actions?: React.ReactNode
  children: React.ReactNode
}) {
  return (
    <section className="rounded border border-neutral-900 bg-white p-4 sm:p-5 md:p-6">
      <div className="mb-5 flex items-start justify-between gap-3">
        <div>
          <p className="font-mono text-[10px] uppercase tracking-wider text-neutral-300 mb-1">{eyebrow}</p>
          <h3 className="text-sm font-semibold text-neutral-700">{title}</h3>
          {description && <p className="text-sm text-neutral-500 mt-1">{description}</p>}
        </div>
        {actions}
      </div>
      {children}
    </section>
  )
}

function Toggle({ checked, onChange }: { checked: boolean; onChange: () => void }) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      onClick={onChange}
      className={cn(
        'relative h-6 w-10 shrink-0 rounded-full transition-colors cursor-pointer',
        checked ? 'bg-accent' : 'bg-neutral-300',
      )}
    >
      <span
        className={cn(
          'absolute left-0.5 top-0.5 h-5 w-5 rounded-full bg-white shadow transition-transform',
          checked && 'translate-x-4',
        )}
      />
    </button>
  )
}

// ── Course editor ─────────────────────────────────────────────────────────

interface CourseDraft {
  code: string
  title: string
  menu_category_id: number
  items: Map<number, number> // menu_item_id -> surcharge
}

function courseDraftFrom(course?: LunchCourse): CourseDraft {
  const items = new Map<number, number>()
  for (const item of course?.items ?? []) {
    items.set(item.menu_item_id, item.surcharge)
  }
  return {
    code: course?.code ?? '',
    title: course?.title ?? '',
    menu_category_id: course?.menu_category_id ?? 0,
    items,
  }
}

function courseRequestFrom(draft: CourseDraft): SaveLunchCourseRequest {
  return {
    code: draft.code.trim(),
    title: draft.title.trim(),
    menu_category_id: draft.menu_category_id,
    items: Array.from(draft.items.entries()).map(([menu_item_id, surcharge]) => ({
      menu_item_id,
      surcharge,
    })),
  }
}

function CourseEditor({
  course,
  categories,
  onSave,
  onDelete,
  onCancel,
}: {
  course?: LunchCourse
  categories: MenuCategory[]
  onSave: (data: SaveLunchCourseRequest) => Promise<void>
  onDelete?: () => Promise<void>
  onCancel?: () => void
}) {
  const [draft, setDraft] = useState<CourseDraft>(() => courseDraftFrom(course))
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const category = categories.find((c) => c.id === draft.menu_category_id)
  const categoryItems = category?.items ?? []

  const validationError = useMemo(() => {
    if (!draft.title.trim()) return 'Укажите название курса'
    if (!draft.code.trim()) return 'Укажите код курса (например, 1)'
    if (!draft.menu_category_id) return 'Выберите категорию меню'
    if (draft.items.size === 0) return 'Отметьте хотя бы одну позицию'
    return null
  }, [draft])

  const save = async () => {
    setIsSaving(true)
    setError(null)
    try {
      await onSave(courseRequestFrom(draft))
    } catch {
      setError('Не удалось сохранить курс. Попробуйте ещё раз.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="rounded border border-neutral-200 bg-neutral-50/70 p-4 space-y-4">
      <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_100px_minmax(0,1fr)]">
        <div className="space-y-2">
          <label className="block text-sm font-medium text-neutral-700">Название</label>
          <input
            type="text"
            value={draft.title}
            onChange={(event) => setDraft((current) => ({ ...current, title: event.target.value }))}
            className={inputClassName}
            placeholder="Первое"
          />
        </div>
        <div className="space-y-2">
          <label className="block text-sm font-medium text-neutral-700">Код</label>
          <input
            type="text"
            value={draft.code}
            onChange={(event) => setDraft((current) => ({ ...current, code: event.target.value }))}
            className={inputClassName}
            placeholder="1"
          />
        </div>
        <div className="space-y-2">
          <label className="block text-sm font-medium text-neutral-700">Категория меню</label>
          <CustomSelect
            value={draft.menu_category_id ? String(draft.menu_category_id) : ''}
            onChange={(value) =>
              setDraft((current) => ({ ...current, menu_category_id: Number(value), items: new Map() }))
            }
            options={categories.map((c) => ({ value: String(c.id), label: c.name }))}
            placeholder="Выберите категорию"
            light
          />
        </div>
      </div>

      {draft.menu_category_id > 0 && (
        <div className="space-y-2">
          <p className="text-sm font-medium text-neutral-700">Позиции курса</p>
          {categoryItems.length === 0 ? (
            <p className="text-sm text-neutral-500">В этой категории нет позиций.</p>
          ) : (
            categoryItems.map((item) => {
              const isChecked = draft.items.has(item.id)
              const surcharge = draft.items.get(item.id) ?? 0
              return (
                <div key={item.id} className="flex flex-wrap items-center gap-3 rounded border border-neutral-200 bg-white px-3 py-2.5">
                  <button
                    type="button"
                    role="switch"
                    aria-checked={isChecked}
                    onClick={() =>
                      setDraft((current) => {
                        const items = new Map(current.items)
                        if (isChecked) items.delete(item.id)
                        else items.set(item.id, 0)
                        return { ...current, items }
                      })
                    }
                    className={cn(
                      'flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors',
                      isChecked ? 'border-accent bg-accent' : 'border-neutral-300 bg-white',
                    )}
                  >
                    {isChecked && <Check className="h-3 w-3 text-white" />}
                  </button>
                  <div className="min-w-0 flex-1">
                    <span className="text-sm text-neutral-900">{item.name}</span>
                    <span className="ml-2 text-xs text-neutral-500">{formatPrice(item.price)}</span>
                  </div>
                  {isChecked && (
                    <div className="flex items-center gap-1.5">
                      <span className="text-xs text-neutral-500">Доплата</span>
                      <input
                        type="text"
                        inputMode="numeric"
                        value={surcharge === 0 ? '' : String(surcharge)}
                        onChange={(event) =>
                          setDraft((current) => {
                            const items = new Map(current.items)
                            items.set(item.id, Number(event.target.value.replace(/\D/g, '')) || 0)
                            return { ...current, items }
                          })
                        }
                        className={cn(inputClassName, 'w-24 px-2 py-1.5')}
                        placeholder="0"
                      />
                      <span className="text-xs text-neutral-500">₽</span>
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>
      )}

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-neutral-200 pt-3">
        <div className="text-sm">
          {error && <span className="text-red-600">{error}</span>}
          {!error && validationError && <span className="text-neutral-400">{validationError}</span>}
        </div>
        <div className="flex items-center gap-2">
          {onDelete && (
            <button
              type="button"
              onClick={onDelete}
              className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-600"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="inline-flex min-h-11 items-center rounded px-4 text-sm text-neutral-500 hover:text-neutral-700"
            >
              Отмена
            </button>
          )}
          <button
            type="button"
            onClick={save}
            disabled={isSaving || Boolean(validationError)}
            className={cn(
              'inline-flex min-h-11 items-center justify-center rounded px-4 text-sm font-medium',
              'bg-accent text-white hover:bg-accent/90 transition-colors',
              'disabled:cursor-not-allowed disabled:opacity-50',
            )}
          >
            {isSaving ? 'Сохранение...' : 'Сохранить курс'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Format editor ─────────────────────────────────────────────────────────

interface FormatDraft {
  name: string
  price_mode: LunchPriceMode
  base_price: string
  is_active: boolean
  course_ids: number[]
}

function formatDraftFrom(format?: LunchFormat): FormatDraft {
  return {
    name: format?.name ?? '',
    price_mode: format?.price_mode ?? 'fixed',
    base_price: format && format.base_price > 0 ? String(format.base_price) : '',
    is_active: format?.is_active ?? true,
    course_ids: format?.course_ids ?? [],
  }
}

function formatRequestFrom(draft: FormatDraft): SaveLunchFormatRequest {
  return {
    name: draft.name.trim(),
    price_mode: draft.price_mode,
    base_price: Number(draft.base_price) || 0,
    is_active: draft.is_active,
    course_ids: draft.course_ids,
  }
}

function FormatEditor({
  format,
  courses,
  onSave,
  onDelete,
  onCancel,
}: {
  format?: LunchFormat
  courses: LunchCourse[]
  onSave: (data: SaveLunchFormatRequest) => Promise<void>
  onDelete?: () => Promise<void>
  onCancel?: () => void
}) {
  const [draft, setDraft] = useState<FormatDraft>(() => formatDraftFrom(format))
  const [isSaving, setIsSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const needsBasePrice = draft.price_mode !== 'sum_of_items'

  const validationError = useMemo(() => {
    if (!draft.name.trim()) return 'Укажите название формата'
    if (draft.course_ids.length === 0) return 'Выберите хотя бы один курс'
    if (draft.price_mode === 'fixed' && !(Number(draft.base_price) > 0)) return 'Для фиксированной цены укажите сумму больше 0'
    const emptyCourse = draft.course_ids
      .map((courseId) => courses.find((c) => c.id === courseId))
      .find((course) => course && course.items.length === 0)
    if (emptyCourse) return `Курс «${emptyCourse.title}» пуст — отметьте в нём позиции`
    return null
  }, [draft, courses])

  const save = async () => {
    setIsSaving(true)
    setError(null)
    try {
      await onSave(formatRequestFrom(draft))
    } catch {
      setError('Не удалось сохранить формат. Попробуйте ещё раз.')
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <div className="rounded border border-neutral-200 bg-neutral-50/70 p-4 space-y-4">
      <div className="flex flex-wrap items-end gap-3">
        <div className="min-w-56 flex-1 space-y-2">
          <label className="block text-sm font-medium text-neutral-700">Название</label>
          <input
            type="text"
            value={draft.name}
            onChange={(event) => setDraft((current) => ({ ...current, name: event.target.value }))}
            className={inputClassName}
            placeholder="Первое + Второе + Напиток"
          />
        </div>
        <div className="flex min-h-11 items-center gap-2 pb-1">
          <span className="text-sm text-neutral-500">Активен</span>
          <Toggle
            checked={draft.is_active}
            onChange={() => setDraft((current) => ({ ...current, is_active: !current.is_active }))}
          />
        </div>
      </div>

      <div className="space-y-2">
        <p className="text-sm font-medium text-neutral-700">Модель цены</p>
        <div className="grid gap-2 sm:grid-cols-3">
          {PRICE_MODES.map((mode) => {
            const isSelected = draft.price_mode === mode.value
            return (
              <button
                key={mode.value}
                type="button"
                onClick={() => setDraft((current) => ({ ...current, price_mode: mode.value }))}
                className={cn(
                  'rounded border p-3 text-left transition-colors',
                  isSelected ? 'border-accent bg-accent/5' : 'border-neutral-200 bg-white',
                )}
              >
                <span className="block text-sm font-medium text-neutral-900">{mode.label}</span>
                <span className="mt-1 block text-xs text-neutral-500">{mode.description}</span>
              </button>
            )
          })}
        </div>
      </div>

      {needsBasePrice && (
        <div className="max-w-56 space-y-2">
          <label className="block text-sm font-medium text-neutral-700">
            {draft.price_mode === 'fixed' ? 'Цена формата' : 'Базовая цена'}
          </label>
          <div className="flex items-stretch">
            <input
              type="text"
              inputMode="numeric"
              value={draft.base_price}
              onChange={(event) =>
                setDraft((current) => ({ ...current, base_price: event.target.value.replace(/\D/g, '') }))
              }
              className={cn(inputClassName, 'rounded-r-none')}
              placeholder="350"
            />
            <span className="inline-flex items-center rounded-r border border-l-0 border-neutral-200 bg-neutral-50 px-3 text-sm text-neutral-500">
              ₽
            </span>
          </div>
        </div>
      )}

      <div className="space-y-2">
        <p className="text-sm font-medium text-neutral-700">Курсы формата</p>
        <p className="text-xs text-neutral-500">Порядок выбора = порядок шагов в боте. Снимите и отметьте заново, чтобы изменить порядок.</p>
        {courses.length === 0 ? (
          <p className="text-sm text-neutral-500">Сначала создайте курсы.</p>
        ) : (
          courses.map((course) => {
            const position = draft.course_ids.indexOf(course.id)
            const isChecked = position >= 0
            return (
              <div key={course.id} className="flex items-center gap-3 rounded border border-neutral-200 bg-white px-3 py-2.5">
                <button
                  type="button"
                  role="switch"
                  aria-checked={isChecked}
                  onClick={() =>
                    setDraft((current) => ({
                      ...current,
                      course_ids: isChecked
                        ? current.course_ids.filter((id) => id !== course.id)
                        : [...current.course_ids, course.id],
                    }))
                  }
                  className={cn(
                    'flex h-4 w-4 shrink-0 items-center justify-center rounded border transition-colors',
                    isChecked ? 'border-accent bg-accent' : 'border-neutral-300 bg-white',
                  )}
                >
                  {isChecked && <Check className="h-3 w-3 text-white" />}
                </button>
                <span className="min-w-0 flex-1 text-sm text-neutral-900">
                  {course.title}
                  <span className="ml-2 text-xs text-neutral-500">{course.items.length} поз.</span>
                </span>
                {isChecked && (
                  <span className="font-mono text-xs text-neutral-400">шаг {position + 1}</span>
                )}
              </div>
            )
          })
        )}
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3 border-t border-neutral-200 pt-3">
        <div className="text-sm">
          {error && <span className="text-red-600">{error}</span>}
          {!error && validationError && <span className="text-neutral-400">{validationError}</span>}
        </div>
        <div className="flex items-center gap-2">
          {onDelete && (
            <button
              type="button"
              onClick={onDelete}
              className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-600"
            >
              <Trash2 className="h-4 w-4" />
            </button>
          )}
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              className="inline-flex min-h-11 items-center rounded px-4 text-sm text-neutral-500 hover:text-neutral-700"
            >
              Отмена
            </button>
          )}
          <button
            type="button"
            onClick={save}
            disabled={isSaving || Boolean(validationError)}
            className={cn(
              'inline-flex min-h-11 items-center justify-center rounded px-4 text-sm font-medium',
              'bg-accent text-white hover:bg-accent/90 transition-colors',
              'disabled:cursor-not-allowed disabled:opacity-50',
            )}
          >
            {isSaving ? 'Сохранение...' : 'Сохранить формат'}
          </button>
        </div>
      </div>
    </div>
  )
}

// ── Orders section ────────────────────────────────────────────────────────

function OrdersSection({ botId }: { botId: number }) {
  const [showAll, setShowAll] = useState(false)
  const { data: orders = [], isError, mutate } = useOrdersQuery(botId, 'lunch', showAll ? undefined : 'new')
  const [busyOrderId, setBusyOrderId] = useState<number | null>(null)

  const changeStatus = async (orderId: number, status: string) => {
    setBusyOrderId(orderId)
    try {
      await ordersApi.updateOrderStatus(orderId, status)
      await mutate()
    } finally {
      setBusyOrderId(null)
    }
  }

  return (
    <SectionCard
      eyebrow="Заказы"
      title={showAll ? 'Все заказы' : 'Новые заказы'}
      description="Гость оформил ланч в боте — заказ появляется здесь. Список обновляется автоматически."
      actions={
        <button
          type="button"
          onClick={() => setShowAll((current) => !current)}
          className="inline-flex min-h-11 shrink-0 items-center rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors"
        >
          {showAll ? 'Только новые' : 'Показать все'}
        </button>
      }
    >
      <div className="space-y-2">
        {isError && (
          <p className="text-sm text-red-600">Не удалось загрузить заказы. Обновите страницу или войдите заново.</p>
        )}
        {!isError && orders.length === 0 && (
          <p className="text-sm text-neutral-500">{showAll ? 'Заказов пока нет.' : 'Новых заказов нет.'}</p>
        )}
        {orders.map((order) => (
          <div key={order.id} className="rounded border border-neutral-200 bg-white px-3 py-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="min-w-0">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="text-sm font-medium text-neutral-900">
                    №{order.id} · Стол {order.table_num} · {formatPrice(order.total_price)}
                  </span>
                  <span className={cn('rounded px-1.5 py-0.5 text-xs font-medium', ORDER_STATUS_STYLES[order.status])}>
                    {ORDER_STATUS_LABELS[order.status]}
                  </span>
                </div>
                <div className="mt-1 text-xs text-neutral-500">
                  {order.format_name} · {new Date(order.created_at).toLocaleString('ru-RU')}
                </div>
                <div className="mt-1 text-xs text-neutral-500">
                  {order.items.map((item) => `${item.course_title}: ${item.item_name}`).join(' · ')}
                </div>
              </div>
              {order.status === 'new' && (
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    disabled={busyOrderId === order.id}
                    onClick={() => changeStatus(order.id, 'sent')}
                    className={cn(
                      'inline-flex min-h-11 items-center rounded px-3 text-sm font-medium',
                      'bg-accent text-white hover:bg-accent/90 transition-colors',
                      'disabled:cursor-not-allowed disabled:opacity-50',
                    )}
                  >
                    Отработан
                  </button>
                  <button
                    type="button"
                    disabled={busyOrderId === order.id}
                    onClick={() => changeStatus(order.id, 'cancelled')}
                    className="inline-flex min-h-11 items-center rounded px-3 text-sm text-neutral-500 hover:bg-red-50 hover:text-red-600 transition-colors disabled:opacity-50"
                  >
                    Отменить
                  </button>
                </div>
              )}
            </div>
          </div>
        ))}
      </div>
    </SectionCard>
  )
}

// ── Page ──────────────────────────────────────────────────────────────────

interface ProgramDraft {
  name: string
  description: string
  is_active: boolean
  availability: LunchAvailabilitySlot[]
}

export default function BotLunchSettingsPage() {
  const { botId } = useParams<{ botId: string }>()
  const id = Number(botId)
  const safeId = Number.isNaN(id) ? 0 : id

  const { data: bot, isLoading: botLoading, isError: botError } = useBotQuery(safeId)
  const { data: program, isError: programError, mutate: mutateProgram } = useLunchProgramQuery(safeId)
  const { data: menus = [] } = useMenusQuery()

  const [boundPosIds, setBoundPosIds] = useState<number[]>([])
  const [draft, setDraft] = useState<ProgramDraft | null>(null)
  const [addingCourse, setAddingCourse] = useState(false)
  const [addingFormat, setAddingFormat] = useState(false)

  useEffect(() => {
    if (!safeId) return
    let mounted = true
    menusApi
      .getBotPOSLocations(safeId)
      .then((response) => mounted && setBoundPosIds(response.pos_ids ?? []))
      .catch(() => mounted && setBoundPosIds([]))
    return () => {
      mounted = false
    }
  }, [safeId])

  const botMenu = useMemo(() => pickBotMenu(menus, boundPosIds), [menus, boundPosIds])
  const { data: fullMenu } = useMenuQuery(botMenu?.id ?? 0)
  const categories = useMemo(() => fullMenu?.categories ?? [], [fullMenu])

  useEffect(() => {
    if (!program) return
    setDraft({
      name: program.name,
      description: program.description,
      is_active: program.is_active,
      availability: program.availability ?? [],
    })
  }, [program])

  const [isSavingProgram, setIsSavingProgram] = useState(false)
  const [programSaveError, setProgramSaveError] = useState<string | null>(null)
  const [programSaveSuccess, setProgramSaveSuccess] = useState(false)

  useEffect(() => {
    if (!programSaveSuccess) return
    const t = setTimeout(() => setProgramSaveSuccess(false), 3000)
    return () => clearTimeout(t)
  }, [programSaveSuccess])

  const saveProgram = async () => {
    if (!draft) return
    setIsSavingProgram(true)
    setProgramSaveError(null)
    setProgramSaveSuccess(false)
    try {
      await lunchApi.updateProgram(safeId, {
        name: draft.name,
        description: draft.description,
        is_active: draft.is_active,
      })
      await lunchApi.setAvailability(safeId, draft.availability)
      await mutateProgram()
      setProgramSaveSuccess(true)
    } catch {
      setProgramSaveError('Не удалось сохранить. Проверьте поля и попробуйте ещё раз.')
    } finally {
      setIsSavingProgram(false)
    }
  }

  const afterMutation = async () => {
    await mutateProgram()
  }

  if (botLoading || !draft || !program) {
    if (Number.isNaN(id) || botError || programError) {
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
      <div className="max-w-4xl">
        <div className="mb-6 h-4 w-32 shimmer rounded" />
        <CardSkeleton />
      </div>
    )
  }

  const courses = program.courses ?? []
  const formats = program.formats ?? []

  return (
    <div className="max-w-4xl">
      <div className="animate-in mb-6">
        <Link
          to={`/dashboard/bots/${safeId}?tab=modules`}
          className="mb-4 inline-flex min-h-11 items-center gap-1.5 rounded text-sm text-neutral-500 transition-colors hover:text-neutral-700"
        >
          <ArrowLeft className="h-4 w-4" />
          К модулям
        </Link>
        <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Бизнес-ланч</h1>
        <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">Настройки модуля</p>
        <p className="mt-2 max-w-2xl text-sm text-neutral-500">
          Курсы, форматы комбо, модели цены и окно доступности. Гость собирает обед в боте
          {bot ? ` @${bot.username}` : ''} и указывает номер стола.
        </p>
      </div>

      <div className="space-y-6">
        <OrdersSection botId={safeId} />

        <SectionCard
          eyebrow="Программа"
          title="Название и статус"
          actions={
            <div className="flex min-h-11 items-center gap-2">
              <span className="text-sm text-neutral-500">Активна</span>
              <Toggle
                checked={draft.is_active}
                onChange={() => setDraft((current) => (current ? { ...current, is_active: !current.is_active } : current))}
              />
            </div>
          }
        >
          <div className="space-y-3">
            <div className="space-y-2">
              <label className="block text-sm font-medium text-neutral-700">Название</label>
              <input
                type="text"
                value={draft.name}
                onChange={(event) => setDraft((current) => (current ? { ...current, name: event.target.value } : current))}
                className={inputClassName}
                placeholder="Бизнес-ланч"
              />
            </div>
            <div className="space-y-2">
              <label className="block text-sm font-medium text-neutral-700">Описание</label>
              <textarea
                value={draft.description}
                onChange={(event) =>
                  setDraft((current) => (current ? { ...current, description: event.target.value } : current))
                }
                className={cn(inputClassName, 'min-h-20 resize-y')}
                placeholder="Собери свой обед: с 12:00 до 16:00 по будням"
              />
            </div>
          </div>
        </SectionCard>

        <SectionCard
          eyebrow="Расписание"
          title="Окно доступности"
          description="Когда пункт «Ланч» виден в боте. Серверное время."
          actions={
            <button
              type="button"
              onClick={() =>
                setDraft((current) =>
                  current
                    ? {
                        ...current,
                        availability: [...current.availability, { weekday: 1, time_from: '12:00', time_to: '16:00' }],
                      }
                    : current,
                )
              }
              className="inline-flex min-h-11 shrink-0 items-center gap-1.5 rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors"
            >
              + Добавить слот
            </button>
          }
        >
          <div className="space-y-2">
            {draft.availability.length === 0 && (
              <p className="text-sm text-neutral-500">Слотов нет — ланч не будет виден гостям.</p>
            )}
            {draft.availability.map((slot, index) => (
              <div key={index} className="grid grid-cols-[110px_minmax(0,1fr)_minmax(0,1fr)_auto] gap-2">
                <CustomSelect
                  value={String(slot.weekday)}
                  onChange={(value) =>
                    setDraft((current) => {
                      if (!current) return current
                      const next = [...current.availability]
                      next[index] = { ...slot, weekday: Number(value) }
                      return { ...current, availability: next }
                    })
                  }
                  options={WEEKDAYS.map((day) => ({ value: String(day.value), label: day.label }))}
                  light
                />
                <input
                  type="text"
                  value={slot.time_from}
                  onChange={(event) =>
                    setDraft((current) => {
                      if (!current) return current
                      const next = [...current.availability]
                      next[index] = { ...slot, time_from: normalizeTimeInput(event.target.value) }
                      return { ...current, availability: next }
                    })
                  }
                  className={inputClassName}
                  placeholder="12:00"
                />
                <input
                  type="text"
                  value={slot.time_to}
                  onChange={(event) =>
                    setDraft((current) => {
                      if (!current) return current
                      const next = [...current.availability]
                      next[index] = { ...slot, time_to: normalizeTimeInput(event.target.value) }
                      return { ...current, availability: next }
                    })
                  }
                  className={inputClassName}
                  placeholder="16:00"
                />
                <button
                  type="button"
                  onClick={() =>
                    setDraft((current) =>
                      current
                        ? { ...current, availability: current.availability.filter((_, slotIndex) => slotIndex !== index) }
                        : current,
                    )
                  }
                  className="inline-flex min-h-11 min-w-11 items-center justify-center rounded text-neutral-400 hover:bg-red-50 hover:text-red-600"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>

          <div className="mt-6 flex flex-col gap-3 border-t border-neutral-200 pt-5 sm:flex-row sm:items-center sm:justify-between">
            <div className="min-h-5 text-sm">
              {programSaveError && <span className="text-red-600">{programSaveError}</span>}
              {programSaveSuccess && <span className="text-green-600">Сохранено</span>}
            </div>
            <button
              type="button"
              onClick={saveProgram}
              disabled={isSavingProgram}
              className={cn(
                'inline-flex min-h-11 items-center justify-center rounded px-5 text-sm font-medium',
                'bg-accent text-white hover:bg-accent/90 transition-colors',
                'disabled:cursor-not-allowed disabled:opacity-50',
              )}
            >
              {isSavingProgram ? 'Сохранение...' : 'Сохранить программу'}
            </button>
          </div>
        </SectionCard>

        <SectionCard
          eyebrow="Курсы"
          title="Блюда-курсы"
          description="Курс — одно «блюдо» комбо: категория меню и отмеченные позиции из неё."
          actions={
            !addingCourse ? (
              <button
                type="button"
                onClick={() => setAddingCourse(true)}
                className="inline-flex min-h-11 shrink-0 items-center gap-1.5 rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors"
              >
                + Добавить курс
              </button>
            ) : undefined
          }
        >
          <div className="space-y-3">
            {courses.length === 0 && !addingCourse && (
              <p className="text-sm text-neutral-500">Курсов пока нет. Начните с «Первое», «Второе», «Напиток».</p>
            )}
            {courses.map((course) => (
              <div key={course.id}>
                {course.items.length === 0 && (
                  <p className="mb-1 text-xs text-red-600">
                    В курсе не осталось позиций — форматы с ним скрыты в боте.
                  </p>
                )}
                <CourseEditor
                  course={course}
                  categories={categories}
                  onSave={async (data) => {
                    await lunchApi.updateCourse(course.id, data)
                    await afterMutation()
                  }}
                  onDelete={async () => {
                    await lunchApi.deleteCourse(course.id)
                    await afterMutation()
                  }}
                />
              </div>
            ))}
            {addingCourse && (
              <CourseEditor
                categories={categories}
                onSave={async (data) => {
                  await lunchApi.createCourse(safeId, data)
                  setAddingCourse(false)
                  await afterMutation()
                }}
                onCancel={() => setAddingCourse(false)}
              />
            )}
          </div>
        </SectionCard>

        <SectionCard
          eyebrow="Форматы"
          title="Комбо-форматы"
          description="Формат — комбинация курсов с моделью цены. Гость выбирает формат первым."
          actions={
            !addingFormat ? (
              <button
                type="button"
                onClick={() => setAddingFormat(true)}
                className="inline-flex min-h-11 shrink-0 items-center gap-1.5 rounded text-sm font-medium text-accent hover:text-accent/80 transition-colors"
              >
                + Добавить формат
              </button>
            ) : undefined
          }
        >
          <div className="space-y-3">
            {formats.length === 0 && !addingFormat && (
              <p className="text-sm text-neutral-500">Форматов пока нет. Например: «Первое + Второе + Напиток» за 350 ₽.</p>
            )}
            {formats.map((format) => (
              <FormatEditor
                key={format.id}
                format={format}
                courses={courses}
                onSave={async (data) => {
                  await lunchApi.updateFormat(format.id, data)
                  await afterMutation()
                }}
                onDelete={async () => {
                  await lunchApi.deleteFormat(format.id)
                  await afterMutation()
                }}
              />
            ))}
            {addingFormat && (
              <FormatEditor
                courses={courses}
                onSave={async (data) => {
                  await lunchApi.createFormat(safeId, data)
                  setAddingFormat(false)
                  await afterMutation()
                }}
                onCancel={() => setAddingFormat(false)}
              />
            )}
          </div>
        </SectionCard>
      </div>
    </div>
  )
}
