import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { ChevronLeft, Sparkles, ArrowRight, LayoutGrid, Loader2 } from 'lucide-react'
import { useRFMOnboardingQuestionsQuery, useRFMRecommendMutation, useRFMSetTemplateMutation, useRFMActiveTemplateQuery, useRFMTemplatesQuery } from '@/features/rfm/queries'
import type { OnboardingQuestion, RFMTemplate } from '@/features/rfm/types'
import { ErrorState } from '@/components/common/ErrorState'

export default function RFMOnboardingPage() {
  const navigate = useNavigate()
  const { data: activeTemplate, isLoading: loadingTemplate } = useRFMActiveTemplateQuery()
  const { data: questions, isError, mutate: retryQuestions } = useRFMOnboardingQuestionsQuery()
  const { data: allTemplates } = useRFMTemplatesQuery()
  const recommendMutation = useRFMRecommendMutation()
  const setTemplateMutation = useRFMSetTemplateMutation()

  const [step, setStep] = useState(0) // 0-2 = questions, 3 = result
  const [answers, setAnswers] = useState<(number | null)[]>([null, null, null])
  const [showAllTemplates, setShowAllTemplates] = useState(false)
  const [templateError, setTemplateError] = useState<string | null>(null)

  // Redirect if template already set
  useEffect(() => {
    if (!loadingTemplate && activeTemplate?.template) {
      navigate('/dashboard/rfm', { replace: true })
    }
  }, [activeTemplate, loadingTemplate, navigate])

  function handleAnswer(questionIndex: number, answerId: number) {
    const newAnswers = [...answers]
    newAnswers[questionIndex] = answerId
    setAnswers(newAnswers)

    // Auto-advance after small delay
    setTimeout(() => {
      if (questionIndex < 2) {
        setStep(questionIndex + 1)
      } else {
        // All answered — get recommendation
        recommendMutation.mutate(newAnswers as number[])
        setStep(3)
      }
    }, 300)
  }

  async function handleUseTemplate(template: RFMTemplate) {
    setTemplateError(null)
    try {
      await setTemplateMutation.mutateAsync({ template_type: 'standard', template_key: template.key })
      navigate('/dashboard/rfm')
    } catch {
      setTemplateError('Не удалось сохранить шаблон. Попробуйте ещё раз.')
    }
  }

  if (isError) {
    return (
      <div className="max-w-2xl mx-auto py-12">
        <ErrorState
          title="Не удалось загрузить вопросы"
          onRetry={() => retryQuestions()}
        />
      </div>
    )
  }

  if (!questions) {
    return (
      <div className="max-w-2xl mx-auto py-12">
        <div className="flex items-center justify-center py-24">
          <div className="w-6 h-6 border-2 border-neutral-300 border-t-accent rounded-full animate-spin" />
        </div>
      </div>
    )
  }

  const recommendation = recommendMutation.data

  return (
    <div className="max-w-2xl mx-auto py-4">
      {/* Header */}
      <div className="mb-8 animate-in">
        <h1 className="font-serif text-3xl font-bold text-neutral-900 tracking-tight mb-2">
          Настройка RFM-сегментации
        </h1>
        <p className="text-neutral-500 text-sm">
          Ответьте на 3 вопроса — мы подберём оптимальный шаблон для вашего заведения
        </p>
      </div>

      {/* Progress bar */}
      <div className="mb-8">
        <div className="flex items-center gap-2 mb-2">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className={cn(
                'h-1.5 flex-1 rounded-full transition-all duration-500',
                i < step ? 'bg-accent' : i === step && step < 3 ? 'bg-accent/50' : 'bg-neutral-200',
              )}
            />
          ))}
        </div>
        <p className="text-xs text-neutral-400">
          {step < 3
            ? `Воп��ос ${step + 1} из 3`
            : 'Готово!'}
        </p>
      </div>

      {/* Questions */}
      {step < 3 && (
        <QuestionStep
          question={questions[step]}
          selectedAnswer={answers[step]}
          onAnswer={(id) => handleAnswer(step, id)}
          onBack={step > 0 ? () => setStep(step - 1) : undefined}
        />
      )}

      {/* Recommendation loading */}
      {step === 3 && recommendMutation.isPending && (
        <div className="flex items-center justify-center py-16 animate-in">
          <Loader2 className="w-6 h-6 text-accent animate-spin" />
          <span className="ml-3 text-sm text-neutral-500">Подбираем шаблон...</span>
        </div>
      )}

      {/* Recommendation error */}
      {step === 3 && recommendMutation.isError && (
        <div className="animate-in">
          <ErrorState
            title="Не удалось получить рекомендацию"
            onRetry={() => recommendMutation.mutate(answers as number[])}
          />
        </div>
      )}

      {/* Template save error */}
      {templateError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 p-3 rounded-lg animate-in">
          {templateError}
        </div>
      )}

      {/* Result */}
      {step === 3 && recommendation && !showAllTemplates && (
        <div className="animate-in space-y-6">
          <div className="bg-white rounded-2xl border border-surface-border p-8 text-center">
            <div className="w-14 h-14 rounded-2xl bg-accent/10 flex items-center justify-center mx-auto mb-4">
              <Sparkles className="w-7 h-7 text-accent" />
            </div>
            <h2 className="font-serif text-xl font-bold text-neutral-900 mb-1">
              Рекомендуем: {recommendation.recommended.name}
            </h2>
            <p className="text-sm text-neutral-500 mb-6 max-w-md mx-auto">
              {recommendation.recommended.description}
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <button
                type="button"
                onClick={() => handleUseTemplate(recommendation.recommended)}
                disabled={setTemplateMutation.isPending}
                className={cn(
                  'flex items-center gap-2 py-2.5 px-6 rounded-lg',
                  'bg-accent text-white text-sm font-medium',
                  'hover:bg-accent-hover active:bg-accent/80',
                  'transition-all duration-150',
                  'shadow-sm shadow-accent/20',
                  'disabled:opacity-50',
                )}
              >
                {setTemplateMutation.isPending ? (
                  <Loader2 className="w-4 h-4 animate-spin" />
                ) : (
                  <ArrowRight className="w-4 h-4" />
                )}
                Использовать
              </button>
              <button
                type="button"
                onClick={() => setShowAllTemplates(true)}
                disabled={setTemplateMutation.isPending}
                className={cn(
                  'flex items-center gap-2 py-2.5 px-5 rounded-lg',
                  'border border-neutral-200 text-sm font-medium text-neutral-700',
                  'hover:bg-neutral-50 hover:border-neutral-300',
                  'transition-all duration-150',
                  'disabled:opacity-50',
                )}
              >
                <LayoutGrid className="w-4 h-4" />
                Выбрать другой
              </button>
              <button
                type="button"
                onClick={() => navigate('/dashboard/rfm/template?mode=custom')}
                disabled={setTemplateMutation.isPending}
                className="text-sm text-neutral-400 hover:text-neutral-600 transition-colors py-2"
              >
                Настроить вручную
              </button>
            </div>
          </div>

          {/* Back to questions */}
          <button
            type="button"
            onClick={() => {
              setStep(0)
              setAnswers([null, null, null])
            }}
            className="flex items-center gap-1 text-sm text-neutral-400 hover:text-neutral-600 transition-colors"
          >
            <ChevronLeft className="w-4 h-4" />
            Пройти заново
          </button>
        </div>
      )}

      {/* All templates picker */}
      {step === 3 && showAllTemplates && allTemplates && (
        <div className="animate-in space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-xl font-bold text-neutral-900">Выберите шаблон</h2>
            <button
              type="button"
              onClick={() => setShowAllTemplates(false)}
              className="text-sm text-neutral-400 hover:text-neutral-600 transition-colors"
            >
              ← Назад к рекомендации
            </button>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
            {allTemplates.map((t, i) => (
              <button
                key={t.key}
                type="button"
                onClick={() => handleUseTemplate(t)}
                disabled={setTemplateMutation.isPending}
                className={cn(
                  'text-left bg-white rounded-2xl border border-surface-border p-5',
                  'hover:border-accent/30 hover:shadow-md',
                  'transition-all duration-200',
                  'animate-in',
                  `animate-in-delay-${i + 1}`,
                  'disabled:opacity-50',
                )}
              >
                <h3 className="font-semibold text-neutral-900 mb-1">{t.name}</h3>
                <p className="text-sm text-neutral-500 mb-3">{t.description}</p>
                <div className="flex gap-4 text-xs text-neutral-400">
                  <span>R: {t.r_thresholds.join(', ')} дн</span>
                  <span>F: {t.f_thresholds.join(', ')} виз</span>
                </div>
              </button>
            ))}
          </div>
          <button
            type="button"
            onClick={() => navigate('/dashboard/rfm/template?mode=custom')}
            className="flex items-center gap-2 text-sm text-neutral-500 hover:text-accent transition-colors"
          >
            Или настройте пороги вручную →
          </button>
        </div>
      )}
    </div>
  )
}

function QuestionStep({
  question,
  selectedAnswer,
  onAnswer,
  onBack,
}: {
  question: OnboardingQuestion
  selectedAnswer: number | null
  onAnswer: (id: number) => void
  onBack?: () => void
}) {
  return (
    <div className="animate-in">
      <h2 className="font-serif text-xl font-bold text-neutral-900 mb-5 tracking-tight">
        {question.text}
      </h2>

      <div className="space-y-3">
        {question.answers.map((answer) => (
          <button
            key={answer.id}
            type="button"
            onClick={() => onAnswer(answer.id)}
            className={cn(
              'w-full text-left p-4 rounded-xl border transition-all duration-200',
              selectedAnswer === answer.id
                ? 'border-accent bg-accent/5 shadow-sm'
                : 'border-surface-border bg-white hover:border-neutral-300 hover:shadow-sm',
            )}
          >
            <div className="flex items-center gap-3">
              <div
                className={cn(
                  'w-5 h-5 rounded-full border-2 flex items-center justify-center shrink-0 transition-all',
                  selectedAnswer === answer.id
                    ? 'border-accent'
                    : 'border-neutral-300',
                )}
              >
                {selectedAnswer === answer.id && (
                  <div className="w-2.5 h-2.5 rounded-full bg-accent" />
                )}
              </div>
              <span className="text-sm text-neutral-700">{answer.text}</span>
            </div>
          </button>
        ))}
      </div>

      {onBack && (
        <button
          type="button"
          onClick={onBack}
          className="flex items-center gap-1 mt-6 text-sm text-neutral-400 hover:text-neutral-600 transition-colors"
        >
          <ChevronLeft className="w-4 h-4" />
          Назад
        </button>
      )}
    </div>
  )
}
