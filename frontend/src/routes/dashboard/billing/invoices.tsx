import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useInvoicesQuery } from '@/features/billing/queries'
import { ArrowLeft, FileText } from 'lucide-react'

function formatPrice(kopeks: number): string {
  return `${(kopeks / 100).toLocaleString('ru-RU')} ₽`
}

function InvoiceStatusBadge({ status }: { status: string }) {
  const config: Record<string, { label: string; className: string }> = {
    pending: { label: 'Ожидает', className: 'bg-amber-50 text-amber-600' },
    paid: { label: 'Оплачен', className: 'bg-green-50 text-green-600' },
    failed: { label: 'Ошибка', className: 'bg-red-50 text-red-600' },
    refunded: { label: 'Возврат', className: 'bg-blue-50 text-blue-600' },
  }
  const c = config[status] ?? { label: status, className: 'bg-neutral-100 text-neutral-500' }
  return (
    <span className={cn('text-xs font-medium px-2.5 py-1 rounded-full', c.className)}>
      {c.label}
    </span>
  )
}

export default function InvoicesPage() {
  const { data: invoices, isLoading } = useInvoicesQuery()

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/dashboard/billing"
          className="text-neutral-400 hover:text-neutral-900 transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">Счета</h1>
          <p className="text-sm text-neutral-500 mt-1">История выставленных счетов</p>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="rounded-2xl border border-surface-border bg-white p-4 h-16 animate-pulse" />
          ))}
        </div>
      ) : !invoices?.length ? (
        <div className="flex flex-col items-center justify-center py-16 text-neutral-300">
          <FileText className="w-12 h-12 mb-4" />
          <p className="text-lg font-medium text-neutral-400">Счетов пока нет</p>
          <p className="text-sm text-neutral-400 mt-1">Счета появятся после выбора платного тарифа</p>
        </div>
      ) : (
        <div className="rounded-2xl border border-surface-border bg-white overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-neutral-100 bg-neutral-50">
                <th className="text-left text-neutral-500 font-medium px-4 py-3">Номер</th>
                <th className="text-left text-neutral-500 font-medium px-4 py-3">Дата</th>
                <th className="text-left text-neutral-500 font-medium px-4 py-3">Сумма</th>
                <th className="text-left text-neutral-500 font-medium px-4 py-3">Срок оплаты</th>
                <th className="text-left text-neutral-500 font-medium px-4 py-3">Статус</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id} className="border-b border-neutral-100 last:border-0 hover:bg-neutral-50 transition-colors">
                  <td className="px-4 py-3 text-neutral-900 font-mono">#{inv.id}</td>
                  <td className="px-4 py-3 text-neutral-500">
                    {new Date(inv.created_at).toLocaleDateString('ru-RU')}
                  </td>
                  <td className="px-4 py-3 text-neutral-900 font-medium">{formatPrice(inv.amount)}</td>
                  <td className="px-4 py-3 text-neutral-500">
                    {new Date(inv.due_date).toLocaleDateString('ru-RU')}
                  </td>
                  <td className="px-4 py-3">
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
