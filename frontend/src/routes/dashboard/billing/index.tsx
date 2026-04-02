import { useState } from 'react'
import { cn } from '@/lib/utils'
import { useTariffsQuery, useSubscriptionQuery, useSubscribeMutation, useChangePlanMutation, useCancelSubscriptionMutation } from '@/features/billing/queries'
import type { Tariff } from '@/features/billing/types'
import { Link } from 'react-router-dom'
import { Check, Crown, Zap, AlertTriangle, CreditCard } from 'lucide-react'

function formatPrice(kopeks: number): string {
  if (kopeks === 0) return 'Бесплатно'
  return `${(kopeks / 100).toLocaleString('ru-RU')} ₽`
}

function StatusBadge({ status }: { status: string }) {
  const config: Record<string, { label: string; className: string }> = {
    trialing: { label: 'Пробный период', className: 'bg-blue-50 text-blue-600' },
    active: { label: 'Активна', className: 'bg-green-50 text-green-600' },
    past_due: { label: 'Просрочена', className: 'bg-red-50 text-red-600' },
    canceled: { label: 'Отменена', className: 'bg-neutral-100 text-neutral-500' },
    expired: { label: 'Истекла', className: 'bg-neutral-100 text-neutral-500' },
  }
  const c = config[status] ?? { label: status, className: 'bg-neutral-100 text-neutral-500' }
  return (
    <span className={cn('text-xs font-medium px-2.5 py-1 rounded-full', c.className)}>
      {c.label}
    </span>
  )
}

function TariffCard({
  tariff,
  isCurrent,
  onSelect,
  isLoading,
}: {
  tariff: Tariff
  isCurrent: boolean
  onSelect: (slug: string) => void
  isLoading: boolean
}) {
  const isPro = tariff.slug === 'pro'
  const isEnterprise = tariff.slug === 'enterprise'

  return (
    <div
      className={cn(
        'relative flex flex-col rounded border p-6 transition-all',
        isCurrent
          ? 'border-accent/50 bg-accent/5 ring-1 ring-accent/20'
          : 'border-neutral-200 bg-white hover:border-neutral-300 hover:shadow-md',
      )}
    >
      {isPro && (
        <div className="absolute -top-3 left-1/2 -translate-x-1/2">
          <span className="bg-accent text-white text-[10px] font-bold px-3 py-1 rounded-full uppercase tracking-wider flex items-center gap-1">
            <Crown className="w-3 h-3" /> Популярный
          </span>
        </div>
      )}

      <div className="mb-4">
        <h3 className="text-lg font-semibold text-neutral-900">{tariff.name}</h3>
        <p className="text-2xl font-bold text-neutral-900 mt-2">
          {formatPrice(tariff.price)}
          {tariff.price > 0 && (
            <span className="text-sm font-normal text-neutral-400">
              /{tariff.interval === 'month' ? 'мес' : 'год'}
            </span>
          )}
        </p>
      </div>

      <div className="flex-1 space-y-3 mb-6">
        <LimitRow label="Клиентов" value={tariff.limits.max_clients} />
        <LimitRow label="Ботов" value={tariff.limits.max_bots} />
        <LimitRow label="Рассылок/мес" value={tariff.limits.max_campaigns_per_month} />
        <LimitRow label="Точек продаж" value={tariff.limits.max_pos} />
        {tariff.features.rfm && <FeatureRow label="RFM-сегментация" />}
        {tariff.features.advanced_campaigns && <FeatureRow label="A/B тесты рассылок" />}
      </div>

      {isEnterprise ? (
        <button
          type="button"
          className="w-full py-2.5 rounded text-sm font-medium border border-neutral-200 text-neutral-600 hover:text-neutral-900 hover:border-neutral-300 transition-colors"
        >
          Связаться
        </button>
      ) : isCurrent ? (
        <div className="w-full py-2.5 rounded text-sm font-medium text-center text-accent">
          Текущий план
        </div>
      ) : (
        <button
          type="button"
          onClick={() => onSelect(tariff.slug)}
          disabled={isLoading}
          className={cn(
            'w-full py-2.5 rounded text-sm font-medium transition-colors',
            isPro
              ? 'bg-accent text-white hover:bg-accent/90'
              : 'bg-neutral-100 text-neutral-700 hover:bg-neutral-200',
            isLoading && 'opacity-50 cursor-not-allowed',
          )}
        >
          {isLoading ? 'Загрузка...' : 'Выбрать'}
        </button>
      )}
    </div>
  )
}

function LimitRow({ label, value }: { label: string; value: number }) {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-neutral-500">{label}</span>
      <span className="text-neutral-900 font-medium">
        {value === -1 ? '∞' : value.toLocaleString('ru-RU')}
      </span>
    </div>
  )
}

function FeatureRow({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <Check className="w-4 h-4 text-green-500 shrink-0" />
      <span className="text-neutral-600">{label}</span>
    </div>
  )
}

export default function BillingPage() {
  const { data: tariffs, isLoading: tariffsLoading } = useTariffsQuery()
  const { data: subscription, isLoading: subLoading } = useSubscriptionQuery()
  const { trigger: subscribe, isMutating: subscribing } = useSubscribeMutation()
  const { trigger: changePlan, isMutating: changing } = useChangePlanMutation()
  const { trigger: cancel, isMutating: canceling } = useCancelSubscriptionMutation()
  const [showCancel, setShowCancel] = useState(false)

  const isLoading = tariffsLoading || subLoading
  const isActing = subscribing || changing

  const handleSelect = async (slug: string) => {
    if (subscription) {
      await changePlan({ tariff_slug: slug })
    } else {
      await subscribe({ tariff_slug: slug })
    }
  }

  const handleCancel = async () => {
    await cancel(undefined as never)
    setShowCancel(false)
  }

  return (
    <div className="space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Биллинг</h1>
          <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">Управление подпиской и тарифом</p>
        </div>
        <Link
          to="/dashboard/billing/invoices"
          className="text-sm text-neutral-500 hover:text-neutral-900 transition-colors flex items-center gap-2"
        >
          <CreditCard className="w-4 h-4" />
          История счетов
        </Link>
      </div>

      {/* Current subscription */}
      {subscription && (
        <div className="rounded border border-neutral-900 bg-white p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div>
                <div className="flex items-center gap-3">
                  <h2 className="text-lg font-semibold text-neutral-900">{subscription.tariff_name}</h2>
                  <StatusBadge status={subscription.status} />
                </div>
                <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">
                  Действует до {new Date(subscription.current_period_end).toLocaleDateString('ru-RU')}
                </p>
              </div>
            </div>

            {subscription.status !== 'canceled' && (
              <div>
                {showCancel ? (
                  <div className="flex items-center gap-2">
                    <button
                      type="button"
                      onClick={handleCancel}
                      disabled={canceling}
                      className="px-4 py-2 text-sm font-medium text-red-600 bg-red-50 rounded hover:bg-red-100 transition-colors"
                    >
                      {canceling ? 'Отмена...' : 'Подтвердить'}
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowCancel(false)}
                      className="px-4 py-2 text-sm text-neutral-400 hover:text-neutral-700 transition-colors"
                    >
                      Нет
                    </button>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={() => setShowCancel(true)}
                    className="text-sm text-neutral-400 hover:text-red-500 transition-colors"
                  >
                    Отменить подписку
                  </button>
                )}
              </div>
            )}
          </div>

          {subscription.status === 'past_due' && (
            <div className="mt-4 flex items-center gap-2 text-sm text-amber-700 bg-amber-50 rounded p-3">
              <AlertTriangle className="w-4 h-4 shrink-0" />
              Подписка просрочена. Оплатите счёт, чтобы сохранить доступ к функциям.
            </div>
          )}
        </div>
      )}

      {/* Tariff cards */}
      {isLoading ? (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="rounded border border-neutral-900 bg-white p-6 h-80 animate-pulse" />
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {tariffs?.map((tariff) => (
            <TariffCard
              key={tariff.id}
              tariff={tariff}
              isCurrent={subscription?.tariff_slug === tariff.slug}
              onSelect={handleSelect}
              isLoading={isActing}
            />
          ))}
        </div>
      )}

      {/* Feature gating info */}
      <div className="rounded border border-neutral-900 bg-white p-6">
        <div className="flex items-center gap-2 mb-4">
          <Zap className="w-5 h-5 text-accent" />
          <h3 className="text-base font-semibold text-neutral-900">Как работают тарифы</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-neutral-500">
          <div>
            <p className="text-neutral-700 font-medium mb-1">Лимиты</p>
            <p>Каждый тариф ограничивает количество клиентов, ботов, рассылок и точек продаж.</p>
          </div>
          <div>
            <p className="text-neutral-700 font-medium mb-1">Функции</p>
            <p>Продвинутые функции (RFM, A/B-тесты) доступны на тарифах Pro и Enterprise.</p>
          </div>
          <div>
            <p className="text-neutral-700 font-medium mb-1">Оплата</p>
            <p>При смене тарифа новый счёт создаётся автоматически. Оплата через ЮKassa.</p>
          </div>
        </div>
      </div>
    </div>
  )
}
