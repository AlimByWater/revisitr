import { useState, useEffect } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { cn } from '@/lib/utils'
import {
  useOnboardingQuery,
  useUpdateOnboardingMutation,
  useCompleteOnboardingMutation,
} from '@/features/onboarding/queries'
import { ONBOARDING_STEPS } from '@/features/onboarding/types'
import { CardSkeleton } from '@/components/common/LoadingSkeleton'
import {
  CheckCircle2,
  Circle,
  ArrowRight,
  ArrowLeft,
  Sparkles,
  Bot,
  Heart,
  Store,
  Workflow,
  Rocket,
  Mail,
  TrendingUp,
  Tag,
  Users,
} from 'lucide-react'

const STEP_ICONS = [Sparkles, Bot, Heart, Store, Workflow, Rocket]

export default function OnboardingPage() {
  const navigate = useNavigate()
  const { data, isLoading, mutate } = useOnboardingQuery()
  const updateStep = useUpdateOnboardingMutation()
  const complete = useCompleteOnboardingMutation()

  const [currentStep, setCurrentStep] = useState(0)

  useEffect(() => {
    if (data?.onboarding_state?.current_step !== undefined) {
      const maxIndex = ONBOARDING_STEPS.length - 1
      setCurrentStep(Math.min(data.onboarding_state.current_step, maxIndex))
    }
  }, [data])

  useEffect(() => {
    if (data?.onboarding_completed) {
      navigate('/dashboard')
    }
  }, [data?.onboarding_completed, navigate])

  if (isLoading) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="h-8 w-64 shimmer rounded mb-4" />
        <CardSkeleton />
      </div>
    )
  }

  const state = data?.onboarding_state
  const steps = ONBOARDING_STEPS

  const handleNext = async () => {
    const stepKey = steps[currentStep]?.key
    if (!stepKey) return

    await updateStep.mutate({
      step: stepKey,
      completed: true,
      skipped: false,
    })

    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1)
    }
    mutate()
  }

  const handleSkip = async () => {
    const stepKey = steps[currentStep]?.key
    if (!stepKey) return

    await updateStep.mutate({
      step: stepKey,
      completed: false,
      skipped: true,
    })

    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1)
    }
    mutate()
  }

  const handlePrev = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1)
    }
  }

  const handleComplete = async () => {
    await complete.mutate(undefined)
    navigate('/dashboard')
  }

  const isStepCompleted = (key: string) => {
    return state?.steps?.[key]?.completed ?? false
  }

  const isLastStep = currentStep === steps.length - 1
  const allCompleted = steps.every((s) => isStepCompleted(s.key))

  return (
    <div className="max-w-2xl mx-auto">
      <div className="text-center mb-8 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight">
          Добро пожаловать в Revisitr
        </h1>
        <p className="text-sm text-neutral-500 mt-2">
          Настройте платформу за несколько шагов
        </p>
      </div>

      {/* Progress bar */}
      <div className="flex items-center gap-1 mb-8 animate-in animate-in-delay-1">
        {steps.map((step, i) => {
          const completed = isStepCompleted(step.key)
          const active = i === currentStep
          return (
            <button
              key={step.key}
              type="button"
              onClick={() => setCurrentStep(i)}
              className="flex-1 group"
            >
              <div
                className={cn(
                  'h-1.5 rounded-full transition-all',
                  completed
                    ? 'bg-green-500'
                    : active
                      ? 'bg-accent'
                      : 'bg-neutral-200 group-hover:bg-neutral-300',
                )}
              />
              <p
                className={cn(
                  'text-[10px] mt-1.5 font-medium text-center truncate',
                  active ? 'text-neutral-900' : 'text-neutral-400',
                )}
              >
                {step.title}
              </p>
            </button>
          )
        })}
      </div>

      {/* Current step card */}
      <div className="bg-white rounded-2xl border border-surface-border p-8 animate-in animate-in-delay-2">
        <StepContent
          step={steps[currentStep]}
          stepIndex={currentStep}
          isCompleted={isStepCompleted(steps[currentStep]?.key)}
        />

        <div className="flex items-center justify-between mt-8 pt-6 border-t border-surface-border">
          <button
            type="button"
            onClick={handlePrev}
            disabled={currentStep === 0}
            className={cn(
              'flex items-center gap-1.5 py-2 px-4 rounded-lg text-sm font-medium',
              'text-neutral-600 hover:bg-neutral-50',
              'disabled:opacity-30 disabled:cursor-not-allowed',
              'transition-colors',
            )}
          >
            <ArrowLeft className="w-4 h-4" />
            Назад
          </button>

          <button
            type="button"
            onClick={handleSkip}
            disabled={updateStep.isPending}
            className="text-sm text-neutral-400 hover:text-neutral-600 transition-colors"
          >
            Пропустить
          </button>

          {isLastStep && allCompleted ? (
            <button
              type="button"
              onClick={handleComplete}
              disabled={complete.isPending}
              className={cn(
                'flex items-center gap-1.5 py-2.5 px-6 rounded-xl text-sm font-semibold',
                'bg-green-500 text-white hover:bg-green-600',
                'disabled:opacity-50 transition-all',
                'shadow-sm shadow-green-500/20',
              )}
            >
              {complete.isPending ? 'Завершение...' : 'Начать работу'}
              <Rocket className="w-4 h-4" />
            </button>
          ) : (
            <button
              type="button"
              onClick={handleNext}
              disabled={updateStep.isPending}
              className={cn(
                'flex items-center gap-1.5 py-2.5 px-6 rounded-xl text-sm font-semibold',
                'bg-accent text-white hover:bg-accent-hover',
                'disabled:opacity-50 transition-all',
                'shadow-sm shadow-accent/20',
              )}
            >
              {updateStep.isPending ? 'Сохранение...' : isLastStep ? 'Завершить шаг' : 'Далее'}
              <ArrowRight className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>

      {/* Step counter */}
      <p className="text-center text-xs text-neutral-400 mt-4">
        Шаг {currentStep + 1} из {steps.length}
      </p>
    </div>
  )
}

const INFO_FEATURES = [
  {
    icon: Heart,
    title: 'Лояльность',
    description: 'Бонусные программы, уровни, автоматическое начисление',
  },
  {
    icon: Mail,
    title: 'Рассылки',
    description: 'Таргетированные Telegram-рассылки по сегментам',
  },
  {
    icon: TrendingUp,
    title: 'Аналитика',
    description: 'Продажи, конверсия, RFM-сегментация',
  },
]

const NEXT_STEP_ACTIONS = [
  {
    icon: TrendingUp,
    title: 'Аналитика',
    description: 'Отслеживайте продажи и конверсию',
    href: '/dashboard/analytics/sales',
  },
  {
    icon: Mail,
    title: 'Рассылки',
    description: 'Создайте первую рассылку',
    href: '/dashboard/campaigns',
  },
  {
    icon: Tag,
    title: 'Акции',
    description: 'Запустите промо-акцию',
    href: '/dashboard/promotions',
  },
  {
    icon: Users,
    title: 'Клиенты',
    description: 'Изучите базу клиентов',
    href: '/dashboard/clients',
  },
]

function StepContent({
  step,
  stepIndex,
  isCompleted,
}: {
  step: (typeof ONBOARDING_STEPS)[number]
  stepIndex: number
  isCompleted: boolean
}) {
  const Icon = STEP_ICONS[stepIndex] ?? Circle
  const navigate = useNavigate()

  const stepActions: Record<string, { label: string; href: string }> = {
    info: { label: 'Перейти к информации', href: '/dashboard' },
    bot: { label: 'Создать бота', href: '/dashboard/bots' },
    loyalty: { label: 'Настроить программу', href: '/dashboard/loyalty' },
    pos: { label: 'Подключить POS', href: '/dashboard/pos' },
    integrations: { label: 'Настроить интеграции', href: '/dashboard/integrations' },
    next_steps: { label: 'К дашборду', href: '/dashboard' },
  }

  const action = stepActions[step.key]

  // Rich content for info step
  if (step.key === 'info') {
    return (
      <div className="text-center">
        <div
          className={cn(
            'w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-5',
            isCompleted ? 'bg-green-50' : 'bg-accent/10',
          )}
        >
          {isCompleted ? (
            <CheckCircle2 className="w-8 h-8 text-green-500" />
          ) : (
            <Icon className="w-8 h-8 text-accent" />
          )}
        </div>

        <h2 className="text-xl font-bold text-neutral-900 mb-2">{step.title}</h2>
        <p className="text-sm text-neutral-500 leading-relaxed max-w-md mx-auto mb-6">
          Revisitr — платформа лояльности для HoReCa на базе Telegram.
          Управляйте бонусами, рассылками и аналитикой из одного места.
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
          {INFO_FEATURES.map((feature) => {
            const FeatureIcon = feature.icon
            return (
              <div
                key={feature.title}
                className="bg-neutral-50 border border-neutral-100 rounded-xl p-4 text-left"
              >
                <FeatureIcon className="w-5 h-5 text-accent mb-2" />
                <p className="text-sm font-semibold text-neutral-900 mb-1">{feature.title}</p>
                <p className="text-xs text-neutral-500 leading-relaxed">{feature.description}</p>
              </div>
            )
          })}
        </div>

        {isCompleted && (
          <p className="mt-5 text-sm text-green-600 font-medium">Выполнено</p>
        )}
      </div>
    )
  }

  // Action cards for next_steps step
  if (step.key === 'next_steps') {
    return (
      <div className="text-center">
        <div
          className={cn(
            'w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-5',
            isCompleted ? 'bg-green-50' : 'bg-accent/10',
          )}
        >
          {isCompleted ? (
            <CheckCircle2 className="w-8 h-8 text-green-500" />
          ) : (
            <Icon className="w-8 h-8 text-accent" />
          )}
        </div>

        <h2 className="text-xl font-bold text-neutral-900 mb-2">{step.title}</h2>
        <p className="text-sm text-neutral-500 leading-relaxed max-w-md mx-auto mb-6">
          Настройка завершена. Выберите, с чего начать работу.
        </p>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
          {NEXT_STEP_ACTIONS.map((item) => {
            const ActionIcon = item.icon
            return (
              <Link
                key={item.href}
                to={item.href}
                className={cn(
                  'group flex items-start gap-3 bg-neutral-50 border border-neutral-100',
                  'rounded-xl p-4 text-left transition-all',
                  'hover:border-accent/30 hover:bg-accent/5 hover:shadow-sm',
                )}
              >
                <div className="w-9 h-9 rounded-lg bg-white border border-neutral-100 flex items-center justify-center shrink-0 group-hover:border-accent/20">
                  <ActionIcon className="w-4 h-4 text-neutral-600 group-hover:text-accent transition-colors" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-semibold text-neutral-900 mb-0.5">{item.title}</p>
                  <p className="text-xs text-neutral-500">{item.description}</p>
                </div>
                <ArrowRight className="w-4 h-4 text-neutral-300 group-hover:text-accent mt-0.5 transition-colors shrink-0" />
              </Link>
            )
          })}
        </div>

        {isCompleted && (
          <p className="mt-5 text-sm text-green-600 font-medium">Выполнено</p>
        )}
      </div>
    )
  }

  // Default step content
  return (
    <div className="text-center">
      <div
        className={cn(
          'w-16 h-16 rounded-2xl flex items-center justify-center mx-auto mb-5',
          isCompleted ? 'bg-green-50' : 'bg-accent/10',
        )}
      >
        {isCompleted ? (
          <CheckCircle2 className="w-8 h-8 text-green-500" />
        ) : (
          <Icon className="w-8 h-8 text-accent" />
        )}
      </div>

      <h2 className="text-xl font-bold text-neutral-900 mb-2">{step.title}</h2>
      <p className="text-sm text-neutral-500 leading-relaxed max-w-md mx-auto">
        {step.description}
      </p>

      {action && !isCompleted && (
        <button
          type="button"
          onClick={() => navigate(action.href)}
          className={cn(
            'inline-flex items-center gap-1.5 mt-5 py-2 px-4 rounded-lg text-sm font-medium',
            'border border-surface-border text-neutral-700',
            'hover:bg-neutral-50 transition-colors',
          )}
        >
          {action.label}
          <ArrowRight className="w-3.5 h-3.5" />
        </button>
      )}

      {isCompleted && (
        <p className="mt-4 text-sm text-green-600 font-medium">Выполнено</p>
      )}
    </div>
  )
}
