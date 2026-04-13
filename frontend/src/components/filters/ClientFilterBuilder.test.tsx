import { fireEvent, render, screen, within } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'

import type { SegmentFilter } from '@/features/segments/types'

import { ClientFilterBuilder } from './ClientFilterBuilder'

function BuilderHarness({
  initialValue = {},
  previewCount,
  onPreview,
  isPreviewing,
  hiddenFields,
}: {
  initialValue?: SegmentFilter
  previewCount?: number | null
  onPreview?: () => void
  isPreviewing?: boolean
  hiddenFields?: string[]
}) {
  const [value, setValue] = useState<SegmentFilter>(initialValue)

  return (
    <>
      <ClientFilterBuilder
        value={value}
        onChange={setValue}
        previewCount={previewCount}
        onPreview={onPreview}
        isPreviewing={isPreviewing}
        hiddenFields={hiddenFields}
      />
      <pre data-testid="value">{JSON.stringify(value)}</pre>
    </>
  )
}

function readFilter() {
  return JSON.parse(screen.getByTestId('value').textContent ?? '{}') as SegmentFilter
}

describe('ClientFilterBuilder', () => {
  it('shows preview count and calls preview action', () => {
    const onPreview = vi.fn()

    render(
      <BuilderHarness
        initialValue={{ city: 'Москва', registered_from: '2026-04-03' }}
        previewCount={12}
        onPreview={onPreview}
      />,
    )

    const demographyButton = screen.getByRole('button', { name: /Демография/ })
    const activityButton = screen.getByRole('button', { name: /Активность/ })

    expect(within(demographyButton).getByText('1')).toBeInTheDocument()
    expect(within(activityButton).getByText('1')).toBeInTheDocument()
    expect(screen.getByText((_, element) => element?.textContent === 'Найдено: 12 клиентов')).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Посчитать' }))
    expect(onPreview).toHaveBeenCalledTimes(1)
  })

  it('keeps hidden fields on reset and excludes hidden controls', () => {
    render(
      <BuilderHarness
        initialValue={{ search: 'Анна', bot_id: 7, city: 'Москва', min_visits: 3 }}
        hiddenFields={['search', 'bot_id']}
      />,
    )

    expect(screen.queryByText('Поиск по имени / телефону')).not.toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Сбросить фильтры' }))
    expect(readFilter()).toEqual({ search: 'Анна', bot_id: 7 })
  })

  it('updates date and numeric fields and removes key after clear', () => {
    render(<BuilderHarness />)

    fireEvent.click(screen.getByRole('button', { name: /Активность/ }))

    const registrationBlock = screen.getByText('Дата регистрации').parentElement
    if (!registrationBlock) throw new Error('Registration block not found')

    const fromDateButton = within(registrationBlock).getByRole('button', { name: 'от' })
    fireEvent.click(fromDateButton)

    const fromDatePicker = fromDateButton.parentElement
    if (!fromDatePicker) throw new Error('DatePicker root not found')
    fireEvent.click(within(fromDatePicker).getByRole('button', { name: '10' }))

    const visitsBlock = screen.getByText('Количество визитов').parentElement
    if (!visitsBlock) throw new Error('Visits block not found')

    const [minVisits, maxVisits] = within(visitsBlock).getAllByRole('spinbutton')
    fireEvent.change(minVisits, { target: { value: '3' } })
    fireEvent.change(maxVisits, { target: { value: '9' } })

    expect(readFilter()).toMatchObject({ min_visits: 3, max_visits: 9 })
    expect(readFilter().registered_from).toMatch(/^\d{4}-\d{2}-10$/)

    fireEvent.change(minVisits, { target: { value: '' } })

    expect(readFilter()).toMatchObject({ max_visits: 9 })
    expect(readFilter()).not.toHaveProperty('min_visits')
  })
})
