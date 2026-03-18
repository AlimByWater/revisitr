import { createFileRoute } from '@tanstack/react-router'
import { LayoutDashboard } from 'lucide-react'

export const Route = createFileRoute('/dashboard/')({
  component: DashboardHome,
})

function DashboardHome() {
  return (
    <div className="max-w-4xl">
      <h1 className="text-2xl font-bold text-neutral-900 mb-2">
        Добро пожаловать в Revisitr
      </h1>
      <p className="text-neutral-500 mb-8">
        Управляйте программой лояльности вашего ресторана
      </p>

      <div className="bg-white rounded-2xl border border-surface-border p-12 text-center">
        <div className="w-16 h-16 rounded-2xl bg-neutral-100 flex items-center justify-center mx-auto mb-4">
          <LayoutDashboard className="w-8 h-8 text-neutral-400" />
        </div>
        <h2 className="text-lg font-semibold text-neutral-700 mb-2">
          Здесь будут виджеты дашборда
        </h2>
        <p className="text-sm text-neutral-400 max-w-md mx-auto">
          Статистика по клиентам, продажам, активности программ лояльности и
          эффективности рассылок
        </p>
      </div>
    </div>
  )
}
