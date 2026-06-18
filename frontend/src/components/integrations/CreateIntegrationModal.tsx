import { useState } from 'react'
import { X } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useCreateIntegrationMutation,
  useDiscoverIntegrationMutation,
} from '@/features/integrations/queries'
import type { POSDiscovery } from '@/features/integrations/types'

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

const inputClass = cn(
  'mt-1 block w-full rounded-lg border border-surface-border px-3 py-2',
  'text-sm text-neutral-900 placeholder:text-neutral-400',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
)

export function CreateIntegrationModal({
  onClose,
}: CreateIntegrationModalProps) {
  const createMutation = useCreateIntegrationMutation()
  const discoverMutation = useDiscoverIntegrationMutation()
  const [selectedType, setSelectedType] = useState('')
  const [config, setConfig] = useState<Record<string, string>>({})

  // iiko 2-step state: after discovery, the user picks an org (and menu).
  const [discovery, setDiscovery] = useState<POSDiscovery | null>(null)

  const selectedDef = TYPES.find((t) => t.value === selectedType)
  const isIiko = selectedType === 'iiko'

  function resetType() {
    setSelectedType('')
    setConfig({})
    setDiscovery(null)
    discoverMutation.reset()
    createMutation.reset()
  }

  async function handleDiscover(e: React.FormEvent) {
    e.preventDefault()
    try {
      const result = await discoverMutation.mutateAsync({
        type: selectedType,
        config: {
          api_key: config.api_key,
          ...(config.api_url ? { api_url: config.api_url } : {}),
        },
      })
      setDiscovery(result)
      // Preselect when there is exactly one choice.
      setConfig((prev) => ({
        ...prev,
        org_id:
          result.organizations.length === 1
            ? result.organizations[0].id
            : prev.org_id || '',
        external_menu_id:
          result.external_menus.length === 1
            ? result.external_menus[0].id
            : prev.external_menu_id || '',
      }))
    } catch {
      // error shown via discoverMutation.isError
    }
  }

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    try {
      await createMutation.mutateAsync({ type: selectedType, config })
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

        <div className="p-6 space-y-4">
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
                onClick={resetType}
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

              {isIiko ? (
                <IikoFlow
                  config={config}
                  setConfig={setConfig}
                  discovery={discovery}
                  onDiscover={handleDiscover}
                  onCreate={handleCreate}
                  onCancel={onClose}
                  onBack={() => {
                    setDiscovery(null)
                    discoverMutation.reset()
                  }}
                  discovering={discoverMutation.isPending}
                  discoverError={discoverMutation.isError}
                  creating={createMutation.isPending}
                  createError={createMutation.isError}
                />
              ) : (
                <form onSubmit={handleCreate} className="space-y-4">
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
                        placeholder={
                          'placeholder' in field ? field.placeholder : ''
                        }
                        className={inputClass}
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
                        'bg-accent text-white hover:bg-accent-hover active:bg-accent/80',
                        'disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-150',
                      )}
                    >
                      {createMutation.isPending ? 'Создание...' : 'Создать'}
                    </button>
                  </div>
                </form>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  )
}

interface IikoFlowProps {
  config: Record<string, string>
  setConfig: React.Dispatch<React.SetStateAction<Record<string, string>>>
  discovery: POSDiscovery | null
  onDiscover: (e: React.FormEvent) => void
  onCreate: (e: React.FormEvent) => void
  onCancel: () => void
  onBack: () => void
  discovering: boolean
  discoverError: boolean
  creating: boolean
  createError: boolean
}

function IikoFlow({
  config,
  setConfig,
  discovery,
  onDiscover,
  onCreate,
  onCancel,
  onBack,
  discovering,
  discoverError,
  creating,
  createError,
}: IikoFlowProps) {
  // Step 1: enter apiLogin and discover organizations.
  if (!discovery) {
    return (
      <form onSubmit={onDiscover} className="space-y-4">
        <label className="block">
          <span className="text-sm font-medium text-neutral-700">
            API Login
          </span>
          <input
            type="password"
            value={config.api_key || ''}
            onChange={(e) =>
              setConfig((prev) => ({ ...prev, api_key: e.target.value }))
            }
            placeholder="apiLogin из iiko Cloud API"
            className={inputClass}
          />
        </label>
        <label className="block">
          <span className="text-sm font-medium text-neutral-700">
            API URL <span className="text-neutral-400">(необязательно)</span>
          </span>
          <input
            type="text"
            value={config.api_url || ''}
            onChange={(e) =>
              setConfig((prev) => ({ ...prev, api_url: e.target.value }))
            }
            placeholder="https://api-ru.iiko.services/api/1"
            className={inputClass}
          />
        </label>

        <p className="text-sm text-neutral-500 bg-neutral-50 p-3 rounded-lg">
          Введите API Login из настроек Cloud API в iiko. Организацию и меню
          подтянем автоматически.
        </p>

        {discoverError && (
          <p className="text-sm text-red-600">
            Не удалось подключиться к iiko с этим API Login. Проверьте ключ.
          </p>
        )}

        <div className="flex gap-3 pt-2">
          <button
            type="button"
            onClick={onCancel}
            className="flex-1 py-2.5 px-4 rounded-lg border border-surface-border text-sm font-medium text-neutral-700 hover:bg-neutral-50 transition-colors"
          >
            Отмена
          </button>
          <button
            type="submit"
            disabled={discovering || !config.api_key}
            className={cn(
              'flex-1 py-2.5 px-4 rounded-lg text-sm font-medium',
              'bg-accent text-white hover:bg-accent-hover active:bg-accent/80',
              'disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-150',
            )}
          >
            {discovering ? 'Подключение...' : 'Продолжить'}
          </button>
        </div>
      </form>
    )
  }

  // Step 2: pick organization (and optional external menu), then create.
  return (
    <form onSubmit={onCreate} className="space-y-4">
      <button
        type="button"
        onClick={onBack}
        className="text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
      >
        &larr; Изменить API Login
      </button>

      <label className="block">
        <span className="text-sm font-medium text-neutral-700">
          Организация
        </span>
        <select
          value={config.org_id || ''}
          onChange={(e) =>
            setConfig((prev) => ({ ...prev, org_id: e.target.value }))
          }
          className={inputClass}
        >
          <option value="" disabled>
            Выберите организацию
          </option>
          {discovery.organizations.map((org) => (
            <option key={org.id} value={org.id}>
              {org.name}
            </option>
          ))}
        </select>
      </label>

      {discovery.external_menus.length > 0 && (
        <label className="block">
          <span className="text-sm font-medium text-neutral-700">
            Внешнее меню{' '}
            <span className="text-neutral-400">(необязательно)</span>
          </span>
          <select
            value={config.external_menu_id || ''}
            onChange={(e) =>
              setConfig((prev) => ({
                ...prev,
                external_menu_id: e.target.value,
              }))
            }
            className={inputClass}
          >
            <option value="">Номенклатура (по умолчанию)</option>
            {discovery.external_menus.map((menu) => (
              <option key={menu.id} value={menu.id}>
                {menu.name}
              </option>
            ))}
          </select>
        </label>
      )}

      {createError && (
        <p className="text-sm text-red-600">
          Ошибка при создании интеграции. Попробуйте снова.
        </p>
      )}

      <div className="flex gap-3 pt-2">
        <button
          type="button"
          onClick={onCancel}
          className="flex-1 py-2.5 px-4 rounded-lg border border-surface-border text-sm font-medium text-neutral-700 hover:bg-neutral-50 transition-colors"
        >
          Отмена
        </button>
        <button
          type="submit"
          disabled={creating || !config.org_id}
          className={cn(
            'flex-1 py-2.5 px-4 rounded-lg text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover active:bg-accent/80',
            'disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-150',
          )}
        >
          {creating ? 'Создание...' : 'Создать'}
        </button>
      </div>
    </form>
  )
}
