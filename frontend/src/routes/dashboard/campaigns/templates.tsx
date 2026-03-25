import { useState } from 'react'
import { cn } from '@/lib/utils'
import {
  useCampaignTemplatesQuery,
  useCreateCampaignTemplateMutation,
  useDeleteCampaignTemplateMutation,
  useCreateFromTemplateMutation,
} from '@/features/campaigns/queries'
import { FileText, Plus, Trash2, Copy, Lock } from 'lucide-react'
import { EmptyState } from '@/components/common/EmptyState'
import { ErrorState } from '@/components/common/ErrorState'
import { TableSkeleton } from '@/components/common/LoadingSkeleton'
import type { CampaignTemplate } from '@/features/campaigns/types'
import { useNavigate } from 'react-router-dom'

const categoryConfig: Record<string, { label: string; className: string }> = {
  general: { label: 'Общий', className: 'bg-neutral-100 text-neutral-600' },
  welcome: { label: 'Приветствие', className: 'bg-blue-100 text-blue-700' },
  promo: { label: 'Промо', className: 'bg-orange-100 text-orange-700' },
  holiday: { label: 'Праздник', className: 'bg-purple-100 text-purple-700' },
  reactivation: { label: 'Реактивация', className: 'bg-red-100 text-red-700' },
}

export default function CampaignTemplatesPage() {
  const navigate = useNavigate()
  const { data: templates, isLoading, error } = useCampaignTemplatesQuery()
  const createTemplate = useCreateCampaignTemplateMutation()
  const deleteTemplate = useDeleteCampaignTemplateMutation()
  const createFromTemplate = useCreateFromTemplateMutation()

  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')
  const [newMessage, setNewMessage] = useState('')
  const [newCategory, setNewCategory] = useState('general')

  if (error) return <ErrorState message="Ошибка загрузки шаблонов" />
  if (isLoading) return <TableSkeleton rows={4} />

  const handleCreate = async () => {
    if (!newName || !newMessage) return
    await createTemplate.trigger({
      name: newName,
      message: newMessage,
      category: newCategory,
    })
    setShowCreate(false)
    setNewName('')
    setNewMessage('')
    setNewCategory('general')
  }

  const handleUseTemplate = async (template: CampaignTemplate) => {
    // For now, navigate to create page; in full implementation would show bot selector
    const botId = 0 // TODO: prompt user for bot selection
    if (botId > 0) {
      const campaign = await createFromTemplate.trigger({ templateId: template.id, botId })
      if (campaign) {
        navigate(`/dashboard/campaigns/${campaign.id}`)
      }
    }
  }

  const systemTemplates = templates?.filter((t) => t.is_system) || []
  const userTemplates = templates?.filter((t) => !t.is_system) || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Шаблоны рассылок</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Готовые шаблоны для быстрого создания рассылок
          </p>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="inline-flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90"
        >
          <Plus className="h-4 w-4" />
          Создать шаблон
        </button>
      </div>

      {showCreate && (
        <div className="rounded-lg border bg-card p-4 space-y-3">
          <h3 className="font-medium">Новый шаблон</h3>
          <input
            type="text"
            placeholder="Название шаблона"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            className="w-full rounded-md border px-3 py-2 text-sm"
          />
          <select
            value={newCategory}
            onChange={(e) => setNewCategory(e.target.value)}
            className="w-full rounded-md border px-3 py-2 text-sm"
          >
            {Object.entries(categoryConfig).map(([key, { label }]) => (
              <option key={key} value={key}>{label}</option>
            ))}
          </select>
          <textarea
            placeholder="Текст сообщения. Используйте {name} для имени клиента"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            rows={4}
            className="w-full rounded-md border px-3 py-2 text-sm"
          />
          <div className="flex gap-2">
            <button
              onClick={handleCreate}
              disabled={!newName || !newMessage}
              className="rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
            >
              Создать
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="rounded-md border px-4 py-2 text-sm font-medium hover:bg-accent"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {/* System templates */}
      {systemTemplates.length > 0 && (
        <div className="space-y-3">
          <h2 className="text-lg font-semibold">Системные шаблоны</h2>
          <div className="grid gap-3 sm:grid-cols-2">
            {systemTemplates.map((template) => (
              <TemplateCard
                key={template.id}
                template={template}
                onUse={() => handleUseTemplate(template)}
              />
            ))}
          </div>
        </div>
      )}

      {/* User templates */}
      <div className="space-y-3">
        <h2 className="text-lg font-semibold">Мои шаблоны</h2>
        {userTemplates.length === 0 ? (
          <EmptyState
            icon={FileText}
            title="Нет шаблонов"
            description="Создайте свой первый шаблон рассылки"
          />
        ) : (
          <div className="grid gap-3 sm:grid-cols-2">
            {userTemplates.map((template) => (
              <TemplateCard
                key={template.id}
                template={template}
                onUse={() => handleUseTemplate(template)}
                onDelete={() => deleteTemplate.trigger(template.id)}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  )
}

function TemplateCard({
  template,
  onUse,
  onDelete,
}: {
  template: CampaignTemplate
  onUse: () => void
  onDelete?: () => void
}) {
  const cat = categoryConfig[template.category] || categoryConfig.general

  return (
    <div className="rounded-lg border bg-card p-4 space-y-3">
      <div className="flex items-start justify-between">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h3 className="font-medium">{template.name}</h3>
            {template.is_system && <Lock className="h-3.5 w-3.5 text-muted-foreground" />}
          </div>
          <span className={cn('inline-block rounded-full px-2 py-0.5 text-xs font-medium', cat.className)}>
            {cat.label}
          </span>
        </div>
      </div>
      {template.description && (
        <p className="text-sm text-muted-foreground">{template.description}</p>
      )}
      <p className="text-sm line-clamp-2 text-muted-foreground italic">
        {template.message}
      </p>
      <div className="flex gap-2 pt-1">
        <button
          onClick={onUse}
          className="inline-flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs font-medium hover:bg-accent"
        >
          <Copy className="h-3.5 w-3.5" />
          Использовать
        </button>
        {onDelete && (
          <button
            onClick={onDelete}
            className="inline-flex items-center gap-1.5 rounded-md border border-red-200 px-3 py-1.5 text-xs font-medium text-red-600 hover:bg-red-50"
          >
            <Trash2 className="h-3.5 w-3.5" />
            Удалить
          </button>
        )}
      </div>
    </div>
  )
}
