import { Users } from 'lucide-react'

export default function SegmentsPage() {
  return (
    <div>
      <div className="mb-6">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Сегментация</h1>
        <p className="text-sm text-neutral-500 mt-1">Группировка клиентов по параметрам</p>
      </div>

      <div className="flex flex-col items-center justify-center py-24 text-center">
        <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mb-4">
          <Users className="w-8 h-8 text-neutral-400" />
        </div>
        <h3 className="font-serif text-xl font-bold text-neutral-800 mb-1.5">Скоро</h3>
        <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">
          Сегментация клиентов по RFM, тегам и поведению появится в следующей версии
        </p>
      </div>
    </div>
  )
}
