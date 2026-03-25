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
    trialing: { label: 'Пробный период', className: 'bg-blue-500/10 text-blue-400' },
    active: { label: 'Активна', className: 'bg-green-500/10 text-green-400' },
    past_due: { label: 'Просрочена', className: 'bg-red-500/10 text-red-400' },
    canceled: { label: 'Отменена', className: 'bg-neutral-500/10 text-neutral-400' },
    expired: { label: 'Истекла', className: 'bg-neutral-500/10 text-neutral-400' },
  }
  const c = config[status] ?? { label: status, className: 'bg-neutral-500/10 text-neutral-400' }
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
        'relative flex flex-col rounded-xl border p-6 transition-all',
        isCurrent
          ? 'border-accent bg-accent/5 ring-1 ring-accent/20'
          : isPro
            ? 'border-white/10 bg-white/[0.03] hover:border-accent/30'
            : 'border-white/10 bg-white/[0.02] hover:border-white/20',
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
        <h3 className="text-lg font-semibold text-white">{tariff.name}</h3>
        <p className="text-2xl font-bold text-white mt-2">
          {formatPrice(tariff.price)}
          {tariff.price > 0 && (
            <span className="text-sm font-normal text-white/40">
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
          className="w-full py-2.5 rounded-lg text-sm font-medium border border-white/10 text-white/60 hover:text-white hover:border-white/20 transition-colors"
        >
          Связаться
        </button>
      ) : isCurrent ? (
        <div className="w-full py-2.5 rounded-lg text-sm font-medium text-center text-accent">
          Текущий план
        </div>
      ) : (
        <button
          type="button"
          onClick={() => onSelect(tariff.slug)}
          disabled={isLoading}
          className={cn(
            'w-full py-2.5 rounded-lg text-sm font-medium transition-colors',
            isPro
              ? 'bg-accent text-white hover:bg-accent/90'
              : 'bg-white/10 text-white hover:bg-white/15',
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
      <span className="text-white/50">{label}</span>
      <span className="text-white font-medium">
        {value === -1 ? '∞' : value.toLocaleString('ru-RU')}
      </span>
    </div>
  )
}

function FeatureRow({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <Check className="w-4 h-4 text-green-400 shrink-0" />
      <span className="text-white/70">{label}</span>
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
          <h1 className="text-2xl font-bold text-white">Биллинг</h1>
          <p className="text-sm text-white/40 mt-1">Управление подпиской и тарифом</p>
        </div>
        <Link
          to="/dashboard/billing/invoices"
          className="text-sm text-white/50 hover:text-white transition-colors flex items-center gap-2"
        >
          <CreditCard className="w-4 h-4" />
          История счетов
        </Link>
      </div>

      {/* Current subscription */}
      {subscription && (
        <div className="rounded-xl border border-white/10 bg-white/[0.02] p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-4">
              <div>
                <div className="flex items-center gap-3">
                  <h2 className="text-lg font-semibold text-white">{subscription.tariff_name}</h2>
                  <StatusBadge status={subscription.status} />
                </div>
                <p className="text-sm text-white/40 mt-1">
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
                      className="px-4 py-2 text-sm font-medium text-red-400 bg-red-500/10 rounded-lg hover:bg-red-500/20 transition-colors"
                    >
                      {canceling ? 'Отмена...' : 'Подтвердить'}
                    </button>
                    <button
                      type="button"
                      onClick={() => setShowCancel(false)}
                      className="px-4 py-2 text-sm text-white/40 hover:text-white transition-colors"
                    >
                      Нет
                    </button>
                  </div>
                ) : (
                  <button
                    type="button"
                    onClick={() => setShowCancel(true)}
                    className="text-sm text-white/30 hover:text-red-400 transition-colors"
                  >
                    Отменить подписку
                  </button>
                )}
              </div>
            )}
          </div>

          {subscription.status === 'past_due' && (
            <div className="mt-4 flex items-center gap-2 text-sm text-amber-400 bg-amber-500/10 rounded-lg p-3">
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
            <div key={i} className="rounded-xl border border-white/10 bg-white/[0.02] p-6 h-80 animate-pulse" />
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
      <div className="rounded-xl border border-white/10 bg-white/[0.02] p-6">
        <div className="flex items-center gap-2 mb-4">
          <Zap className="w-5 h-5 text-accent" />
          <h3 className="text-base font-semibold text-white">Как работают тарифы</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-white/50">
          <div>
            <p className="text-white/70 font-medium mb-1">Лимиты</p>
            <p>Каждый тариф ограничивает количество клиентов, ботов, рассылок и точек продаж.</p>
          </div>
          <div>
            <p className="text-white/70 font-medium mb-1">Функции</p>
            <p>Продвинутые функции (RFM, A/B-тесты) доступны на тарифах Pro и Enterprise.</p>
          </div>
          <div>
            <p className="text-white/70 font-medium mb-1">Оплата</p>
            <p>При смене тарифа новый счёт создаётся автоматически. Оплата через ЮKassa.</p>
          </div>
        </div>
      </div>
    </div>
  )
}
