import { fireEvent, render, screen } from '@testing-library/react'
import { useState } from 'react'
import { describe, expect, it } from 'vitest'

import { CustomSelect, type SelectGroup, type SelectOption } from './CustomSelect'

function ControlledSelect({
  initialValue = '',
  options,
  groups,
  disabled = false,
}: {
  initialValue?: string
  options: SelectOption[]
  groups?: SelectGroup[]
  disabled?: boolean
}) {
  const [value, setValue] = useState(initialValue)

  return (
    <>
      <CustomSelect value={value} onChange={setValue} options={options} groups={groups} disabled={disabled} />
      <output data-testid="value">{value}</output>
    </>
  )
}

function getDropdown(container: HTMLElement) {
  const dropdown = container.querySelector('div.absolute')
  if (!dropdown) throw new Error('Dropdown not found')
  return dropdown
}

describe('CustomSelect', () => {
  const options = [
    { value: 'basic', label: 'Базовый' },
    { value: 'vip', label: 'VIP' },
    { value: 'premium', label: 'Премиум' },
  ]

  it('opens on click and closes on outside mousedown', () => {
    const { container } = render(<ControlledSelect options={options} />)

    const trigger = screen.getByRole('button', { name: 'Выберите...' })
    const dropdown = getDropdown(container)

    expect(dropdown).toHaveClass('pointer-events-none')

    fireEvent.click(trigger)
    expect(dropdown).toHaveClass('pointer-events-auto')

    fireEvent.mouseDown(document.body)
    expect(dropdown).toHaveClass('pointer-events-none')
  })

  it('selects option, updates value, and closes dropdown', () => {
    const { container } = render(<ControlledSelect options={options} />)

    fireEvent.click(screen.getByRole('button', { name: 'Выберите...' }))
    fireEvent.click(screen.getAllByRole('button', { name: 'VIP' })[0])

    expect(screen.getByTestId('value')).toHaveTextContent('vip')
    expect(getDropdown(container)).toHaveClass('pointer-events-none')
    expect(screen.getAllByRole('button', { name: 'VIP' })[0]).toHaveTextContent('VIP')
  })

  it('stays closed when disabled', () => {
    const { container } = render(<ControlledSelect options={options} disabled />)

    const trigger = screen.getByRole('button', { name: 'Выберите...' })
    expect(trigger).toBeDisabled()

    fireEvent.click(trigger)
    expect(getDropdown(container)).toHaveClass('pointer-events-none')
  })

  it('renders grouped options and resolves selected label from groups', () => {
    const groups: SelectGroup[] = [
      { options: [{ value: 'basic', label: 'Базовый' }] },
      { options: [{ value: 'premium', label: 'Премиум' }, { value: 'vip', label: 'VIP' }] },
    ]

    render(<ControlledSelect initialValue="vip" options={[]} groups={groups} />)

    expect(screen.getAllByRole('button', { name: 'VIP' })[0]).toBeInTheDocument()

    fireEvent.click(screen.getAllByRole('button', { name: 'VIP' })[0])
    expect(screen.getByRole('button', { name: 'Базовый' })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Премиум' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Премиум' }))
    expect(screen.getByTestId('value')).toHaveTextContent('premium')
  })
})
