import { cn } from '@/lib/utils'

const BUTTON_STYLES = [
  { value: '' as const, color: 'bg-neutral-400', label: 'Обычная' },
  { value: 'primary' as const, color: 'bg-blue-500', label: 'Синяя' },
  { value: 'success' as const, color: 'bg-green-500', label: 'Зелёная' },
  { value: 'danger' as const, color: 'bg-red-500', label: 'Красная' },
]

type ButtonStyle = '' | 'primary' | 'success' | 'danger'

export function ButtonStylePicker({ value, onChange }: { value: string; onChange: (style: ButtonStyle) => void }) {
  const currentIdx = BUTTON_STYLES.findIndex((s) => s.value === value)
  const current = BUTTON_STYLES[Math.max(currentIdx, 0)]
  const next = () => {
    const nextIdx = (currentIdx + 1) % BUTTON_STYLES.length
    onChange(BUTTON_STYLES[nextIdx].value)
  }

  return (
    <button
      type="button"
      onClick={next}
      title={`Стиль: ${current.label}. Нажмите для смены`}
      className={cn(
        'w-5 h-5 rounded-full shrink-0 transition-all border-2 border-white shadow-sm mt-0.5',
        current.color,
      )}
    />
  )
}
