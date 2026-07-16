import { useState } from 'react'
import { KeyRound, Copy, Check, Plus, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  usePluginKeysQuery,
  useCreatePluginKeyMutation,
  useRevokePluginKeyMutation,
} from '@/features/plugin-keys/queries'

function formatDate(dateStr?: string | null) {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function PluginKeysSection({ integrationId }: { integrationId: number }) {
  const { data, isLoading } = usePluginKeysQuery(integrationId)
  const createMutation = useCreatePluginKeyMutation(integrationId)
  const revokeMutation = useRevokePluginKeyMutation(integrationId)

  const [showForm, setShowForm] = useState(false)
  const [label, setLabel] = useState('')
  const [newKey, setNewKey] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [confirmRevoke, setConfirmRevoke] = useState<number | null>(null)

  const keys = data ?? []

  async function handleCreate() {
    const res = await createMutation.mutateAsync(label.trim() || 'Ключ кассы')
    setNewKey(res.key)
    setLabel('')
    setShowForm(false)
  }

  async function handleCopy() {
    if (!newKey) return
    await navigator.clipboard.writeText(newKey)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function handleRevoke(id: number) {
    await revokeMutation.mutateAsync(id)
    setConfirmRevoke(null)
  }

  return (
    <div className="bg-white rounded border border-neutral-900 p-4 space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <KeyRound className="w-4 h-4 text-neutral-500" />
          <h3 className="text-sm font-semibold text-neutral-900">Ключи плагина</h3>
        </div>
        {!showForm && (
          <button
            type="button"
            onClick={() => setShowForm(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded bg-accent text-white hover:bg-accent-hover transition-colors"
          >
            <Plus className="w-4 h-4" />
            Создать ключ
          </button>
        )}
      </div>

      <p className="text-xs text-neutral-500">
        API-ключ для плагина на кассе (iikoFront и др.). Впишите его в{' '}
        <code className="font-mono">revisitr.plugin.config.json</code> на кассе.
        Ключ показывается один раз при создании.
      </p>

      {/* Newly created key — shown once */}
      {newKey && (
        <div className="rounded border border-emerald-500/40 bg-emerald-50 p-3 space-y-2">
          <p className="text-xs font-medium text-emerald-800">
            Ключ создан. Скопируйте сейчас — больше он не будет показан.
          </p>
          <div className="flex items-center gap-2">
            <code className="flex-1 text-xs font-mono text-neutral-800 break-all bg-white rounded border border-neutral-200 px-2 py-1.5">
              {newKey}
            </code>
            <button
              type="button"
              onClick={handleCopy}
              className="flex items-center gap-1.5 px-2.5 py-1.5 text-xs font-medium rounded border border-neutral-900 text-neutral-700 hover:bg-neutral-50 transition-colors whitespace-nowrap"
            >
              {copied ? (
                <>
                  <Check className="w-3.5 h-3.5" /> Скопировано
                </>
              ) : (
                <>
                  <Copy className="w-3.5 h-3.5" /> Копировать
                </>
              )}
            </button>
          </div>
          <button
            type="button"
            onClick={() => setNewKey(null)}
            className="text-xs text-neutral-500 hover:text-neutral-700 transition-colors"
          >
            Закрыть
          </button>
        </div>
      )}

      {/* Create form */}
      {showForm && (
        <div className="flex gap-2">
          <input
            value={label}
            onChange={(e) => setLabel(e.target.value)}
            placeholder="Название (напр. Касса №1)"
            className="flex-1 text-sm rounded border border-neutral-300 px-3 py-1.5 focus:outline-none focus:border-accent"
          />
          <button
            type="button"
            onClick={handleCreate}
            disabled={createMutation.isPending}
            className="px-3 py-1.5 text-sm font-medium rounded bg-accent text-white hover:bg-accent-hover disabled:opacity-50 transition-colors whitespace-nowrap"
          >
            {createMutation.isPending ? 'Создание...' : 'Создать'}
          </button>
          <button
            type="button"
            onClick={() => {
              setShowForm(false)
              setLabel('')
            }}
            className="px-3 py-1.5 text-sm text-neutral-600 hover:text-neutral-800 transition-colors"
          >
            Отмена
          </button>
        </div>
      )}

      {createMutation.isError && (
        <p className="text-xs text-red-600">
          Не удалось создать ключ. Попробуйте снова.
        </p>
      )}

      {/* List */}
      {isLoading ? (
        <p className="text-sm text-neutral-500">Загрузка...</p>
      ) : keys.length === 0 ? (
        <p className="text-sm text-neutral-500">Ключей пока нет.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-200">
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Название
                </th>
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Создан
                </th>
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Использован
                </th>
                <th className="text-left py-2 px-3 text-neutral-500 font-medium">
                  Статус
                </th>
                <th className="py-2 px-3" />
              </tr>
            </thead>
            <tbody>
              {keys.map((k) => {
                const revoked = !!k.revoked_at
                return (
                  <tr
                    key={k.id}
                    className={cn(
                      'border-b border-neutral-200/50',
                      revoked && 'opacity-50',
                    )}
                  >
                    <td className="py-2.5 px-3 text-neutral-800">
                      {k.label || '—'}
                    </td>
                    <td className="py-2.5 px-3 text-neutral-600">
                      {formatDate(k.created_at)}
                    </td>
                    <td className="py-2.5 px-3 text-neutral-600">
                      {formatDate(k.last_used_at)}
                    </td>
                    <td className="py-2.5 px-3">
                      <span
                        className={cn(
                          'font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border',
                          revoked
                            ? 'bg-neutral-100 text-neutral-600 border-neutral-300'
                            : 'bg-emerald-500/10 text-emerald-700 border-emerald-500/30',
                        )}
                      >
                        {revoked ? 'Отозван' : 'Активен'}
                      </span>
                    </td>
                    <td className="py-2.5 px-3 text-right whitespace-nowrap">
                      {!revoked &&
                        (confirmRevoke === k.id ? (
                          <span className="flex items-center justify-end gap-2">
                            <button
                              type="button"
                              onClick={() => handleRevoke(k.id)}
                              disabled={revokeMutation.isPending}
                              className="text-xs font-medium text-red-600 hover:text-red-700 disabled:opacity-50 transition-colors"
                            >
                              Отозвать
                            </button>
                            <button
                              type="button"
                              onClick={() => setConfirmRevoke(null)}
                              className="text-xs text-neutral-500 hover:text-neutral-700 transition-colors"
                            >
                              Отмена
                            </button>
                          </span>
                        ) : (
                          <button
                            type="button"
                            onClick={() => setConfirmRevoke(k.id)}
                            className="text-neutral-400 hover:text-red-600 transition-colors"
                            title="Отозвать ключ"
                          >
                            <Trash2 className="w-4 h-4" />
                          </button>
                        ))}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
