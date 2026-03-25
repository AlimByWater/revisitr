import { useState } from 'react'
import { cn, getApiErrorMessage } from '@/lib/utils'
import {
  usePromoCodesQuery,
  usePromotionsQuery,
  useCreatePromoCodeMutation,
  useDeactivatePromoCodeMutation,
  useGenerateCodeMutation,
} from '@/features/promotions/queries'
import type { PromoCode, CreatePromoCodeRequest } from '@/features/promotions/types'
import { Ticket, Plus, X, Ban, Copy, Sparkles } from 'lucide-react'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'

function getCodeStatus(code: PromoCode): { label: string; className: string } {
  if (!code.active) {
    return { label: 'Деактивирован', className: 'bg-neutral-100 text-neutral-500' }
  }
  if (code.usage_limit && code.usage_count >= code.usage_limit) {
    return { label: 'Исчерпан', className: 'bg-amber-100 text-amber-700' }
  }
  if (code.ends_at && new Date(code.ends_at) < new Date()) {
    return { label: 'Истёк', className: 'bg-red-100 text-red-700' }
  }
  return { label: 'Активен', className: 'bg-green-100 text-green-700' }
}

function CreatePromoCodeModal({ onClose }: { onClose: () => void }) {
  const [code, setCode] = useState('')
  const [promotionId, setPromotionId] = useState<string>('')
  const [discountPercent, setDiscountPercent] = useState<string>('')
  const [bonusAmount, setBonusAmount] = useState<string>('')
  const [usageLimit, setUsageLimit] = useState<string>('')
  const [perUserLimit, setPerUserLimit] = useState<string>('')
  const [channel, setChannel] = useState<string>('')
  const [description, setDescription] = useState('')
  const [startsAt, setStartsAt] = useState('')
  const [endsAt, setEndsAt] = useState('')
  const [minAmount, setMinAmount] = useState<string>('')

  const { data: promotions } = usePromotionsQuery()
  const createCode = useCreatePromoCodeMutation()
  const generateCode = useGenerateCodeMutation()

  const handleGenerate = async () => {
    try {
      const generated = await generateCode.mutateAsync(undefined as never)
      setCode(generated)
    } catch {
      // error handled silently
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    const data: CreatePromoCodeRequest = {
      ...(code ? { code } : {}),
      ...(promotionId ? { promotion_id: Number(promotionId) } : {}),
      ...(discountPercent ? { discount_percent: Number(discountPercent) } : {}),
      ...(bonusAmount ? { bonus_amount: Number(bonusAmount) } : {}),
      ...(usageLimit ? { usage_limit: Number(usageLimit) } : {}),
      ...(perUserLimit ? { per_user_limit: Number(perUserLimit) } : {}),
      ...(channel ? { channel } : {}),
      ...(description ? { description } : {}),
      ...(startsAt ? { starts_at: new Date(startsAt).toISOString() } : {}),
      ...(endsAt ? { ends_at: new Date(endsAt).toISOString() } : {}),
      ...(minAmount ? { conditions: { min_amount: Number(minAmount) } } : {}),
    }

    try {
      await createCode.mutateAsync(data)
      onClose()
    } catch {
      // error is available via createCode.error
    }
  }

  const inputClass = cn(
    'w-full px-4 py-2.5 rounded-lg border border-surface-border',
    'text-sm placeholder:text-neutral-400',
    'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
    'transition-colors',
    'disabled:opacity-50 disabled:cursor-not-allowed',
  )

  return (
    <div
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose()
      }}
      role="dialog"
      aria-modal="true"
      aria-labelledby="create-code-title"
    >
      <div className="bg-white rounded-2xl p-6 w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto">
        <div className="flex items-center justify-between mb-6">
          <h2
            id="create-code-title"
            className="text-lg font-semibold text-neutral-900"
          >
            Создать промокод
          </h2>
          <button
            onClick={onClose}
            type="button"
            className="p-1 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
            aria-label="Закрыть"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label
              htmlFor="code-value"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Код{' '}
              <span className="text-neutral-400 font-normal">
                (оставьте пустым для автогенерации)
              </span>
            </label>
            <div className="flex gap-2">
              <input
                id="code-value"
                type="text"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
                placeholder="SALE20"
                maxLength={32}
                disabled={createCode.isPending}
                className={inputClass}
              />
              <button
                type="button"
                onClick={handleGenerate}
                disabled={generateCode.isPending || createCode.isPending}
                className={cn(
                  'px-3 py-2.5 rounded-lg border border-surface-border',
                  'text-sm text-neutral-600 hover:bg-neutral-50',
                  'transition-colors flex-shrink-0',
                  'disabled:opacity-50 disabled:cursor-not-allowed',
                )}
                title="Сгенерировать код"
                aria-label="Сгенерировать код"
              >
                <Sparkles className="w-4 h-4" />
              </button>
            </div>
          </div>

          <div>
            <label
              htmlFor="code-promotion"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Привязка к акции{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <select
              id="code-promotion"
              value={promotionId}
              onChange={(e) => setPromotionId(e.target.value)}
              disabled={createCode.isPending}
              className={inputClass}
            >
              <option value="">Без привязки</option>
              {promotions
                ?.filter((p) => p.active)
                .map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label
                htmlFor="code-discount"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Скидка %
              </label>
              <input
                id="code-discount"
                type="number"
                value={discountPercent}
                onChange={(e) => setDiscountPercent(e.target.value)}
                placeholder="10"
                min={1}
                max={100}
                disabled={createCode.isPending}
                className={inputClass}
              />
            </div>
            <div>
              <label
                htmlFor="code-bonus"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Бонус
              </label>
              <input
                id="code-bonus"
                type="number"
                value={bonusAmount}
                onChange={(e) => setBonusAmount(e.target.value)}
                placeholder="500"
                min={1}
                disabled={createCode.isPending}
                className={inputClass}
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="code-channel"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Канал{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <select
              id="code-channel"
              value={channel}
              onChange={(e) => setChannel(e.target.value)}
              disabled={createCode.isPending}
              className={inputClass}
            >
              <option value="">Не указан</option>
              <option value="smm">SMM</option>
              <option value="targeting">Таргетинг</option>
              <option value="yandex_maps">Яндекс Карты</option>
              <option value="flyer">Флаер</option>
              <option value="partner">Партнёр</option>
              <option value="custom">Другой</option>
            </select>
          </div>

          <div>
            <label
              htmlFor="code-description"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Описание{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <input
              id="code-description"
              type="text"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Промокод для Instagram"
              maxLength={200}
              disabled={createCode.isPending}
              className={inputClass}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label
                htmlFor="code-usage-limit"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Лимит использований
              </label>
              <input
                id="code-usage-limit"
                type="number"
                value={usageLimit}
                onChange={(e) => setUsageLimit(e.target.value)}
                placeholder="100"
                min={1}
                disabled={createCode.isPending}
                className={inputClass}
              />
            </div>
            <div>
              <label
                htmlFor="code-per-user-limit"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Лимит на клиента
              </label>
              <input
                id="code-per-user-limit"
                type="number"
                value={perUserLimit}
                onChange={(e) => setPerUserLimit(e.target.value)}
                placeholder="1"
                min={1}
                disabled={createCode.isPending}
                className={inputClass}
              />
            </div>
          </div>

          <div>
            <label
              htmlFor="code-min-amount"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Мин. сумма заказа{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <input
              id="code-min-amount"
              type="number"
              value={minAmount}
              onChange={(e) => setMinAmount(e.target.value)}
              placeholder="1000"
              min={0}
              disabled={createCode.isPending}
              className={inputClass}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label
                htmlFor="code-starts-at"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Дата начала
              </label>
              <input
                id="code-starts-at"
                type="date"
                value={startsAt}
                onChange={(e) => setStartsAt(e.target.value)}
                disabled={createCode.isPending}
                className={inputClass}
              />
            </div>
            <div>
              <label
                htmlFor="code-ends-at"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Дата окончания
              </label>
              <input
                id="code-ends-at"
                type="date"
                value={endsAt}
                onChange={(e) => setEndsAt(e.target.value)}
                disabled={createCode.isPending}
                className={inputClass}
              />
            </div>
          </div>

          {createCode.isError && (
            <p className="text-sm text-red-600">
              {getApiErrorMessage(
                createCode.error,
                'Не удалось создать промокод. Попробуйте снова.',
              )}
            </p>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              disabled={createCode.isPending}
              className={cn(
                'flex-1 py-2.5 px-4 rounded-lg',
                'border border-surface-border text-sm font-medium text-neutral-700',
                'hover:bg-neutral-50 active:bg-neutral-100',
                'transition-colors',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={createCode.isPending}
              className={cn(
                'flex-1 py-2.5 px-4 rounded-lg',
                'bg-accent text-white text-sm font-medium',
                'hover:bg-accent/90 active:bg-accent/80',
                'transition-colors',
                'focus:outline-none focus:ring-2 focus:ring-accent/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {createCode.isPending ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

const channelLabels: Record<string, string> = {
  smm: 'SMM',
  targeting: 'Таргетинг',
  yandex_maps: 'Яндекс Карты',
  flyer: 'Флаер',
  partner: 'Партнёр',
  custom: 'Другой',
}

export default function PromoCodesPage() {
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [filterPromotionId, setFilterPromotionId] = useState<string>('')
  const { data: codes, isLoading, isError, mutate } = usePromoCodesQuery()
  const { data: promotions } = usePromotionsQuery()
  const deactivateCode = useDeactivatePromoCodeMutation()

  const filteredCodes = (codes ?? []).filter((c) => {
    if (filterPromotionId && c.promotion_id !== Number(filterPromotionId)) return false
    return true
  })

  const handleDeactivate = async (code: PromoCode) => {
    if (!confirm(`Деактивировать промокод "${code.code}"?`)) return
    try {
      await deactivateCode.mutateAsync(code.id)
      mutate()
    } catch {
      // error handled silently
    }
  }

  const handleCopyCode = (code: string) => {
    navigator.clipboard.writeText(code)
  }

  const getPromotionName = (promotionId?: number): string => {
    if (!promotionId || !promotions) return '—'
    const promo = promotions.find((p) => p.id === promotionId)
    return promo?.name ?? '—'
  }

  return (
    <div className="max-w-4xl">
      <div className="flex items-center justify-between mb-6 animate-in">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
            Промокоды
          </h1>
          <p className="text-sm text-neutral-500 mt-1">
            Создание и управление промокодами
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          type="button"
          className={cn(
            'flex items-center gap-2 py-2.5 px-4 rounded-lg',
            'bg-accent text-white text-sm font-medium',
            'hover:bg-accent-hover active:bg-accent/80',
            'transition-all duration-150',
            'focus:outline-none focus:ring-2 focus:ring-accent/20',
            'shadow-sm shadow-accent/20',
          )}
        >
          <Plus className="w-4 h-4" />
          <span className="hidden sm:inline">Создать промокод</span>
          <span className="sm:hidden">Создать</span>
        </button>
      </div>

      {!isLoading && !isError && (codes ?? []).length > 0 && (
        <div className="mb-4 animate-in">
          <select
            value={filterPromotionId}
            onChange={(e) => setFilterPromotionId(e.target.value)}
            className={cn(
              'px-4 py-2 rounded-lg border border-surface-border',
              'text-sm text-neutral-700',
              'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
              'transition-colors',
            )}
            aria-label="Фильтр по акции"
          >
            <option value="">Все акции</option>
            {promotions?.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
        </div>
      )}

      {isLoading ? (
        <div className="animate-in animate-in-delay-1">
          <TableSkeleton />
        </div>
      ) : isError ? (
        <ErrorState
          title="Не удалось загрузить промокоды"
          message="Проверьте подключение к серверу и попробуйте снова."
          onRetry={() => mutate()}
        />
      ) : filteredCodes.length === 0 ? (
        <EmptyState
          icon={Ticket}
          title="Промокодов пока нет"
          description="Создайте промокод для привлечения клиентов через различные каналы."
          actionLabel="Создать промокод"
          onAction={() => setShowCreateModal(true)}
        />
      ) : (
        <div className="bg-white rounded-2xl shadow-sm border border-surface-border overflow-hidden animate-in animate-in-delay-1">
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-surface-border">
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Код
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Акция
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Статус
                  </th>
                  <th className="text-left text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden md:table-cell">
                    Канал
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3 hidden sm:table-cell">
                    Использований
                  </th>
                  <th className="text-right text-xs font-medium text-neutral-400 uppercase tracking-wider px-6 py-3">
                    Действия
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-surface-border">
                {filteredCodes.map((promoCode) => {
                  const status = getCodeStatus(promoCode)
                  return (
                    <tr
                      key={promoCode.id}
                      className="hover:bg-neutral-50 transition-colors"
                    >
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-2">
                          <span className="text-sm font-mono font-medium text-neutral-900">
                            {promoCode.code}
                          </span>
                          <button
                            type="button"
                            onClick={() => handleCopyCode(promoCode.code)}
                            className="p-1 rounded text-neutral-300 hover:text-neutral-500 transition-colors"
                            title="Скопировать"
                            aria-label={`Скопировать код ${promoCode.code}`}
                          >
                            <Copy className="w-3.5 h-3.5" />
                          </button>
                        </div>
                        {promoCode.description && (
                          <p className="text-xs text-neutral-400 mt-0.5">
                            {promoCode.description}
                          </p>
                        )}
                      </td>
                      <td className="px-6 py-4 hidden sm:table-cell">
                        <span className="text-sm text-neutral-500">
                          {getPromotionName(promoCode.promotion_id)}
                        </span>
                      </td>
                      <td className="px-6 py-4">
                        <span
                          className={cn(
                            'text-xs font-medium px-2 py-1 rounded-full',
                            status.className,
                          )}
                        >
                          {status.label}
                        </span>
                      </td>
                      <td className="px-6 py-4 hidden md:table-cell">
                        <span className="text-sm text-neutral-500">
                          {promoCode.channel
                            ? channelLabels[promoCode.channel] ?? promoCode.channel
                            : '—'}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right hidden sm:table-cell">
                        <span className="text-sm font-mono text-neutral-500 tabular-nums">
                          {promoCode.usage_count}
                          {promoCode.usage_limit
                            ? ` / ${promoCode.usage_limit}`
                            : ''}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right">
                        {promoCode.active && (
                          <button
                            type="button"
                            onClick={() => handleDeactivate(promoCode)}
                            className="p-1.5 rounded-lg text-neutral-400 hover:text-red-600 hover:bg-red-50 transition-colors"
                            title="Деактивировать"
                            aria-label={`Деактивировать код ${promoCode.code}`}
                          >
                            <Ban className="w-4 h-4" />
                          </button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {showCreateModal && (
        <CreatePromoCodeModal onClose={() => setShowCreateModal(false)} />
      )}
    </div>
  )
}
