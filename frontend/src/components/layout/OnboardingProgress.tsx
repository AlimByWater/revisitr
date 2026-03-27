import { useState } from 'react'
import { Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useOnboardingQuery } from '@/features/onboarding/queries'
import { X } from 'lucide-react'

const ACTION_STEPS = ['loyalty', 'bot', 'pos', 'integrations'] as const

export function OnboardingProgress({ isAurora = false }: { isAurora?: boolean }) {
  const { data } = useOnboardingQuery()
  const [dismissed, setDismissed] = useState(false)

  if (dismissed) return null
  if (!data) return null
  if (data.onboarding_completed) return null

  const steps = data.onboarding_state?.steps ?? {}
  const completed = ACTION_STEPS.filter((k) => steps[k]?.completed).length
  const total = ACTION_STEPS.length
  const progress = total > 0 ? (completed / total) * 100 : 0

  return (
    <div className={cn(
      'mx-3 mb-2 rounded-xl p-3 relative',
      isAurora
        ? 'bg-white/[0.04] border border-white/[0.06]'
        : 'bg-accent/5 border border-accent/10',
    )}>
      <button
        type="button"
        onClick={() => setDismissed(true)}
        className={cn(
          'absolute top-2 right-2 w-5 h-5 rounded flex items-center justify-center transition-colors',
          isAurora
            ? 'text-white/20 hover:text-white/50 hover:bg-white/10'
            : 'text-neutral-400 hover:text-neutral-600 hover:bg-neutral-100',
        )}
        aria-label="Закрыть"
      >
        <X className="w-3 h-3" />
      </button>

      <Link
        to="/dashboard/onboarding"
        className="block"
      >
        <p className={cn(
          'text-xs font-semibold mb-2',
          isAurora ? 'text-white/60' : 'text-neutral-700',
        )}>
          Настройка {completed}/{total}
        </p>
        <div className={cn(
          'h-1.5 rounded-full overflow-hidden',
          isAurora ? 'bg-white/10' : 'bg-neutral-200',
        )}>
          <div
            className={cn(
              'h-full rounded-full transition-all duration-500',
              isAurora ? 'bg-violet-400' : 'bg-accent',
            )}
            style={{ width: `${progress}%` }}
          />
        </div>
        <p className={cn(
          'text-[10px] mt-1.5',
          isAurora ? 'text-white/30' : 'text-neutral-400',
        )}>
          Продолжить настройку →
        </p>
      </Link>
    </div>
  )
}
