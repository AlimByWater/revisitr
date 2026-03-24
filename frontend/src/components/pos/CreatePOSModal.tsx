import { useState } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCreatePOSMutation } from '@/features/pos/queries'
import { useBotsQuery } from '@/features/bots/queries'

interface CreatePOSModalProps {
  onClose: () => void
}

export function CreatePOSModal({ onClose }: CreatePOSModalProps) {
  const createMutation = useCreatePOSMutation()

  const [name, setName] = useState('')
  const [address, setAddress] = useState('')
  const [phone, setPhone] = useState('')
  const [botId, setBotId] = useState<number | undefined>(undefined)
  const { data: bots } = useBotsQuery()

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    try {
      await createMutation.mutateAsync({
        name,
        address,
        phone,
        schedule: {},
        bot_id: botId,
      })
      onClose()
    } catch {
      // error displayed below via createMutation.isError
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-label="Добавление точки продаж"
    >
      <div
        className="absolute inset-0 bg-black/40"
        onClick={onClose}
        aria-hidden="true"
      />
      <div className="relative bg-white rounded-2xl shadow-lg border border-surface-border p-6 w-full max-w-md mx-4">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-semibold">Добавить точку продаж</h2>
          <button
            type="button"
            onClick={onClose}
            className="p-1 text-neutral-400 hover:text-neutral-700 transition-colors"
            aria-label="Закрыть"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label
              htmlFor="pos-create-name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Название заведения
            </label>
            <input
              id="pos-create-name"
              type="text"
              required
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Кафе «Уют» — Центр"
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>

          <div>
            <label
              htmlFor="pos-create-address"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Адрес
            </label>
            <input
              id="pos-create-address"
              type="text"
              value={address}
              onChange={(e) => setAddress(e.target.value)}
              placeholder="Москва, ул. Ленина, д. 15"
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>

          <div>
            <label
              htmlFor="pos-create-phone"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Телефон{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <input
              id="pos-create-phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              placeholder="+7 (999) 123-45-67"
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            />
          </div>

          <div>
            <label
              htmlFor="pos-create-bot"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Привязать к боту{' '}
              <span className="text-neutral-400 font-normal">(необязательно)</span>
            </label>
            <select
              id="pos-create-bot"
              value={botId ?? ''}
              onChange={(e) =>
                setBotId(e.target.value ? Number(e.target.value) : undefined)
              }
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm placeholder:text-neutral-400',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            >
              <option value="">Не привязывать</option>
              {bots?.map((b) => (
                <option key={b.id} value={b.id}>
                  {b.name}
                </option>
              ))}
            </select>
          </div>

          {createMutation.isError && (
            <p className="text-sm text-red-600">
              Не удалось добавить точку. Попробуйте ещё раз.
            </p>
          )}

          <div className="flex gap-3 pt-2">
            <button
              type="button"
              onClick={onClose}
              className={cn(
                'flex-1 px-4 py-2.5 rounded-lg text-sm font-medium',
                'border border-surface-border text-neutral-700',
                'hover:bg-neutral-50 transition-colors',
              )}
            >
              Отмена
            </button>
            <button
              type="submit"
              disabled={createMutation.isPending || !name.trim()}
              className={cn(
                'flex-1 px-4 py-2.5 rounded-xl text-sm font-semibold',
                'bg-accent text-white',
                'hover:bg-accent-hover transition-all duration-150',
                'shadow-sm shadow-accent/20',
                'disabled:opacity-50 disabled:cursor-not-allowed',
              )}
            >
              {createMutation.isPending ? 'Создание...' : 'Добавить'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
