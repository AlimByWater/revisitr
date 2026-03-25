export interface OnboardingStep {
  completed: boolean
  skipped: boolean
  entity_id?: number
}

export interface OnboardingState {
  current_step: number
  steps: Record<string, OnboardingStep>
}

export interface OnboardingResponse {
  onboarding_completed: boolean
  onboarding_state: OnboardingState
}

export interface UpdateOnboardingRequest {
  step: string
  completed: boolean
  skipped?: boolean
  entity_id?: number
}

export const ONBOARDING_STEPS = [
  { key: 'info', number: 1, title: 'Информация', description: 'О системе, FAQ, демо' },
  { key: 'loyalty', number: 2, title: 'Программа лояльности', description: 'Создание и настройка программы' },
  { key: 'bot', number: 3, title: 'Создание бота', description: 'Telegram-бот для клиентов' },
  { key: 'pos', number: 4, title: 'Точки продаж', description: 'Добавление POS-локаций' },
  { key: 'integrations', number: 5, title: 'Интеграции', description: 'Подключение POS-систем' },
  { key: 'next_steps', number: 6, title: 'Следующие шаги', description: 'Начать работу' },
] as const
