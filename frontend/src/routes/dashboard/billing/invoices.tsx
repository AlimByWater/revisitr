import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useInvoicesQuery } from '@/features/billing/queries'
import { ArrowLeft, FileText } from 'lucide-react'

function formatPrice(kopeks: number): string {
  return `${(kopeks / 100).toLocaleString('ru-RU')} ₽`
}

function InvoiceStatusBadge({ status }: { status: string }) {
  const config: Record<string, { label: string; className: string }> = {
    pending: { label: 'Ожидает', className: 'bg-amber-500/10 text-amber-400' },
    paid: { label: 'Оплачен', className: 'bg-green-500/10 text-green-400' },
    failed: { label: 'Ошибка', className: 'bg-red-500/10 text-red-400' },
    refunded: { label: 'Возврат', className: 'bg-blue-500/10 text-blue-400' },
  }
  const c = config[status] ?? { label: status, className: 'bg-neutral-500/10 text-neutral-400' }
  return (
    <span className={cn('text-xs font-medium px-2 py-0.5 rounded-full', c.className)}>
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
          className="text-white/40 hover:text-white transition-colors"
        >
          <ArrowLeft className="w-5 h-5" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-white">Счета</h1>
          <p className="text-sm text-white/40 mt-1">История выставленных счетов</p>
        </div>
      </div>

      {isLoading ? (
        <div className="space-y-3">
          {[...Array(5)].map((_, i) => (
            <div key={i} className="rounded-lg border border-white/10 bg-white/[0.02] p-4 h-16 animate-pulse" />
          ))}
        </div>
      ) : !invoices?.length ? (
        <div className="flex flex-col items-center justify-center py-16 text-white/30">
          <FileText className="w-12 h-12 mb-4" />
          <p className="text-lg font-medium">Счетов пока нет</p>
          <p className="text-sm mt-1">Счета появятся после выбора платного тарифа</p>
        </div>
      ) : (
        <div className="rounded-xl border border-white/10 overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-white/10 bg-white/[0.02]">
                <th className="text-left text-white/40 font-medium px-4 py-3">Номер</th>
                <th className="text-left text-white/40 font-medium px-4 py-3">Дата</th>
                <th className="text-left text-white/40 font-medium px-4 py-3">Сумма</th>
                <th className="text-left text-white/40 font-medium px-4 py-3">Срок оплаты</th>
                <th className="text-left text-white/40 font-medium px-4 py-3">Статус</th>
              </tr>
            </thead>
            <tbody>
              {invoices.map((inv) => (
                <tr key={inv.id} className="border-b border-white/[0.05] hover:bg-white/[0.02] transition-colors">
                  <td className="px-4 py-3 text-white font-mono">#{inv.id}</td>
                  <td className="px-4 py-3 text-white/60">
                    {new Date(inv.created_at).toLocaleDateString('ru-RU')}
                  </td>
                  <td className="px-4 py-3 text-white font-medium">{formatPrice(inv.amount)}</td>
                  <td className="px-4 py-3 text-white/60">
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
