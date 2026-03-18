import { cn } from '@/lib/utils'

interface SkeletonProps {
  className?: string
}

function Bone({ className }: SkeletonProps) {
  return <div className={cn('shimmer rounded', className)} />
}

export function CardSkeleton() {
  return (
    <div className="bg-white rounded-2xl border border-surface-border p-6 space-y-4">
      <div className="flex items-center gap-3">
        <Bone className="w-10 h-10 rounded-xl" />
        <div className="space-y-2 flex-1">
          <Bone className="h-4 w-32" />
          <Bone className="h-3 w-24" />
        </div>
      </div>
      <Bone className="h-3 w-full" />
      <Bone className="h-3 w-3/4" />
    </div>
  )
}

export function MetricSkeleton() {
  return (
    <div className="bg-white rounded-2xl border border-surface-border p-6">
      <div className="flex items-center gap-2 mb-3">
        <Bone className="w-4 h-4 rounded" />
        <Bone className="h-3 w-16" />
      </div>
      <Bone className="h-8 w-28 mb-2" />
      <Bone className="h-4 w-16" />
    </div>
  )
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="bg-white rounded-2xl border border-surface-border overflow-hidden">
      {/* Header */}
      <div className="px-6 py-4 border-b border-surface-border flex gap-6">
        <Bone className="h-3 w-20" />
        <Bone className="h-3 w-24" />
        <Bone className="h-3 w-16" />
        <Bone className="h-3 w-20" />
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          className="px-6 py-4 border-b border-surface-border last:border-0 flex items-center gap-6"
          style={{ animationDelay: `${i * 0.05}s` }}
        >
          <Bone className="w-8 h-8 rounded-full" />
          <Bone className="h-4 w-28" />
          <Bone className="h-4 w-20" />
          <Bone className="h-4 w-16" />
          <Bone className="h-4 w-24 ml-auto" />
        </div>
      ))}
    </div>
  )
}

export function ChartSkeleton() {
  return (
    <div className="h-[240px] rounded-xl flex items-end gap-1.5 px-4 pb-4 pt-8 relative overflow-hidden">
      {/* Fake grid lines */}
      <div className="absolute inset-x-0 top-8 bottom-4 flex flex-col justify-between px-4">
        {[0, 1, 2, 3].map((i) => (
          <div key={i} className="border-t border-dashed border-neutral-100" />
        ))}
      </div>
      {/* Fake bars */}
      {[40, 65, 45, 80, 55, 70, 50, 85, 60, 75, 45, 90].map((h, i) => (
        <div
          key={i}
          className="flex-1 shimmer rounded-t"
          style={{ height: `${h}%`, animationDelay: `${i * 0.08}s` }}
        />
      ))}
    </div>
  )
}

export function PageSkeleton() {
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <Bone className="h-8 w-40" />
        <Bone className="h-10 w-36 rounded-lg" />
      </div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[0, 1, 2, 3].map((i) => (
          <MetricSkeleton key={i} />
        ))}
      </div>
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {[0, 1].map((i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    </div>
  )
}
