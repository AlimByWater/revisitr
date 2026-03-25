import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  useWalletConfigsQuery,
  useWalletPassesQuery,
  useWalletStatsQuery,
  useSaveWalletConfigMutation,
  useRevokeWalletPassMutation,
} from '@/features/wallet/queries'
import { Smartphone, Apple, CreditCard, Shield, Palette, Users, Ban } from 'lucide-react'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'
import { EmptyState } from '@/components/common/EmptyState'
import type { WalletConfig, WalletDesign, WalletCredentials } from '@/features/wallet/types'

const platformConfig = {
  apple: { label: 'Apple Wallet', icon: Apple, color: 'bg-gray-900 text-white' },
  google: { label: 'Google Wallet', icon: Smartphone, color: 'bg-blue-600 text-white' },
}

export default function WalletPage() {
  const { data: configs, isLoading: configsLoading, error: configsError } = useWalletConfigsQuery()
  const { data: passes, isLoading: passesLoading } = useWalletPassesQuery()
  const { data: stats } = useWalletStatsQuery()
  const saveConfig = useSaveWalletConfigMutation()
  const revokePass = useRevokeWalletPassMutation()

  const [editPlatform, setEditPlatform] = useState<'apple' | 'google' | null>(null)

  if (configsError) return <ErrorState message="Ошибка загрузки настроек Wallet" />
  if (configsLoading) return <TableSkeleton rows={3} />

  const appleConfig = configs?.find((c) => c.platform === 'apple')
  const googleConfig = configs?.find((c) => c.platform === 'google')

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Wallet</h1>
        <p className="text-sm text-muted-foreground mt-1">
          Карты лояльности в Apple Wallet и Google Wallet
        </p>
      </div>

      {/* Stats */}
      {stats && (
        <div className="grid gap-4 sm:grid-cols-4">
          <StatCard label="Всего карт" value={stats.total_passes} icon={CreditCard} />
          <StatCard label="Apple Wallet" value={stats.apple_passes} icon={Apple} />
          <StatCard label="Google Wallet" value={stats.google_passes} icon={Smartphone} />
          <StatCard label="Активных" value={stats.active_passes} icon={Users} />
        </div>
      )}

      {/* Platform configs */}
      <div className="grid gap-4 sm:grid-cols-2">
        <PlatformCard
          platform="apple"
          config={appleConfig}
          onEdit={() => setEditPlatform('apple')}
        />
        <PlatformCard
          platform="google"
          config={googleConfig}
          onEdit={() => setEditPlatform('google')}
        />
      </div>

      {/* Edit form */}
      {editPlatform && (
        <ConfigForm
          platform={editPlatform}
          config={configs?.find((c) => c.platform === editPlatform)}
          onSave={async (credentials, design, isEnabled) => {
            await saveConfig.trigger({
              platform: editPlatform,
              data: { platform: editPlatform, is_enabled: isEnabled, credentials, design },
            })
            setEditPlatform(null)
          }}
          onCancel={() => setEditPlatform(null)}
        />
      )}

      {/* Passes list */}
      <div className="space-y-3">
        <h2 className="text-lg font-semibold">Выданные карты</h2>
        {passesLoading ? (
          <TableSkeleton rows={5} />
        ) : !passes || passes.length === 0 ? (
          <EmptyState
            icon={CreditCard}
            title="Нет выданных карт"
            description="Карты создаются при выдаче клиенту через API или бота"
          />
        ) : (
          <div className="rounded-lg border overflow-hidden">
            <table className="w-full text-sm">
              <thead className="bg-muted/50">
                <tr>
                  <th className="text-left px-4 py-2 font-medium">Клиент</th>
                  <th className="text-left px-4 py-2 font-medium">Платформа</th>
                  <th className="text-left px-4 py-2 font-medium">Баланс</th>
                  <th className="text-left px-4 py-2 font-medium">Уровень</th>
                  <th className="text-left px-4 py-2 font-medium">Статус</th>
                  <th className="text-right px-4 py-2 font-medium">Действия</th>
                </tr>
              </thead>
              <tbody className="divide-y">
                {passes.map((pass) => {
                  const plt = platformConfig[pass.platform]
                  return (
                    <tr key={pass.id} className="hover:bg-muted/30">
                      <td className="px-4 py-2 tabular-nums">#{pass.client_id}</td>
                      <td className="px-4 py-2">
                        <span className={cn('inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium', plt.color)}>
                          <plt.icon className="h-3 w-3" />
                          {plt.label}
                        </span>
                      </td>
                      <td className="px-4 py-2 tabular-nums">{pass.last_balance}</td>
                      <td className="px-4 py-2">{pass.last_level || '—'}</td>
                      <td className="px-4 py-2">
                        <span className={cn(
                          'inline-block rounded-full px-2 py-0.5 text-xs font-medium',
                          pass.status === 'active' ? 'bg-green-100 text-green-700' :
                          pass.status === 'revoked' ? 'bg-red-100 text-red-700' :
                          'bg-yellow-100 text-yellow-700',
                        )}>
                          {pass.status === 'active' ? 'Активна' : pass.status === 'revoked' ? 'Отозвана' : 'Приостановлена'}
                        </span>
                      </td>
                      <td className="px-4 py-2 text-right">
                        {pass.status === 'active' && (
                          <button
                            onClick={() => revokePass.trigger(pass.id)}
                            className="inline-flex items-center gap-1 rounded-md border border-red-200 px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50"
                          >
                            <Ban className="h-3 w-3" />
                            Отозвать
                          </button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}

function StatCard({ label, value, icon: Icon }: { label: string; value: number; icon: typeof CreditCard }) {
  return (
    <div className="rounded-lg border bg-card p-4">
      <div className="flex items-center justify-between">
        <p className="text-sm text-muted-foreground">{label}</p>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </div>
      <p className="text-2xl font-bold mt-1 tabular-nums">{value}</p>
    </div>
  )
}

function PlatformCard({
  platform,
  config,
  onEdit,
}: {
  platform: 'apple' | 'google'
  config?: WalletConfig
  onEdit: () => void
}) {
  const plt = platformConfig[platform]
  const Icon = plt.icon

  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className={cn('rounded-lg p-2', plt.color)}>
            <Icon className="h-5 w-5" />
          </div>
          <div>
            <h3 className="font-medium">{plt.label}</h3>
            <p className="text-xs text-muted-foreground">
              {config?.is_enabled ? 'Подключён' : 'Не настроен'}
            </p>
          </div>
        </div>
        <span className={cn(
          'inline-block h-2.5 w-2.5 rounded-full',
          config?.is_enabled ? 'bg-green-500' : 'bg-gray-300',
        )} />
      </div>
      {config?.design.description && (
        <p className="text-sm text-muted-foreground">{config.design.description}</p>
      )}
      <button
        onClick={onEdit}
        className="w-full rounded-md border px-3 py-2 text-sm font-medium hover:bg-accent"
      >
        {config ? 'Настроить' : 'Подключить'}
      </button>
    </div>
  )
}

function ConfigForm({
  platform,
  config,
  onSave,
  onCancel,
}: {
  platform: 'apple' | 'google'
  config?: WalletConfig
  onSave: (credentials: WalletCredentials, design: WalletDesign, isEnabled: boolean) => void
  onCancel: () => void
}) {
  const [isEnabled, setIsEnabled] = useState(config?.is_enabled ?? false)
  const [design, setDesign] = useState<WalletDesign>(config?.design ?? {})
  const [credentials, setCredentials] = useState<WalletCredentials>({})

  const plt = platformConfig[platform]

  return (
    <div className="rounded-lg border bg-card p-4 space-y-4">
      <h3 className="font-medium flex items-center gap-2">
        <Shield className="h-4 w-4" />
        Настройка {plt.label}
      </h3>

      <label className="flex items-center gap-2">
        <input
          type="checkbox"
          checked={isEnabled}
          onChange={(e) => setIsEnabled(e.target.checked)}
          className="rounded"
        />
        <span className="text-sm">Включить {plt.label}</span>
      </label>

      {/* Credentials */}
      <div className="space-y-2">
        <h4 className="text-sm font-medium flex items-center gap-1">
          <Shield className="h-3.5 w-3.5" />
          Учётные данные
        </h4>
        {platform === 'apple' ? (
          <>
            <input
              type="text"
              placeholder="Pass Type ID (pass.com.example.loyalty)"
              value={credentials.pass_type_id ?? ''}
              onChange={(e) => setCredentials({ ...credentials, pass_type_id: e.target.value })}
              className="w-full rounded-md border px-3 py-2 text-sm"
            />
            <input
              type="text"
              placeholder="Team ID"
              value={credentials.team_id ?? ''}
              onChange={(e) => setCredentials({ ...credentials, team_id: e.target.value })}
              className="w-full rounded-md border px-3 py-2 text-sm"
            />
            <textarea
              placeholder="Сертификат (.p12 в base64)"
              value={credentials.certificate ?? ''}
              onChange={(e) => setCredentials({ ...credentials, certificate: e.target.value })}
              rows={3}
              className="w-full rounded-md border px-3 py-2 text-sm font-mono"
            />
          </>
        ) : (
          <>
            <input
              type="text"
              placeholder="Issuer ID"
              value={credentials.issuer_id ?? ''}
              onChange={(e) => setCredentials({ ...credentials, issuer_id: e.target.value })}
              className="w-full rounded-md border px-3 py-2 text-sm"
            />
            <textarea
              placeholder="Service Account Key (JSON)"
              value={credentials.service_account_key ?? ''}
              onChange={(e) => setCredentials({ ...credentials, service_account_key: e.target.value })}
              rows={4}
              className="w-full rounded-md border px-3 py-2 text-sm font-mono"
            />
          </>
        )}
      </div>

      {/* Design */}
      <div className="space-y-2">
        <h4 className="text-sm font-medium flex items-center gap-1">
          <Palette className="h-3.5 w-3.5" />
          Оформление карты
        </h4>
        <input
          type="text"
          placeholder="Описание на карте"
          value={design.description ?? ''}
          onChange={(e) => setDesign({ ...design, description: e.target.value })}
          className="w-full rounded-md border px-3 py-2 text-sm"
        />
        <div className="grid grid-cols-3 gap-2">
          <div>
            <label className="text-xs text-muted-foreground">Фон</label>
            <input
              type="color"
              value={design.background_color || '#1a1a2e'}
              onChange={(e) => setDesign({ ...design, background_color: e.target.value })}
              className="w-full h-8 rounded border cursor-pointer"
            />
          </div>
          <div>
            <label className="text-xs text-muted-foreground">Текст</label>
            <input
              type="color"
              value={design.foreground_color || '#ffffff'}
              onChange={(e) => setDesign({ ...design, foreground_color: e.target.value })}
              className="w-full h-8 rounded border cursor-pointer"
            />
          </div>
          <div>
            <label className="text-xs text-muted-foreground">Метки</label>
            <input
              type="color"
              value={design.label_color || '#a0a0a0'}
              onChange={(e) => setDesign({ ...design, label_color: e.target.value })}
              className="w-full h-8 rounded border cursor-pointer"
            />
          </div>
        </div>
      </div>

      <div className="flex gap-2 pt-2">
        <button
          onClick={() => onSave(credentials, design, isEnabled)}
          className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          Сохранить
        </button>
        <button
          onClick={onCancel}
          className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
        >
          Отмена
        </button>
      </div>
    </div>
  )
}
