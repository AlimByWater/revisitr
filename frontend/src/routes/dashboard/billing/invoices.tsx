import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useInvoicesQuery } from '@/features/billing/queries'
import { ArrowLeft, FileText } from 'lucide-react'

function formatPrice(kopeks: number): string {
  return `${(kopeks / 100).toLocaleString('ru-RU')} ₽`
}

function InvoiceStatusBadge({ status }: { status: string }) {
  const config: Record<string, { label: string; className: string }> = {
    pending: { label: 'Ожидает', className: 'bg-amber-500/10 text-amber-700 border-amber-500/30' },
    paid: { label: 'Оплачен', className: 'bg-emerald-500/10 text-emerald-700 border-emerald-500/30' },
    failed: { label: 'Ошибка', className: 'bg-red-500/10 text-red-700 border-red-500/30' },
    refunded: { label: 'Возврат', className: 'bg-neutral-100 text-neutral-600 border-neutral-300' },
  }
  const c = config[status] ?? { label: status, className: 'bg-neutral-100 text-neutral-600 border-neutral-300' }
  return (
    <span className={cn('font-mono text-[10px] uppercase tracking-wider px-2 py-0.5 rounded border', c.className)}>
      {c.label}
    </span>
  )
}

export default function InvoicesPage() {
  const { data: invoices, isLoading } = useInvoicesQuery()

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4 animate-in">
        <Link
          to="/dashboard/billing"
          className="text-neutral-400 hover:text-neutral-900 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="font-display text-3xl font-bold text-neutral-900 tracking-tight">Счета</h1>
          <p className="text-xs text-neutral-400 uppercase tracking-wider mt-1">История выставленных счетов</p>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="rounded border border-neutral-900 bg-white p-4 h-16 animate-pulse" />
          ))}
        </div>
      ) : !invoices?.length ? (
        <div className="flex flex-col items-center justify-center py-24 text-center">
          <div className="w-16 h-16 rounded bg-neutral-100 flex items-center justify-center mb-4">
            <FileText className="w-8 h-8 text-neutral-400" />
          </div>
          <h3 className="font-display text-xl font-bold text-neutral-800 mb-1.5 tracking-tight">Счетов пока нет</h3>
          <p className="text-sm text-neutral-400 max-w-xs leading-relaxed">Счета появятся после выбора платного тарифа</p>
        </div>
      ) : (
        <div className="rounded border border-neutral-900 bg-white overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 bg-neutral-50">
                <th className="text-center text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">Номер</th>
                <th className="text-center text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">Дата</th>
                <th className="text-center text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">Сумма</th>
                <th className="text-center text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">Срок оплаты</th>
                <th className="text-center text-xs font-medium text-neutral-500 uppercase tracking-wider px-4 py-3">Статус</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50 transition-colors">
                  <td className="px-4 py-3 text-center text-neutral-900 font-mono">#{inv.id}</td>
                  <td className="px-4 py-3 text-center text-neutral-500 font-mono tabular-nums">
                    {new Date(inv.created_at).toLocaleDateString('ru-RU')}
                  </td>
                  <td className="px-4 py-3 text-center text-neutral-900 font-medium font-mono tabular-nums">{formatPrice(inv.amount)}</td>
                  <td className="px-4 py-3 text-center text-neutral-500 font-mono tabular-nums">
                    {new Date(inv.due_date).toLocaleDateString('ru-RU')}
                  </td>
                  <td className="px-4 py-3 text-center">
                    <InvoiceStatusBadge status={inv.status} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
