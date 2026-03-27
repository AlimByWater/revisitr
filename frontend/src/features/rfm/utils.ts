export function formatMoney(value: number | null | undefined): string {
  if (value == null) return '—'
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M ₽`
  if (value >= 1_000) return `${Math.round(value).toLocaleString('ru-RU')} ₽`
  return `${Math.round(value)} ₽`
}

export function formatDate(dateStr: string | null | undefined): string {
  if (!dateStr) return '—'
  return new Date(dateStr).toLocaleDateString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
  })
}

export function pluralClients(n: number): string {
  const mod10 = n % 10
  const mod100 = n % 100
  if (mod10 === 1 && mod100 !== 11) return 'клиент'
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return 'клиента'
  return 'клиентов'
}

export function escapeCsvField(value: unknown): string {
  const str = String(value ?? '')
  if (str.includes(',') || str.includes('"') || str.includes('\n')) {
    return `"${str.replace(/"/g, '""')}"`
  }
  return str
}
