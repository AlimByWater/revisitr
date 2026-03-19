import { useState } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCreateProgramMutation } from '@/features/loyalty/queries'
import type { CreateProgramRequest } from '@/features/loyalty/types'

interface CreateProgramModalProps {
  onClose: () => void
}

export function CreateProgramModal({ onClose }: CreateProgramModalProps) {
  const createMutation = useCreateProgramMutation()

  const [name, setName] = useState('')
  const [type, setType] = useState<'bonus' | 'discount'>('bonus')
  const [welcomeBonus, setWelcomeBonus] = useState(0)
  const [currencyName, setCurrencyName] = useState('')

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const body: CreateProgramRequest = {
      name,
      type,
      config: {
        welcome_bonus: welcomeBonus,
        currency_name: currencyName,
      },
    }
    try {
      await createMutation.mutateAsync(body)
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
      aria-label="Создание программы лояльности"
    >
      <div
        className="absolute inset-0 bg-black/40"
        onClick={onClose}
        aria-hidden="true"
      />
      <div className="relative bg-white rounded-2xl shadow-lg border border-surface-border p-6 w-full max-w-md mx-4">
        <div className="flex items-center justify-between mb-5">
          <h2 className="text-lg font-semibold">Создать программу</h2>
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
              htmlFor="program-name"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Название программы
            </label>
            <input
              id="program-name"
              type="text"
              required
              autoFocus
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Например: Золотой гость"
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
              htmlFor="program-type"
              className="block text-sm font-medium text-neutral-700 mb-1.5"
            >
              Тип программы
            </label>
            <select
              id="program-type"
              value={type}
              onChange={(e) =>
                setType(e.target.value as 'bonus' | 'discount')
              }
              className={cn(
                'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                'text-sm bg-white',
                'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                'transition-colors',
              )}
            >
              <option value="bonus">Бонусная</option>
              <option value="discount">Скидочная</option>
            </select>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label
                htmlFor="welcome-bonus"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Приветств. бонус
              </label>
              <input
                id="welcome-bonus"
                type="number"
                min={0}
                value={welcomeBonus}
                onChange={(e) => setWelcomeBonus(Number(e.target.value))}
                aria-describedby="welcome-bonus-hint"
                className={cn(
                  'w-full max-w-[120px] px-4 py-2.5 rounded-lg border border-surface-border',
                  'text-sm focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                )}
              />
              <p id="welcome-bonus-hint" className="mt-1.5 text-xs text-neutral-400">
                0 — без бонуса при регистрации
              </p>
            </div>

            <div>
              <label
                htmlFor="currency-name"
                className="block text-sm font-medium text-neutral-700 mb-1.5"
              >
                Название бонусов
              </label>
              <input
                id="currency-name"
                type="text"
                value={currencyName}
                onChange={(e) => setCurrencyName(e.target.value)}
                placeholder="баллы"
                aria-describedby="currency-name-hint"
                className={cn(
                  'w-full px-4 py-2.5 rounded-lg border border-surface-border',
                  'text-sm placeholder:text-neutral-400',
                  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                  'transition-colors',
                )}
              />
              <p id="currency-name-hint" className="mt-1.5 text-xs text-neutral-400">
                Напр.: «звёзды», «рубли», «бонусы»
              </p>
            </div>
          </div>

          {createMutation.isError && (
            <p className="text-sm text-red-600">
              Не удалось создать программу. Попробуйте ещё раз.
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
              {createMutation.isPending ? 'Создание...' : 'Создать'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
