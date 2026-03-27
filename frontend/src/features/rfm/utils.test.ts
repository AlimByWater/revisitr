import { describe, it, expect } from 'vitest'
import { formatMoney, formatDate, pluralClients, escapeCsvField } from './utils'

describe('formatMoney', () => {
  it('returns dash for null', () => {
    expect(formatMoney(null)).toBe('—')
  })

  it('returns dash for undefined', () => {
    expect(formatMoney(undefined)).toBe('—')
  })

  it('formats small amounts without locale separator', () => {
    expect(formatMoney(500)).toBe('500 ₽')
  })

  it('formats thousands with locale separator', () => {
    const result = formatMoney(12500)
    expect(result).toContain('₽')
    expect(result).toContain('12')
    expect(result).toContain('500')
  })

  it('formats millions with M suffix', () => {
    expect(formatMoney(2_500_000)).toBe('2.5M ₽')
  })

  it('formats exactly 1M', () => {
    expect(formatMoney(1_000_000)).toBe('1.0M ₽')
  })

  it('rounds to nearest integer for values under 1M', () => {
    expect(formatMoney(99)).toBe('99 ₽')
  })

  it('handles zero', () => {
    expect(formatMoney(0)).toBe('0 ₽')
  })
})

describe('formatDate', () => {
  it('returns dash for null', () => {
    expect(formatDate(null)).toBe('—')
  })

  it('returns dash for undefined', () => {
    expect(formatDate(undefined)).toBe('—')
  })

  it('returns dash for empty string', () => {
    expect(formatDate('')).toBe('—')
  })

  it('formats ISO date to ru-RU locale', () => {
    const result = formatDate('2026-03-15T12:00:00Z')
    expect(result).toMatch(/15/)
    expect(result).toMatch(/03/)
    expect(result).toMatch(/2026/)
  })
})

describe('pluralClients', () => {
  it('1 → клиент', () => {
    expect(pluralClients(1)).toBe('клиент')
  })

  it('2 → клиента', () => {
    expect(pluralClients(2)).toBe('клиента')
  })

  it('3 → клиента', () => {
    expect(pluralClients(3)).toBe('клиента')
  })

  it('4 → клиента', () => {
    expect(pluralClients(4)).toBe('клиента')
  })

  it('5 → клиентов', () => {
    expect(pluralClients(5)).toBe('клиентов')
  })

  it('11 → клиентов (special case)', () => {
    expect(pluralClients(11)).toBe('клиентов')
  })

  it('12 → клиентов', () => {
    expect(pluralClients(12)).toBe('клиентов')
  })

  it('21 → клиент', () => {
    expect(pluralClients(21)).toBe('клиент')
  })

  it('22 → клиента', () => {
    expect(pluralClients(22)).toBe('клиента')
  })

  it('0 → клиентов', () => {
    expect(pluralClients(0)).toBe('клиентов')
  })

  it('111 → клиентов', () => {
    expect(pluralClients(111)).toBe('клиентов')
  })

  it('101 → клиент', () => {
    expect(pluralClients(101)).toBe('клиент')
  })
})

describe('escapeCsvField', () => {
  it('returns plain string as-is', () => {
    expect(escapeCsvField('hello')).toBe('hello')
  })

  it('wraps comma-containing fields in quotes', () => {
    expect(escapeCsvField('a,b')).toBe('"a,b"')
  })

  it('escapes internal double quotes', () => {
    expect(escapeCsvField('say "hi"')).toBe('"say ""hi"""')
  })

  it('wraps newline-containing fields', () => {
    expect(escapeCsvField('line1\nline2')).toBe('"line1\nline2"')
  })

  it('handles null/undefined by converting to empty string', () => {
    expect(escapeCsvField(null)).toBe('')
    expect(escapeCsvField(undefined)).toBe('')
  })

  it('converts numbers to string', () => {
    expect(escapeCsvField(42)).toBe('42')
  })
})
