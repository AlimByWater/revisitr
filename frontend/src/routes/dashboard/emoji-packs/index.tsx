import { useState, useRef } from 'react'
import { Plus, Trash2, X, Smile } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  useEmojiPacksQuery,
  useCreateEmojiPackMutation,
  useUpdateEmojiPackMutation,
  useDeleteEmojiPackMutation,
  useAddEmojiItemMutation,
  useDeleteEmojiItemMutation,
} from '@/features/emoji-packs/queries'
import { campaignsApi } from '@/features/campaigns/api'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import { ErrorState } from '@/components/common/ErrorState'

const inputClassName = cn(
  'px-3 py-2 rounded border border-neutral-200 text-sm bg-white',
  'focus:outline-none focus:ring-2 focus:ring-accent/20 focus:border-accent',
  'transition-colors',
)

interface AddItemFormProps {
  packId: number
  onDone: () => void
}

function AddItemForm({ packId, onDone }: AddItemFormProps) {
  const [name, setName] = useState('')
  const [uploading, setUploading] = useState(false)
  const [imageUrl, setImageUrl] = useState('')
  const fileRef = useRef<HTMLInputElement>(null)
  const addItem = useAddEmojiItemMutation(packId)

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (!file) return
    setUploading(true)
    try {
      const url = await campaignsApi.uploadFile(file)
      setImageUrl(url)
    } finally {
      setUploading(false)
    }
  }

  const handleSubmit = async () => {
    if (!name.trim() || !imageUrl) return
    await addItem.mutate({ name: name.trim(), image_url: imageUrl })
    onDone()
  }

  return (
    <div className="mt-3 p-3 rounded border border-neutral-200 bg-neutral-50 space-y-2">
      <div className="text-xs font-medium text-neutral-500 uppercase tracking-wider">Новый эмодзи</div>
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="Название"
        className={cn(inputClassName, 'w-full')}
      />
      <div className="flex items-center gap-2">
        {imageUrl ? (
          <img src={imageUrl} alt="preview" className="w-10 h-10 rounded border border-neutral-200 object-cover" />
        ) : (
          <div className="w-10 h-10 rounded border border-dashed border-neutral-300 flex items-center justify-center bg-white">
            <Smile className="w-5 h-5 text-neutral-300" />
          </div>
        )}
        <button
          type="button"
          onClick={() => fileRef.current?.click()}
          disabled={uploading}
          className={cn(
            'px-3 py-1.5 rounded border border-neutral-300 text-xs font-medium text-neutral-600',
            'hover:bg-neutral-100 transition-colors disabled:opacity-50',
          )}
        >
          {uploading ? 'Загрузка...' : 'Выбрать файл'}
        </button>
        <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handleFileChange} />
      </div>
      <div className="flex gap-2 pt-1">
        <button
          type="button"
          onClick={handleSubmit}
          disabled={addItem.isPending || !name.trim() || !imageUrl}
          className={cn(
            'py-1.5 px-3 rounded text-xs font-medium bg-accent text-white hover:bg-accent-hover',
            'disabled:opacity-50 transition-all',
          )}
        >
          {addItem.isPending ? 'Добавление...' : 'Добавить'}
        </button>
        <button
          type="button"
          onClick={onDone}
          className="py-1.5 px-3 rounded text-xs text-neutral-500 hover:text-neutral-700 transition-colors"
        >
          Отмена
        </button>
      </div>
    </div>
  )
}

interface PackCardProps {
  pack: {
    id: number
    name: string
    sort_order: number
    items?: Array<{ id: number; name: string; image_url: string }>
  }
  onDelete: (id: number) => void
  onMutate: () => void
}

function PackCard({ pack, onDelete, onMutate }: PackCardProps) {
  const [editName, setEditName] = useState(pack.name)
  const [showAddForm, setShowAddForm] = useState(false)
  const [deletingItem, setDeletingItem] = useState<number | null>(null)
  const updatePack = useUpdateEmojiPackMutation()
  const deleteItem = useDeleteEmojiItemMutation()

  const handleNameBlur = async () => {
    if (editName.trim() && editName.trim() !== pack.name) {
      await updatePack.mutate({ id: pack.id, data: { name: editName.trim() } })
      onMutate()
    } else {
      setEditName(pack.name)
    }
  }

  const handleDeleteItem = async (itemId: number) => {
    setDeletingItem(itemId)
    try {
      await deleteItem.mutate(itemId)
      onMutate()
    } finally {
      setDeletingItem(null)
    }
  }

  const items = pack.items ?? []

  return (
    <div className="bg-white rounded border border-neutral-900 p-5 group">
      <div className="flex items-center justify-between mb-4">
        <input
          type="text"
          value={editName}
          onChange={(e) => setEditName(e.target.value)}
          onBlur={handleNameBlur}
          className={cn(
            'flex-1 text-base font-semibold text-neutral-900 bg-transparent border-b border-transparent',
            'hover:border-neutral-300 focus:border-accent focus:outline-none transition-colors py-0.5',
          )}
        />
        <button
          type="button"
          onClick={() => onDelete(pack.id)}
          className="ml-3 p-1.5 rounded text-neutral-300 hover:text-red-500 hover:bg-red-50 opacity-0 group-hover:opacity-100 transition-all"
          aria-label="Удалить пак"
        >
          <Trash2 className="w-4 h-4" />
        </button>
      </div>

      {items.length > 0 ? (
        <div className="flex flex-wrap gap-2 mb-4">
          {items.map((item) => (
            <div key={item.id} className="relative group/item">
              <div className="w-12 h-12 rounded border border-neutral-200 overflow-hidden">
                <img
                  src={item.image_url}
                  alt={item.name}
                  className="w-full h-full object-cover"
                  title={item.name}
                />
              </div>
              <button
                type="button"
                onClick={() => handleDeleteItem(item.id)}
                disabled={deletingItem === item.id}
                className={cn(
                  'absolute -top-1.5 -right-1.5 w-4 h-4 rounded-full bg-white border border-neutral-300',
                  'flex items-center justify-center text-neutral-400 hover:text-red-500 hover:border-red-300',
                  'opacity-0 group-hover/item:opacity-100 transition-all shadow-sm',
                  deletingItem === item.id && 'opacity-50',
                )}
                aria-label={`Удалить ${item.name}`}
              >
                <X className="w-2.5 h-2.5" />
              </button>
            </div>
          ))}
        </div>
      ) : (
        <p className="text-sm text-neutral-400 mb-4">Нет эмодзи в этом паке</p>
      )}

      {showAddForm ? (
        <AddItemForm
          packId={pack.id}
          onDone={() => { setShowAddForm(false); onMutate() }}
        />
      ) : (
        <button
          type="button"
          onClick={() => setShowAddForm(true)}
          className="flex items-center gap-1.5 text-xs font-medium text-neutral-500 hover:text-accent transition-colors"
        >
          <Plus className="w-3.5 h-3.5" />
          Добавить
        </button>
      )}
    </div>
  )
}

export default function EmojiPacksPage() {
  const { data: packs, isLoading, isError, mutate } = useEmojiPacksQuery()
  const createPack = useCreateEmojiPackMutation()
  const deletePack = useDeleteEmojiPackMutation()
  const [showCreate, setShowCreate] = useState(false)
  const [newName, setNewName] = useState('')

  const handleCreate = async () => {
    if (!newName.trim()) return
    await createPack.mutate({ name: newName.trim() })
    setNewName('')
    setShowCreate(false)
    mutate()
  }

  const handleDelete = async (id: number) => {
    await deletePack.mutate(id)
    mutate()
  }

  if (isLoading) {
    return (
      <div>
        <div className="h-8 w-48 shimmer rounded mb-6" />
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[0, 1, 2].map((i) => <CardSkeleton key={i} />)}
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <ErrorState
        title="Ошибка загрузки"
        message="Не удалось загрузить список эмодзи-паков."
        onRetry={() => mutate()}
      />
    )
  }

  return (
    <div>
      <div className="flex items-start justify-between mb-6">
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Эмодзи-паки</h1>
          <p className="font-mono text-xs text-neutral-400 uppercase tracking-wider mt-1">
            Иконки для категорий меню
          </p>
        </div>
        <button
          type="button"
          onClick={() => setShowCreate(true)}
          className={cn(
            'flex items-center gap-1.5 py-2 px-4 rounded text-sm font-medium',
            'bg-accent text-white hover:bg-accent-hover',
            'transition-all',
          )}
        >
          <Plus className="w-4 h-4" />
          Создать пак
        </button>
      </div>

      {showCreate && (
        <div className="bg-white rounded border border-neutral-900 p-5 mb-6 animate-in">
          <h3 className="text-sm font-semibold text-neutral-900 mb-3">Новый пак</h3>
          <div className="flex gap-3">
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="Название пака"
              className={cn(inputClassName, 'flex-1')}
              onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
            />
            <button
              type="button"
              onClick={handleCreate}
              disabled={createPack.isPending || !newName.trim()}
              className={cn(
                'py-2 px-4 rounded text-sm font-medium',
                'bg-accent text-white hover:bg-accent-hover',
                'disabled:opacity-50 transition-all',
              )}
            >
              {createPack.isPending ? 'Создание...' : 'Создать'}
            </button>
            <button
              type="button"
              onClick={() => { setShowCreate(false); setNewName('') }}
              className="py-2 px-3 rounded text-sm text-neutral-500 hover:text-neutral-700 transition-colors"
            >
              Отмена
            </button>
          </div>
        </div>
      )}

      {!packs || packs.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
            <Smile className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">Нет эмодзи-паков</h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
            Создайте пак и добавьте иконки для категорий меню
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {packs.map((pack) => (
            <PackCard
              key={pack.id}
              pack={pack}
              onDelete={handleDelete}
              onMutate={() => mutate()}
            />
          ))}
        </div>
      )}
    </div>
  )
}
