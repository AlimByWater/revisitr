import { fireEvent, render, screen, within } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { DatePicker } from './DatePicker'

function ControlledDatePicker({ initialValue = '', disabled = false }: { initialValue?: string; disabled?: boolean }) {
  const [value, setValue] = useState(initialValue)

  return (
    <>
      <DatePicker value={value} onChange={setValue} disabled={disabled} />
      <output data-testid="value">{value}</output>
    </>
  )
}

function getPopover(container: HTMLElement) {
  const popover = container.querySelector('div.absolute')
  if (!popover) throw new Error('Popover not found')
  return popover
}

describe('DatePicker', () => {
  it('opens on click and closes on outside mousedown', () => {
    const { container } = render(<ControlledDatePicker />)

    const trigger = screen.getByRole('button', { name: 'Выберите дату' })
    const popover = getPopover(container)

    expect(popover).toHaveClass('pointer-events-none')
    fireEvent.click(trigger)
    expect(popover).toHaveClass('pointer-events-auto')

    fireEvent.mouseDown(document.body)
    expect(popover).toHaveClass('pointer-events-none')
  })

  it('navigates months and selects a day', () => {
    const { container } = render(<ControlledDatePicker />)

    fireEvent.click(screen.getByRole('button', { name: 'Выберите дату' }))

    const currentHeader = screen.getByText(/\d{4}/)
    const nav = currentHeader.parentElement
    if (!nav) throw new Error('Month nav not found')

    const [prevButton, nextButton] = within(nav).getAllByRole('button')
    fireEvent.click(nextButton)
    fireEvent.click(prevButton)
    fireEvent.click(screen.getByRole('button', { name: '20' }))

    expect(screen.getByTestId('value')).toHaveTextContent(/^\d{4}-\d{2}-20$/)
    expect(getPopover(container)).toHaveClass('pointer-events-none')
  })

  it('resets selected date back to empty state', () => {
    render(<ControlledDatePicker initialValue="2026-04-12" />)

    fireEvent.click(screen.getByRole('button', { name: /12.*2026/i }))
    fireEvent.click(screen.getByRole('button', { name: 'Сбросить' }))

    expect(screen.getByTestId('value')).toBeEmptyDOMElement()
    expect(screen.getByRole('button', { name: 'Выберите дату' })).toBeInTheDocument()
  })
})
