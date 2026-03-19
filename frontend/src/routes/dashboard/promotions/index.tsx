import { Tag } from 'lucide-react'

export default function PromotionsPage() {
  return (
    <div>
      <div className="mb-6">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Акции</h1>
        <p className="text-sm text-neutral-500 mt-1">Управление акциями и специальными предложениями</p>
      </div>

      <div className="flex flex-col items-center justify-center py-24 text-center">
        <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mb-4">
          <Tag className="w-8 h-8 text-neutral-400" />
        </div>
        <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">Скоро</h3>
        <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
          Акции и специальные предложения появятся в следующей версии
        </p>
      </div>
    </div>
  )
}
