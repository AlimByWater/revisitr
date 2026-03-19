import { useState } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCreateIntegrationMutation } from '@/features/integrations/queries'

interface CreateIntegrationModalProps {
  onClose: () => void
}

const TYPES = [
  {
    value: 'mock',
    label: 'Mock (Dev)',
    description: 'Тестовые данные для разработки',
    configFields: [],
  },
  {
    value: 'iiko',
    label: 'iiko',
    description: 'iiko Cloud API',
    configFields: [
      { key: 'api_key', label: 'API Login', type: 'password' as const },
      {
        key: 'api_url',
        label: 'API URL',
        type: 'text' as const,
        placeholder: 'https://api-ru.iiko.services/api/1',
      },
    ],
  },
  {
    value: 'rkeeper',
    label: 'r-keeper',
    description: 'r-keeper 7 XML Interface',
    configFields: [
      {
        key: 'api_url',
        label: 'Server URL',
        type: 'text' as const,
        placeholder: 'https://ip:port/rk7api/v0/xmlinterface.xml',
      },
      { key: 'username', label: 'Логин', type: 'text' as const },
      { key: 'password', label: 'Пароль', type: 'password' as const },
    ],
  },
] as const

export function CreateIntegrationModal({
  onClose,
}: CreateIntegrationModalProps) {
  const createMutation = useCreateIntegrationMutation()
  const [selectedType, setSelectedType] = useState('')
  const [config, setConfig] = useState<Record<string, string>>({})

  const selectedDef = TYPES.find((t) => t.value === selectedType)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    try {
      await createMutation.mutateAsync({
        type: selectedType,
        config,
      })
      onClose()
    } catch {
      // error shown via createMutation.isError
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center"
      role="dialog"
      aria-modal="true"
      aria-label="Добавление интеграции"
    >
      <div
        className="absolute inset-0 bg-black/40"
        onClick={onClose}
        onKeyDown={(e) => e.key === 'Escape' && onClose()}
        role="presentation"
      />
      <div className="relative bg-white rounded-2xl shadow-xl w-full max-w-lg mx-4 animate-in">
        <div className="flex items-center justify-between p-6 border-b border-surface-border">
          <h2 className="text-lg font-semibold text-neutral-900">
            Новая интеграция
          </h2>
          <button
            type="button"
            onClick={onClose}
            className="p-1.5 rounded-lg text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100 transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {!selectedType ? (
            <div className="space-y-2">
              <p className="text-sm text-neutral-500 mb-3">
                Выберите тип POS-системы:
              </p>
              {TYPES.map((t) => (
                <button
                  key={t.value}
                  type="button"
                  onClick={() => setSelectedType(t.value)}
                  className={cn(
                    'w-full text-left p-4 rounded-xl border border-surface-border',
                    'hover:border-neutral-300 hover:shadow-sm transition-all',
                  )}
                >
                  <div className="flex items-center gap-3">
                    <span className="font-medium text-neutral-900">
                      {t.label}
                    </span>
                    {t.value === 'mock' && (
                      <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-orange-100 text-orange-700">
                        DEV
                      </span>
                    )}
                  </div>
                  <p className="text-sm text-neutral-500 mt-0.5">
                    {t.description}
                  </p>
                </button>
              ))}
            </div>
          ) : (
            <>
              <button
                type="button"
                onClick={() => {
                  setSelectedType('')
                  setConfig({})
                }}
                className="text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
              >
                &larr; Назад к выбору типа
              </button>

              <div className="flex items-center gap-2 mb-2">
                <h3 className="font-medium text-neutral-900">
                  {selectedDef?.label}
                </h3>
                {selectedType === 'mock' && (
                  <span className="text-[10px] font-semibold px-1.5 py-0.5 rounded bg-orange-100 text-orange-700">
                    DEV
                  </span>
                )}
              </div>

              {selectedDef?.configFields.map((field) => (
                <label key={field.key} className="block">
                  <span className="text-sm font-medium text-neutral-700">
                    {field.label}
                  </span>
                  <input
                    type={field.type}
                    value={config[field.key] || ''}
                    onChange={(e) =>
                      setConfig((prev) => ({
                        ...prev,
                        [field.key]: e.target.value,
                      }))
                    }
                    placeholder={'placeholder' in field ? field.placeholder : ''}
                    className={cn(
                      'mt-1 block w-full rounded-lg border border-surface-border px-3 py-2',
                      'text-sm text-neutral-900 placeholder:text-neutral-400',
                      'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
                    )}
                  />
                </label>
              ))}

              {selectedType === 'mock' && (
                <p className="text-sm text-neutral-500 bg-orange-50 p-3 rounded-lg">
                  Mock-интеграция создаст тестовые данные: клиентов, заказы и
                  меню. Настройка не требуется.
                </p>
              )}

              {createMutation.isError && (
                <p className="text-sm text-red-600">
                  Ошибка при создании интеграции. Попробуйте снова.
                </p>
              )}

              <div className="flex gap-3 pt-2">
                <button
                  type="button"
                  onClick={onClose}
                  className="flex-1 py-2.5 px-4 rounded-lg border border-surface-border text-sm font-medium text-neutral-700 hover:bg-neutral-50 transition-colors"
                >
                  Отмена
                </button>
                <button
                  type="submit"
                  disabled={createMutation.isPending}
                  className={cn(
                    'flex-1 py-2.5 px-4 rounded-lg text-sm font-medium',
                    'bg-accent text-white',
                    'hover:bg-accent-hover active:bg-accent/80',
                    'disabled:opacity-50 disabled:cursor-not-allowed',
                    'transition-all duration-150',
                  )}
                >
                  {createMutation.isPending ? 'Создание...' : 'Создать'}
                </button>
              </div>
            </>
          )}
        </form>
      </div>
    </div>
  )
}
