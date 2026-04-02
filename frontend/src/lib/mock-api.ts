/**
 * Mock API interceptor for local frontend development without backend.
 * Enabled via VITE_MOCK_API=true in .env or environment.
 *
 * Intercepts axios requests and returns realistic mock data.
 */
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

const now = new Date().toISOString()
const ago = (days: number) => new Date(Date.now() - days * 86400000).toISOString()
const dayStr = (d: Date) => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`

let idCounter = 100

function nextId() { return ++idCounter }

function randomBetween(min: number, max: number) {
  return Math.round(min + Math.random() * (max - min))
}

// ── Seeded random for deterministic data generation ─────────────────────────
let _seed = 42
function seeded() { _seed = (_seed * 16807 + 0) % 2147483647; return (_seed - 1) / 2147483646 }
function seededBetween(min: number, max: number) { return Math.round(min + seeded() * (max - min)) }
function seededPick<T>(arr: T[]): T { return arr[Math.floor(seeded() * arr.length)] }

function chartPoints(days: number, key = 'date'): Record<string, string | number>[] {
  return Array.from({ length: days }, (_, i) => ({
    [key]: ago(days - 1 - i).slice(0, 10),
    value: randomBetween(5000, 50000),
  }))
}

// ── Date range helpers ──────────────────────────────────────────────────────
function getDateRange(query?: { period?: string; from?: string; to?: string }): [Date, Date] {
  if (query?.from && query?.to) {
    // Parse as local dates (YYYY-MM-DD → local midnight)
    const [fy, fm, fd] = query.from.split('-').map(Number)
    const [ty, tm, td] = query.to.split('-').map(Number)
    const start = new Date(fy, fm - 1, fd, 0, 0, 0, 0)
    const end = new Date(ty, tm - 1, td, 23, 59, 59, 999)
    return [start, end]
  }
  const days: Record<string, number> = { today: 0, yesterday: 1, '7d': 6, '14d': 13, '30d': 29, '90d': 89, '180d': 179, '365d': 364 }
  const d = days[query?.period ?? '30d'] ?? 29
  const today = new Date(); today.setHours(23, 59, 59, 999)
  if (query?.period === 'yesterday') {
    const yEnd = new Date(); yEnd.setDate(yEnd.getDate() - 1); yEnd.setHours(23, 59, 59, 999)
    const yStart = new Date(yEnd); yStart.setHours(0, 0, 0, 0)
    return [yStart, yEnd]
  }
  const start = new Date(); start.setDate(start.getDate() - d); start.setHours(0, 0, 0, 0)
  return [start, today]
}

function groupByDay<T>(items: T[], dateKey: string): Record<string, T[]> {
  const groups: Record<string, T[]> = {}
  for (const item of items) {
    const raw = (item as any)[dateKey]
    if (!raw) continue
    const d = dayStr(new Date(raw))
    if (!groups[d]) groups[d] = []
    groups[d].push(item)
  }
  return groups
}

function fillDayChart(start: Date, end: Date, grouped: Record<string, any[]>, valueFn: (items: any[]) => number, key = 'date'): Record<string, string | number>[] {
  const result: Record<string, string | number>[] = []
  const d = new Date(start)
  while (d <= end) {
    const ds = dayStr(d)
    result.push({ [key]: ds, value: valueFn(grouped[ds] ?? []) })
    d.setDate(d.getDate() + 1)
  }
  return result
}

// ── Transaction generator (runs once at module load) ────────────────────────
interface MockTransaction {
  id: number
  client_id: number
  bot_id: number
  pos_id: number
  amount: number
  items_count: number
  date: string // ISO string
  has_loyalty: boolean
}

function generateTransactions(): MockTransaction[] {
  _seed = 42 // reset seed for deterministic output
  const txs: MockTransaction[] = []
  let txId = 1

  // 180 days of history, ~25 transactions/day avg for a small HoReCa chain
  for (let daysAgo = 179; daysAgo >= 0; daysAgo--) {
    const d = new Date()
    d.setDate(d.getDate() - daysAgo)
    // Weekends: +30% traffic
    const isWeekend = d.getDay() === 0 || d.getDay() === 6
    const baseTx = isWeekend ? 32 : 22
    const txCount = seededBetween(baseTx - 6, baseTx + 8)

    for (let t = 0; t < txCount; t++) {
      const clientId = seededBetween(1, 360)
      const botId = clientId % 2 === 0 ? 1 : 2
      // Amount: lognormal-ish — mostly 400-2000, sometimes up to 5000
      const base = 300 + seeded() * 700 // 300-1000
      const mult = 1 + seeded() * seeded() * 4 // 1-5, skewed low
      const amount = Math.round(base * mult / 10) * 10 // round to 10₽
      const hour = seededBetween(8, 22)
      const minute = seededBetween(0, 59)
      const txDate = new Date(d)
      txDate.setHours(hour, minute, seededBetween(0, 59))

      txs.push({
        id: txId++,
        client_id: clientId,
        bot_id: botId,
        pos_id: botId, // pos matches bot for simplicity
        amount,
        items_count: seededBetween(1, 5),
        date: txDate.toISOString(),
        has_loyalty: seeded() > 0.3, // 70% have loyalty
      })
    }
  }
  return txs
}

// ── Analytics computation from transactions ─────────────────────────────────
function filterTransactions(txs: MockTransaction[], query?: any): MockTransaction[] {
  const [start, end] = getDateRange(query)
  let filtered = txs.filter(tx => {
    const d = new Date(tx.date)
    return d >= start && d <= end
  })
  if (query?.bot_id) filtered = filtered.filter(tx => tx.bot_id === +query.bot_id)
  return filtered
}

function computeSalesAnalytics(txs: MockTransaction[], query?: any) {
  const filtered = filterTransactions(txs, query)
  const [start, end] = getDateRange(query)
  const totalAmount = filtered.reduce((s, tx) => s + tx.amount, 0)
  const uniqueClients = new Set(filtered.map(tx => tx.client_id)).size
  const avgAmount = filtered.length > 0 ? Math.round(totalAmount / filtered.length) : 0
  const clientTxCounts = new Map<number, number>()
  for (const tx of filtered) clientTxCounts.set(tx.client_id, (clientTxCounts.get(tx.client_id) ?? 0) + 1)
  const buyFreq = uniqueClients > 0 ? +(filtered.length / uniqueClients).toFixed(1) : 0

  const grouped = groupByDay(filtered, 'date')
  const loyaltyTx = filtered.filter(tx => tx.has_loyalty)
  const nonLoyaltyTx = filtered.filter(tx => !tx.has_loyalty)
  const loyaltyAvg = loyaltyTx.length > 0 ? Math.round(loyaltyTx.reduce((s, tx) => s + tx.amount, 0) / loyaltyTx.length) : 0
  const nonLoyaltyAvg = nonLoyaltyTx.length > 0 ? Math.round(nonLoyaltyTx.reduce((s, tx) => s + tx.amount, 0) / nonLoyaltyTx.length) : 0

  return {
    metrics: {
      transaction_count: filtered.length,
      unique_clients: uniqueClients,
      total_amount: totalAmount,
      avg_amount: avgAmount,
      buy_frequency: buyFreq,
    },
    charts: {
      revenue: fillDayChart(start, end, grouped, items => items.reduce((s, tx) => s + tx.amount, 0), 'day'),
      transactions: fillDayChart(start, end, grouped, items => items.length, 'day'),
    },
    comparison: {
      participants_avg_amount: loyaltyAvg,
      non_participants_avg_amount: nonLoyaltyAvg,
    },
  }
}

function computeDashboardWidgets(txs: MockTransaction[], clients: any[], query?: any) {
  const [start, end] = getDateRange(query)
  const filtered = filterTransactions(txs, query)
  const totalAmount = filtered.reduce((s, tx) => s + tx.amount, 0)
  const avgCheck = filtered.length > 0 ? Math.round(totalAmount / filtered.length) : 0
  const newClients = clients.filter(c => { const d = new Date(c.registered_at); return d >= start && d <= end }).length
  const activeClientIds = new Set(filtered.map(tx => tx.client_id)).size

  // Previous period for trend calculation
  const periodMs = end.getTime() - start.getTime()
  const prevStart = new Date(start.getTime() - periodMs)
  const prevEnd = new Date(start.getTime() - 1)
  const prevFiltered = txs.filter(tx => { const d = new Date(tx.date); return d >= prevStart && d <= prevEnd })
  const prevAmount = prevFiltered.reduce((s, tx) => s + tx.amount, 0)
  const prevAvg = prevFiltered.length > 0 ? Math.round(prevAmount / prevFiltered.length) : 0
  const prevNew = clients.filter(c => { const d = new Date(c.registered_at); return d >= prevStart && d <= prevEnd }).length
  const prevActive = new Set(prevFiltered.map(tx => tx.client_id)).size

  const trend = (cur: number, prev: number) => prev > 0 ? +((cur - prev) / prev * 100).toFixed(1) : 0

  return {
    revenue: { value: totalAmount, previous: prevAmount, trend: trend(totalAmount, prevAmount) },
    avg_check: { value: avgCheck, previous: prevAvg, trend: trend(avgCheck, prevAvg) },
    new_clients: { value: newClients, previous: prevNew, trend: trend(newClients, prevNew) },
    active_clients: { value: activeClientIds, previous: prevActive, trend: trend(activeClientIds, prevActive) },
  }
}

function computeDashboardCharts(txs: MockTransaction[], clients: any[], query?: any) {
  const [start, end] = getDateRange(query)
  const filtered = filterTransactions(txs, query)
  const grouped = groupByDay(filtered, 'date')

  const clientsByDay = groupByDay(clients.map(c => ({ ...c, _day: c.registered_at })), '_day')

  return {
    revenue: fillDayChart(start, end, grouped, items => items.reduce((s, tx) => s + tx.amount, 0)),
    new_clients: fillDayChart(start, end, clientsByDay, items => items.length),
  }
}

function computeLoyaltyAnalytics(txs: MockTransaction[], clients: any[], query?: any) {
  const [start, end] = getDateRange(query)
  const filtered = filterTransactions(txs, query)
  if (query?.bot_id) clients = clients.filter(c => c.bot_id === +query.bot_id)

  const newClients = clients.filter(c => { const d = new Date(c.registered_at); return d >= start && d <= end }).length
  const activeClientIds = new Set(filtered.map(tx => tx.client_id))
  const loyaltyTx = filtered.filter(tx => tx.has_loyalty)
  const bonusEarned = Math.round(loyaltyTx.reduce((s, tx) => s + tx.amount * 0.07, 0))
  const bonusSpent = Math.round(bonusEarned * 0.6)

  const males = clients.filter(c => c.gender === 'male').length
  const females = clients.length - males
  const mPct = clients.length > 0 ? Math.round(males / clients.length * 100) : 50

  return {
    new_clients: newClients,
    active_clients: activeClientIds.size,
    bonus_earned: bonusEarned,
    bonus_spent: bonusSpent,
    demographics: {
      by_gender: [{ label: 'Мужской', value: males, percent: mPct }, { label: 'Женский', value: females, percent: 100 - mPct }],
      by_age_group: [{ label: '18-25', value: 20, percent: 20 }, { label: '26-35', value: 40, percent: 40 }, { label: '36-45', value: 25, percent: 25 }, { label: '46+', value: 15, percent: 15 }],
      by_os: [{ label: 'iOS', value: 42, percent: 42 }, { label: 'Android', value: 58, percent: 58 }],
      loyalty_percent: clients.length > 0 ? Math.round(loyaltyTx.length / Math.max(filtered.length, 1) * 100) : 0,
    },
    bot_funnel: (() => {
      const total = clients.length
      const registered = Math.round(total * 0.69)
      const firstBuy = activeClientIds.size
      const repeat = clients.filter(c => {
        const count = filtered.filter(tx => tx.client_id === c.id).length
        return count > 1
      }).length
      return [
        { step: 'Открыли бота', count: total, percent: 100 },
        { step: 'Зарегистрировались', count: registered, percent: total > 0 ? +((registered / total) * 100).toFixed(1) : 0 },
        { step: 'Первая покупка', count: firstBuy, percent: total > 0 ? +((firstBuy / total) * 100).toFixed(1) : 0 },
        { step: 'Повторная покупка', count: repeat, percent: total > 0 ? +((repeat / total) * 100).toFixed(1) : 0 },
      ]
    })(),
  }
}

function computeCampaignAnalytics(campaigns: any[], query?: any) {
  const [start, end] = getDateRange(query)
  const filtered = campaigns.filter(c => {
    if (!c.sent_at) return false
    const d = new Date(c.sent_at)
    return d >= start && d <= end
  })
  const totalSent = filtered.reduce((s, c) => s + (c.stats?.sent ?? 0), 0)
  const opened = Math.round(totalSent * 0.7)
  const conversions = Math.round(opened * 0.2)
  return {
    total_sent: totalSent,
    total_opened: opened,
    open_rate: totalSent > 0 ? +((opened / totalSent) * 100).toFixed(1) : 0,
    conversions,
    conv_rate: totalSent > 0 ? +((conversions / totalSent) * 100).toFixed(1) : 0,
    by_campaign: filtered.map(c => ({
      campaign_id: c.id, campaign_name: c.name,
      sent: c.stats?.sent ?? 0,
      open_rate: 65 + (c.id * 7) % 20,
      conversions: 10 + (c.id * 13) % 40,
    })),
  }
}

// ---------------------------------------------------------------------------
// Mock Data Store (mutable — supports CRUD)
// ---------------------------------------------------------------------------

const store = {
  bots: [
    {
      id: 1, org_id: 1, name: 'Кофейня на Маросейке', username: 'marosejka_bot',
      token_masked: '7281***:AAH***', status: 'active' as const,
      settings: { modules: ['loyalty', 'campaigns'], buttons: [{ label: 'Меню', type: 'webapp', value: 'https://example.com' }], registration_form: [{ name: 'phone', label: 'Телефон', type: 'phone', required: true }], welcome_message: 'Привет! Добро пожаловать в нашу кофейню.' },
      created_at: ago(90), updated_at: ago(2), client_count: 210, program_id: 1,
    },
    {
      id: 2, org_id: 1, name: 'Бар «Хмель»', username: 'hmel_bar_bot',
      token_masked: '6543***:BBX***', status: 'active' as const,
      settings: { modules: ['loyalty'], buttons: [], registration_form: [], welcome_message: 'Добро пожаловать!' },
      created_at: ago(45), updated_at: ago(5), client_count: 150, program_id: 2,
    },
  ] as any[],

  pos: [
    { id: 1, org_id: 1, name: 'Маросейка, 12', address: 'ул. Маросейка, д. 12, Москва', phone: '+7 495 123-45-67', schedule: { mon: { open: '08:00', close: '22:00' }, tue: { open: '08:00', close: '22:00' }, wed: { open: '08:00', close: '22:00' }, thu: { open: '08:00', close: '22:00' }, fri: { open: '08:00', close: '23:00' }, sat: { open: '10:00', close: '23:00' }, sun: { open: '10:00', close: '21:00' } }, is_active: true, created_at: ago(90), updated_at: ago(10), bot_id: 1 },
    { id: 2, org_id: 1, name: 'Покровка, 3', address: 'ул. Покровка, д. 3, Москва', phone: '+7 495 765-43-21', schedule: { mon: { open: '12:00', close: '02:00' }, tue: { open: '12:00', close: '02:00' }, wed: { open: '12:00', close: '02:00' }, thu: { open: '12:00', close: '02:00' }, fri: { open: '12:00', close: '04:00' }, sat: { open: '12:00', close: '04:00' }, sun: { open: '14:00', close: '00:00', closed: false } }, is_active: true, created_at: ago(45), updated_at: ago(3), bot_id: 2 },
  ] as any[],

  loyaltyPrograms: [
    { id: 1, org_id: 1, name: 'Бонусная программа', type: 'bonus', config: { welcome_bonus: 100, currency_name: 'баллов' }, is_active: true, created_at: ago(90), updated_at: ago(15), levels: [
      { id: 1, program_id: 1, name: 'Бронза', threshold: 0, reward_percent: 5, reward_type: 'percent', reward_amount: 5, sort_order: 1 },
      { id: 2, program_id: 1, name: 'Серебро', threshold: 5000, reward_percent: 7, reward_type: 'percent', reward_amount: 7, sort_order: 2 },
      { id: 3, program_id: 1, name: 'Золото', threshold: 15000, reward_percent: 10, reward_type: 'percent', reward_amount: 10, sort_order: 3 },
    ] },
    { id: 2, org_id: 1, name: 'Скидочная карта', type: 'discount', config: { welcome_bonus: 0, currency_name: '' }, is_active: true, created_at: ago(45), updated_at: ago(20), levels: [
      { id: 4, program_id: 2, name: 'Стандарт', threshold: 0, reward_percent: 5, reward_type: 'percent', reward_amount: 5, sort_order: 1 },
      { id: 5, program_id: 2, name: 'VIP', threshold: 10000, reward_percent: 10, reward_type: 'percent', reward_amount: 10, sort_order: 2 },
    ] },
  ] as any[],

  campaigns: [
    { id: 1, org_id: 1, bot_id: 1, name: 'Весенняя акция', type: 'manual', status: 'sent', audience_filter: { bot_id: 1, tags: [] }, message: 'Весенние скидки до 30%! Ждём вас.', media_url: '', buttons: [{ text: 'Подробнее', url: 'https://example.com' }], tracking_mode: 'clicks', scheduled_at: ago(5), sent_at: ago(5), stats: { total: 300, sent: 295, failed: 5 }, created_at: ago(10), updated_at: ago(5) },
    { id: 2, org_id: 1, bot_id: 1, name: 'День рождения', type: 'auto', status: 'completed', audience_filter: { bot_id: 1 }, message: 'С днём рождения! Дарим 500 бонусов.', media_url: '', buttons: [], tracking_mode: 'none', scheduled_at: null, sent_at: null, stats: { total: 42, sent: 42, failed: 0 }, created_at: ago(60), updated_at: ago(1) },
  ] as any[],

  promotions: [
    { id: 1, org_id: 1, name: 'Счастливые часы', type: 'discount', conditions: { min_amount: 500, segment_id: null, min_visits: 0 }, result: { discount_percent: 20, bonus_amount: 0, tag_add: '', campaign_id: null }, recurrence: 'daily', starts_at: ago(30), ends_at: ago(-30), usage_limit: 0, combinable: false, active: true, created_at: ago(30) },
    { id: 2, org_id: 1, name: 'Приведи друга', type: 'bonus', conditions: { min_amount: 0, segment_id: null, min_visits: 1 }, result: { discount_percent: 0, bonus_amount: 300, tag_add: 'referral', campaign_id: null }, recurrence: 'one_time', starts_at: ago(60), ends_at: ago(-90), usage_limit: 1, combinable: true, active: true, created_at: ago(60) },
  ] as any[],

  clients: (() => {
    // Generate realistic registration dates: more recent = more clients (growth curve)
    // Use seeded random so dates are deterministic
    _seed = 777
    const regDays: number[] = []
    for (let i = 0; i < 360; i++) {
      // Exponential bias toward recent: most clients registered in last 60 days
      const raw = seeded()
      const daysAgo = Math.round(raw * raw * 179) // squared → skewed toward 0 (recent)
      regDays.push(daysAgo)
    }
    regDays.sort((a, b) => b - a) // oldest first

    return regDays.map((daysAgo, i) => ({
      id: i + 1, bot_id: (i % 2) + 1, telegram_id: 100000 + i,
      username: `user_${i + 1}`, first_name: ['Алексей', 'Мария', 'Дмитрий', 'Анна', 'Сергей', 'Елена', 'Иван', 'Ольга', 'Павел', 'Наталья'][i % 10],
      last_name: ['Иванов', 'Петрова', 'Сидоров', 'Козлова', 'Смирнов', 'Новикова', 'Морозов', 'Волкова', 'Лебедев', 'Соколова'][i % 10],
      phone: `+7 9${10 + (i % 90)} ${100 + (i % 900)}-${10 + (i % 90)}-${10 + (i % 90)}`,
      gender: i % 2 === 0 ? 'male' : 'female', birth_date: `199${i % 10}-0${(i % 9) + 1}-${10 + (i % 20)}`,
      city: ['Москва', 'Санкт-Петербург', 'Казань'][i % 3], os: i % 3 === 0 ? 'iOS' : 'Android',
      tags: i % 4 === 0 ? ['vip'] : [], registered_at: ago(daysAgo),
      bot_name: i % 2 === 0 ? 'Кофейня на Маросейке' : 'Бар «Хмель»',
      loyalty_balance: 100 + (i * 13) % 5000, loyalty_level: ['Бронза', 'Серебро', 'Золото'][i % 3],
      total_purchases: 500 + (i * 277) % 100000, purchase_count: 1 + (i % 50),
      rfm_segment: ['vip', 'regular', 'promising', 'churn_risk', 'new', 'rare_valuable', 'lost'][i % 7],
    }))
  })(),

  integrations: [
    { id: 1, org_id: 1, type: 'iiko', config: { api_url: 'https://api.iiko.services', api_key: '***', org_id: 'org-1' }, status: 'active', last_sync_at: ago(0.1), created_at: ago(60), updated_at: ago(0.1) },
  ] as any[],

  menus: [
    { id: 1, org_id: 1, integration_id: 1, name: 'Основное меню', source: 'manual', last_synced_at: null, created_at: ago(30), updated_at: ago(5), categories: [
      { id: 1, menu_id: 1, name: 'Кофе', sort_order: 1, items: [
        { id: 1, category_id: 1, name: 'Капучино', description: 'Классический капучино', price: 350, image_url: '', tags: ['popular'], is_available: true, sort_order: 1 },
        { id: 2, category_id: 1, name: 'Латте', description: 'Нежный латте', price: 380, image_url: '', tags: [], is_available: true, sort_order: 2 },
        { id: 3, category_id: 1, name: 'Эспрессо', description: 'Двойной эспрессо', price: 250, image_url: '', tags: [], is_available: true, sort_order: 3 },
      ] },
      { id: 2, menu_id: 1, name: 'Выпечка', sort_order: 2, items: [
        { id: 4, category_id: 2, name: 'Круассан', description: 'Свежий круассан с маслом', price: 220, image_url: '', tags: [], is_available: true, sort_order: 1 },
        { id: 5, category_id: 2, name: 'Чизкейк', description: 'Нью-Йорк чизкейк', price: 420, image_url: '', tags: ['popular'], is_available: true, sort_order: 2 },
      ] },
    ] },
  ] as any[],

  transactions: generateTransactions(),

  billing: {
    tariffs: [
      { id: 1, name: 'Старт', slug: 'start', price: 0, currency: 'RUB', interval: 'month', features: { loyalty: true, campaigns: false, promotions: false, integrations: false, analytics: false, rfm: false, advanced_campaigns: false }, limits: { max_clients: 100, max_bots: 1, max_campaigns_per_month: 0, max_pos: 1 }, active: true, sort_order: 1, created_at: ago(365) },
      { id: 2, name: 'Бизнес', slug: 'business', price: 299000, currency: 'RUB', interval: 'month', features: { loyalty: true, campaigns: true, promotions: true, integrations: true, analytics: true, rfm: false, advanced_campaigns: false }, limits: { max_clients: 1000, max_bots: 3, max_campaigns_per_month: 10, max_pos: 5 }, active: true, sort_order: 2, created_at: ago(365) },
      { id: 3, name: 'Про', slug: 'pro', price: 599000, currency: 'RUB', interval: 'month', features: { loyalty: true, campaigns: true, promotions: true, integrations: true, analytics: true, rfm: true, advanced_campaigns: true }, limits: { max_clients: 10000, max_bots: 10, max_campaigns_per_month: 100, max_pos: 20 }, active: true, sort_order: 3, created_at: ago(365) },
    ],
    subscription: { id: 1, org_id: 1, tariff_id: 3, status: 'active', current_period_start: ago(15), current_period_end: ago(-15), canceled_at: null, created_at: ago(90), updated_at: ago(15), tariff_name: 'Про', tariff_slug: 'pro', tariff_price: 599000, tariff_features: { loyalty: true, campaigns: true, promotions: true, integrations: true, analytics: true, rfm: true, advanced_campaigns: true }, tariff_limits: { max_clients: 10000, max_bots: 10, max_campaigns_per_month: 100, max_pos: 20 } },
    invoices: [
      { id: 1, org_id: 1, subscription_id: 1, amount: 599000, currency: 'RUB', status: 'paid', due_date: ago(15), paid_at: ago(15), created_at: ago(15) },
      { id: 2, org_id: 1, subscription_id: 1, amount: 599000, currency: 'RUB', status: 'pending', due_date: ago(-15), paid_at: null, created_at: ago(0) },
    ],
  },

  onboarding: {
    onboarding_completed: true,
    onboarding_state: { current_step: 'next_steps', steps: { info: { completed: true, skipped: false }, loyalty: { completed: true, skipped: false, entity_id: '1' }, bot: { completed: true, skipped: false, entity_id: '1' }, pos: { completed: true, skipped: false, entity_id: '1' }, integrations: { completed: false, skipped: true }, next_steps: { completed: true, skipped: false } } },
  },

  profile: { id: 1, email: 'demo@revisitr.ru', name: 'Марк Демо', role: 'admin', org_id: 1, phone: '+7 999 123-45-67' },

  billingDetails: { entity_type: 'none' as string, fields: {} },

  segments: [
    { id: 1, org_id: 1, name: 'VIP клиенты', type: 'custom', filter: { min_spend: 10000, tags: ['vip'] }, auto_assign: true, client_count: 45, created_at: ago(60), updated_at: ago(5) },
    { id: 2, org_id: 1, name: 'Риск оттока', type: 'rfm', filter: { rfm_category: 'churn_risk' }, auto_assign: true, client_count: 23, created_at: ago(30), updated_at: ago(2) },
  ] as any[],

  wallet: {
    configs: [{ id: 1, org_id: 1, platform: 'apple', is_enabled: true, design: { logo_url: '', background_color: '#171717', foreground_color: '#ffffff', label_color: '#cccccc', description: 'Карта лояльности' }, created_at: ago(30), updated_at: ago(10) }],
    stats: { total_passes: 156, apple_passes: 98, google_passes: 58, active_passes: 142 },
  },

  rfm: {
    dashboard: {
      segments: [
        { segment: 'vip', client_count: 52, percentage: 14.4, avg_check: 3200, total_check: 166400 },
        { segment: 'regular', client_count: 98, percentage: 27.2, avg_check: 1500, total_check: 147000 },
        { segment: 'promising', client_count: 72, percentage: 20.0, avg_check: 800, total_check: 57600 },
        { segment: 'new', client_count: 52, percentage: 14.4, avg_check: 600, total_check: 31200 },
        { segment: 'churn_risk', client_count: 38, percentage: 10.6, avg_check: 1200, total_check: 45600 },
        { segment: 'rare_valuable', client_count: 26, percentage: 7.2, avg_check: 4500, total_check: 117000 },
        { segment: 'lost', client_count: 22, percentage: 6.1, avg_check: 900, total_check: 19800 },
      ],
      trends: (() => {
        const segs = ['new', 'promising', 'regular', 'vip', 'rare_valuable', 'churn_risk', 'lost']
        const base: Record<string, number> = { new: 52, promising: 72, regular: 98, vip: 52, rare_valuable: 26, churn_risk: 38, lost: 22 }
        const result: any[] = []
        for (let d = 29; d >= 0; d--) {
          const date = ago(d)
          for (const seg of segs) {
            result.push({ id: result.length + 1, org_id: 1, segment: seg, client_count: base[seg] + Math.round(Math.sin(d * 0.3 + segs.indexOf(seg)) * 5), calculated_at: date })
          }
        }
        return result
      })(),
      config: { id: 1, org_id: 1, period_days: 90, recalc_interval: 86400, last_calc_at: ago(0.5), clients_processed: 360, active_template_type: 'preset', active_template_key: 'coffeegng' },
    },
    config: { id: 1, org_id: 1, period_days: 90, recalc_interval: 86400, last_calc_at: ago(0.5), clients_processed: 360, active_template_type: 'preset', active_template_key: 'coffeegng' },
  },

  marketplace: {
    products: [
      { id: 1, org_id: 1, name: 'Фирменная кружка', description: 'Керамическая кружка с логотипом', image_url: '', price_points: 500, stock: 20, is_active: true, sort_order: 1, created_at: ago(30), updated_at: ago(10) },
      { id: 2, org_id: 1, name: 'Пакет кофе 250г', description: 'Свежеобжаренный кофе', image_url: '', price_points: 800, stock: 50, is_active: true, sort_order: 2, created_at: ago(20), updated_at: ago(5) },
    ],
    orders: [] as any[],
    stats: { total_products: 2, active_products: 2, total_orders: 15, total_spent_points: 12500 },
  },
}

// ---------------------------------------------------------------------------
// Route matcher
// ---------------------------------------------------------------------------

type Handler = (params: { method: string; url: string; id?: string; data?: any; query?: any }) => any

const routes: [RegExp, Handler][] = [
  // Auth
  [/^\/auth\/login$/, () => ({ user: store.profile, tokens: { access_token: 'mock-token', refresh_token: 'mock-refresh', expires_in: 3600 } })],
  [/^\/auth\/refresh$/, () => ({ access_token: 'mock-token', refresh_token: 'mock-refresh', expires_in: 3600 })],
  [/^\/auth\/logout$/, () => ({})],

  // Profile / Account
  [/^\/account\/profile$/, ({ method, data }) => {
    if (method === 'patch' || method === 'put') Object.assign(store.profile, data)
    return store.profile
  }],
  [/^\/account\/change-email$/, () => ({})],
  [/^\/account\/change-phone$/, () => ({})],
  [/^\/account\/change-password$/, () => ({})],
  [/^\/account\/billing-details$/, ({ method, data }) => {
    if (method === 'patch' || method === 'put') Object.assign(store.billingDetails, data)
    return store.billingDetails
  }],

  // Bots
  [/^\/bots\/(\d+)\/settings$/, ({ method, id, data }) => {
    const bot = store.bots.find(b => b.id === +id!)
    if (method === 'patch' && bot) Object.assign(bot.settings, data)
    return bot?.settings ?? {}
  }],
  [/^\/bots\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.bots.findIndex(b => b.id === +id!)
    if (method === 'delete') { store.bots.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.bots[idx], data)
    return store.bots[idx]
  }],
  [/^\/bots$/, ({ method, data }) => {
    if (method === 'post') {
      const bot = { id: nextId(), org_id: 1, status: 'active', settings: { modules: [], buttons: [], registration_form: [], welcome_message: '' }, created_at: now, updated_at: now, client_count: 0, ...data }
      store.bots.push(bot)
      return bot
    }
    return store.bots
  }],

  // POS
  [/^\/pos\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.pos.findIndex(p => p.id === +id!)
    if (method === 'delete') { store.pos.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.pos[idx], data)
    return store.pos[idx]
  }],
  [/^\/pos$/, ({ method, data }) => {
    if (method === 'post') {
      const pos = { id: nextId(), org_id: 1, is_active: true, created_at: now, updated_at: now, ...data }
      store.pos.push(pos)
      return pos
    }
    return store.pos
  }],

  // Loyalty
  [/^\/loyalty\/programs\/(\d+)\/levels\/(\d+)$/, ({ method }) => {
    if (method === 'delete') return {}
    return {}
  }],
  [/^\/loyalty\/programs\/(\d+)\/levels$/, ({ method, data }) => {
    if (method === 'post' || method === 'put') return { id: nextId(), ...data }
    return []
  }],
  [/^\/loyalty\/programs\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.loyaltyPrograms.findIndex(p => p.id === +id!)
    if (method === 'patch' && idx >= 0) Object.assign(store.loyaltyPrograms[idx], data)
    return store.loyaltyPrograms[idx]
  }],
  [/^\/loyalty\/programs$/, ({ method, data }) => {
    if (method === 'post') {
      const prog = { id: nextId(), org_id: 1, is_active: true, created_at: now, updated_at: now, levels: [], ...data }
      store.loyaltyPrograms.push(prog)
      return prog
    }
    return store.loyaltyPrograms
  }],

  // Dashboard
  [/^\/dashboard\/widgets$/, ({ query }) => computeDashboardWidgets(store.transactions, store.clients, query)],
  [/^\/dashboard\/charts$/, ({ query }) => computeDashboardCharts(store.transactions, store.clients, query)],
  [/^\/dashboard\/sales$/, ({ query }) => {
    const s = computeSalesAnalytics(store.transactions, query)
    return { revenue: s.metrics.total_amount, avg_check: s.metrics.avg_amount, tx_count: s.metrics.transaction_count, loyalty_avg: s.comparison.participants_avg_amount, non_loyalty_avg: s.comparison.non_participants_avg_amount }
  }],

  // Analytics
  [/^\/analytics\/sales$/, ({ query }) => computeSalesAnalytics(store.transactions, query)],
  [/^\/analytics\/loyalty$/, ({ query }) => computeLoyaltyAnalytics(store.transactions, store.clients, query)],
  [/^\/analytics\/campaigns$/, ({ query }) => computeCampaignAnalytics(store.campaigns, query)],

  // Clients
  [/^\/clients\/stats$/, () => ({ total_clients: store.clients.length, total_balance: 324000, new_this_month: 47, active_this_week: 85 })],
  [/^\/clients\/count$/, () => ({ count: store.clients.length })],
  [/^\/clients\/(\d+)\/tags$/, ({ id, data }) => {
    const c = store.clients.find(cl => cl.id === +id!)
    if (c && data) c.tags = data.tags
    return c
  }],
  [/^\/clients\/(\d+)$/, ({ id }) => store.clients.find(c => c.id === +id!)],
  [/^\/clients$/, ({ query }) => {
    let items = [...store.clients]
    if (query?.bot_id) items = items.filter(c => c.bot_id === +query.bot_id)
    if (query?.search) {
      const s = query.search.toLowerCase()
      items = items.filter(c => c.first_name.toLowerCase().includes(s) || c.last_name.toLowerCase().includes(s) || c.phone.includes(s))
    }
    const limit = +(query?.limit ?? 20)
    const offset = +(query?.offset ?? 0)
    return { items: items.slice(offset, offset + limit), total: items.length }
  }],

  // File upload
  [/^\/files\/upload$/, () => ({ url: '/revisitr/storage/mock-file-' + Date.now() + '.jpg' })],

  // Campaigns
  [/^\/campaigns\/scenarios\/templates$/, () => []],
  [/^\/campaigns\/templates$/, () => []],
  [/^\/campaigns\/(\d+)\/analytics$/, () => ({ total: 300, sent: 295, failed: 5, clicked: 87, click_rate: 29.5 })],
  [/^\/campaigns\/(\d+)\/send$/, ({ id }) => {
    const c = store.campaigns.find(c => c.id === +id!)
    if (c) c.status = 'sending'
    return c
  }],
  [/^\/campaigns\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.campaigns.findIndex(c => c.id === +id!)
    if (method === 'delete') { store.campaigns.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.campaigns[idx], data)
    return store.campaigns[idx]
  }],
  [/^\/campaigns\/preview$/, () => ({ count: randomBetween(50, 300) })],
  [/^\/campaigns$/, ({ method, data, query }) => {
    if (method === 'post') {
      const c = { id: nextId(), org_id: 1, status: 'draft', stats: { total: 0, sent: 0, failed: 0 }, created_at: now, updated_at: now, ...data }
      store.campaigns.push(c)
      return c
    }
    const limit = +(query?.limit ?? 20)
    const offset = +(query?.offset ?? 0)
    return { items: store.campaigns.slice(offset, offset + limit), total: store.campaigns.length }
  }],

  // Auto scenarios
  [/^\/campaigns\/scenarios\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.campaigns.findIndex(c => c.id === +id! && c.type === 'auto')
    if (method === 'delete') { if (idx >= 0) store.campaigns.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.campaigns[idx], data)
    return idx >= 0 ? store.campaigns[idx] : {}
  }],
  [/^\/campaigns\/scenarios$/, ({ method, data }) => {
    if (method === 'post') {
      const s = { id: nextId(), org_id: 1, bot_id: 1, trigger_type: 'inactive_days', trigger_config: { days: 7 }, message: '', actions: [], timing: {}, conditions: {}, is_template: false, is_active: true, created_at: now, updated_at: now, ...data }
      return s
    }
    return []
  }],

  // Promotions
  [/^\/promotions\/promo-codes\/validate$/, () => ({ valid: true, reason: '', promo: { code: 'TEST', discount_percent: 10, bonus_amount: 0 } })],
  [/^\/promotions\/(\d+)\/codes$/, () => []],
  [/^\/promotions\/promo-codes\/(.+)\/deactivate$/, () => ({})],
  [/^\/promotions\/promo-codes\/(.+)\/activate$/, () => ({})],
  [/^\/promotions\/promo-codes\/generate$/, () => ({ code: 'PROMO-' + randomBetween(1000, 9999) })],
  [/^\/promotions\/promo-codes\/analytics$/, () => []],
  [/^\/promotions\/promo-codes\/(\d+)$/, ({ method }) => { if (method === 'delete') return {}; return {} }],
  [/^\/promotions\/promo-codes$/, ({ method, data }) => {
    if (method === 'post') return { id: nextId(), ...data, usage_count: 0, active: true, created_at: now }
    return []
  }],
  [/^\/promotions\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.promotions.findIndex(p => p.id === +id!)
    if (method === 'delete') { store.promotions.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.promotions[idx], data)
    return store.promotions[idx]
  }],
  [/^\/promotions$/, ({ method, data }) => {
    if (method === 'post') {
      const p = { id: nextId(), org_id: 1, active: true, created_at: now, ...data }
      store.promotions.push(p)
      return p
    }
    return store.promotions
  }],

  // Integrations
  [/^\/integrations\/(\d+)\/sync$/, () => ({})],
  [/^\/integrations\/(\d+)\/test$/, () => ({ success: true })],
  [/^\/integrations\/(\d+)\/orders$/, () => []],
  [/^\/integrations\/(\d+)\/customers$/, () => []],
  [/^\/integrations\/(\d+)\/menu$/, () => ({ categories: [] })],
  [/^\/integrations\/(\d+)\/stats$/, () => ({ total_orders: 1250, total_revenue: 1875000, matched_clients: 180, unmatched_orders: 42 })],
  [/^\/integrations\/(\d+)\/aggregates$/, () => []],
  [/^\/integrations\/(\d+)\/sales$/, () => ({ revenue: 487500, avg_check: 1250, tx_count: 390, loyalty_avg: 1450, non_loyalty_avg: 980 })],
  [/^\/integrations\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.integrations.findIndex(i => i.id === +id!)
    if (method === 'delete') { store.integrations.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.integrations[idx], data)
    return store.integrations[idx]
  }],
  [/^\/integrations$/, ({ method, data }) => {
    if (method === 'post') {
      const i = { id: nextId(), org_id: 1, status: 'active', last_sync_at: now, created_at: now, updated_at: now, ...data }
      store.integrations.push(i)
      return i
    }
    return store.integrations
  }],

  // Menus
  [/^\/menus\/(\d+)\/categories$/, ({ method, data }) => {
    if (method === 'post') return { id: nextId(), items: [], ...data }
    return []
  }],
  [/^\/menus\/(\d+)\/items\/(\d+)$/, ({ data }) => ({ id: nextId(), ...data })],
  [/^\/menus\/(\d+)\/items$/, ({ method, data }) => {
    if (method === 'post') return { id: nextId(), ...data }
    return []
  }],
  [/^\/menus\/client-order-stats$/, () => ({ total_orders: 12, total_amount: 15600, avg_amount: 1300, last_order_at: ago(3), top_items: [{ name: 'Капучино', order_count: 8, total_qty: 12, total_sum: 4200 }] })],
  [/^\/menus\/bot-pos-locations$/, ({ method }) => {
    if (method === 'put') return {}
    return store.pos.map(p => p.id)
  }],
  [/^\/menus\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.menus.findIndex(m => m.id === +id!)
    if (method === 'delete') { store.menus.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.menus[idx], data)
    return store.menus[idx]
  }],
  [/^\/menus$/, ({ method, data }) => {
    if (method === 'post') {
      const m = { id: nextId(), org_id: 1, source: 'manual', categories: [], created_at: now, updated_at: now, ...data }
      store.menus.push(m)
      return m
    }
    return store.menus
  }],

  // Segments
  [/^\/segments\/predictions\/summary$/, () => ({ high_churn_count: 38, avg_churn_risk: 0.42, high_upsell_count: 45, total_predicted: 360 })],
  [/^\/segments\/predictions\/high-churn$/, () => store.clients.slice(0, 5).map(c => ({ id: c.id, org_id: 1, client_id: c.id, churn_risk: 0.7 + Math.random() * 0.3, upsell_score: Math.random(), predicted_value: randomBetween(1000, 5000), factors: { days_since_last_visit: randomBetween(30, 90), visit_trend: 'declining', spend_trend: 'declining', avg_check: randomBetween(500, 2000), total_orders: randomBetween(3, 20), loyalty_level: 'Бронза' }, computed_at: ago(1) }))],
  [/^\/segments\/predictions$/, () => store.clients.slice(0, 10).map(c => ({ id: c.id, org_id: 1, client_id: c.id, churn_risk: Math.random(), upsell_score: Math.random(), predicted_value: randomBetween(500, 5000), factors: { days_since_last_visit: randomBetween(1, 90), visit_trend: ['increasing', 'stable', 'declining'][randomBetween(0, 2)], spend_trend: ['increasing', 'stable', 'declining'][randomBetween(0, 2)], avg_check: randomBetween(500, 3000), total_orders: randomBetween(1, 50), loyalty_level: ['Бронза', 'Серебро', 'Золото'][randomBetween(0, 2)] }, computed_at: ago(1) }))],
  [/^\/segments\/(\d+)\/clients$/, () => ({ items: store.clients.slice(0, 10), total: 10 })],
  [/^\/segments\/(\d+)\/recalculate$/, () => ({})],
  [/^\/segments\/(\d+)\/rules\/(\d+)$/, ({ method }) => { if (method === 'delete') return {}; return {} }],
  [/^\/segments\/(\d+)\/rules$/, ({ method, data }) => {
    if (method === 'post') return { id: nextId(), ...data, created_at: now }
    return []
  }],
  [/^\/segments\/preview-count$/, () => ({ count: randomBetween(20, 200) })],
  [/^\/segments\/preview$/, () => ({ count: randomBetween(20, 200) })],
  [/^\/segments\/(\d+)\/predictions$/, () => []],
  [/^\/segments\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.segments.findIndex(s => s.id === +id!)
    if (method === 'delete') { store.segments.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.segments[idx], data)
    return store.segments[idx]
  }],
  [/^\/segments$/, ({ method, data }) => {
    if (method === 'post') {
      const s = { id: nextId(), org_id: 1, client_count: 0, created_at: now, updated_at: now, ...data }
      store.segments.push(s)
      return s
    }
    return store.segments
  }],

  // RFM
  [/^\/rfm\/dashboard$/, () => store.rfm.dashboard],
  [/^\/rfm\/recalculate$/, () => ({})],
  [/^\/rfm\/config$/, ({ method, data }) => {
    if (method === 'patch' || method === 'put') Object.assign(store.rfm.config, data)
    return store.rfm.config
  }],
  [/^\/rfm\/templates$/, () => ({ templates: [
    { key: 'coffeegng', name: 'Кофейни и Grab&Go', description: 'Высокая частота, короткий чек. Для заведений, куда гости возвращаются часто — кофейни, кондитерские, точки с едой навынос.', r_thresholds: [3, 7, 14, 30], f_thresholds: [12, 8, 4, 2] },
    { key: 'qsr', name: 'Быстрое питание', description: 'Средне-высокая частота. Для фастфуда, столовых, пекарен — заведений с быстрым обслуживанием и средним чеком.', r_thresholds: [5, 10, 21, 45], f_thresholds: [9, 6, 3, 2] },
    { key: 'tsr', name: 'Кафе и рестораны', description: 'Средняя частота, более высокий чек. Для ресторанов с посадкой, кафе — заведений, куда гости приходят посидеть.', r_thresholds: [10, 21, 45, 90], f_thresholds: [6, 4, 3, 2] },
    { key: 'bar', name: 'Бары и пабы', description: 'Вечерний формат, событийные визиты. Для баров, пабов, караоке — заведений с вечерним и выходным трафиком.', r_thresholds: [7, 21, 45, 75], f_thresholds: [8, 5, 3, 2] },
  ] })],
  [/^\/rfm\/template$/, ({ method, data }) => {
    const allTemplates: Record<string, any> = {
      coffeegng: { key: 'coffeegng', name: 'Кофейни и Grab&Go', description: 'Высокая частота, короткий чек.', r_thresholds: [3, 7, 14, 30], f_thresholds: [12, 8, 4, 2] },
      qsr: { key: 'qsr', name: 'Быстрое питание', description: 'Средне-высокая частота.', r_thresholds: [5, 10, 21, 45], f_thresholds: [9, 6, 3, 2] },
      tsr: { key: 'tsr', name: 'Кафе и рестораны', description: 'Средняя частота, более высокий чек.', r_thresholds: [10, 21, 45, 90], f_thresholds: [6, 4, 3, 2] },
      bar: { key: 'bar', name: 'Бары и пабы', description: 'Вечерний формат, событийные визиты.', r_thresholds: [7, 21, 45, 75], f_thresholds: [8, 5, 3, 2] },
    }
    if (method === 'put' || method === 'post') {
      if (data?.template_key) {
        store.rfm.config.active_template_key = data.template_key
        store.rfm.config.active_template_type = 'preset'
      } else if (data?.template_type === 'custom') {
        store.rfm.config.active_template_type = 'custom'
        store.rfm.config.active_template_key = 'custom'
      }
    }
    const key = store.rfm.config.active_template_key
    return { active_template_type: store.rfm.config.active_template_type, active_template_key: key, template: allTemplates[key] ?? allTemplates.coffeegng, message: 'ok' }
  }],
  [/^\/rfm\/set-template$/, ({ data }) => {
    if (data?.template_key) {
      store.rfm.config.active_template_key = data.template_key
      store.rfm.config.active_template_type = 'preset'
    } else if (data?.template_type === 'custom') {
      store.rfm.config.active_template_type = 'custom'
      store.rfm.config.active_template_key = 'custom'
    }
    return {}
  }],
  [/^\/rfm\/onboarding-questions$/, () => []],
  [/^\/rfm\/recommend-template$/, () => ({ recommended: 'coffeegng', alternative: 'tsr', all_scores: {} })],
  [/^\/rfm\/segments\/(.+)\/clients$/, ({ id, query }) => {
    const seg = id ?? ''
    const segClients = store.clients.filter(c => c.rfm_segment === seg)
    const page = +(query?.page ?? 1)
    const perPage = +(query?.per_page ?? 20)
    const offset = (page - 1) * perPage
    const segLabels: Record<string, string> = { vip: 'VIP / Ядро', regular: 'Регулярные', promising: 'Перспективные', churn_risk: 'На грани оттока', new: 'Новые', rare_valuable: 'Редкие, но ценные', lost: 'Потерянные' }
    const rfmScores: Record<string, [number, number, number]> = { vip: [5, 5, 5], regular: [4, 4, 3], promising: [4, 2, 2], churn_risk: [2, 3, 4], new: [5, 1, 1], rare_valuable: [2, 1, 5], lost: [1, 1, 2] }
    const scores = rfmScores[seg] ?? [3, 3, 3]
    const clients = segClients.slice(offset, offset + perPage).map(c => ({
      id: c.id, first_name: c.first_name, last_name: c.last_name, phone: c.phone,
      r_score: scores[0] + (c.id % 2 === 0 ? 0 : -1 > 0 ? -1 : 0),
      f_score: scores[1],
      m_score: scores[2],
      last_visit_date: ago(c.id % 30),
      monetary_sum: c.total_purchases,
      frequency_count: c.purchase_count,
      total_visits_lifetime: c.purchase_count,
    }))
    return { segment: seg, segment_name: segLabels[seg] ?? seg, total: segClients.length, page, per_page: perPage, clients }
  }],

  // Marketplace
  [/^\/marketplace\/products\/(\d+)$/, ({ method, id, data }) => {
    const idx = store.marketplace.products.findIndex(p => p.id === +id!)
    if (method === 'delete') { store.marketplace.products.splice(idx, 1); return {} }
    if (method === 'patch' && idx >= 0) Object.assign(store.marketplace.products[idx], data)
    return store.marketplace.products[idx]
  }],
  [/^\/marketplace\/products$/, ({ method, data }) => {
    if (method === 'post') {
      const p = { id: nextId(), org_id: 1, is_active: true, sort_order: 1, created_at: now, updated_at: now, ...data }
      store.marketplace.products.push(p)
      return p
    }
    return store.marketplace.products
  }],
  [/^\/marketplace\/orders\/(\d+)$/, ({ method, id, data }) => {
    const o = store.marketplace.orders.find(o => o.id === +id!)
    if (method === 'patch' && o) Object.assign(o, data)
    return o
  }],
  [/^\/marketplace\/orders$/, ({ method, data }) => {
    if (method === 'post') {
      const o = { id: nextId(), org_id: 1, status: 'pending', total_points: 0, items: [], created_at: now, updated_at: now, ...data }
      store.marketplace.orders.push(o)
      return o
    }
    return store.marketplace.orders
  }],
  [/^\/marketplace\/stats$/, () => store.marketplace.stats],

  // Wallet
  [/^\/wallet\/configs\/(\d+)$/, ({ method, data }) => {
    if (method === 'patch') Object.assign(store.wallet.configs[0], data)
    if (method === 'delete') return {}
    return store.wallet.configs[0]
  }],
  [/^\/wallet\/configs$/, ({ method, data }) => {
    if (method === 'post') {
      const c = { id: nextId(), org_id: 1, is_enabled: true, created_at: now, updated_at: now, ...data }
      store.wallet.configs.push(c)
      return c
    }
    return store.wallet.configs
  }],
  [/^\/wallet\/passes$/, () => []],
  [/^\/wallet\/issue$/, () => ({ id: nextId(), status: 'active', created_at: now })],
  [/^\/wallet\/stats$/, () => store.wallet.stats],

  // Billing
  [/^\/billing\/tariffs$/, () => store.billing.tariffs],
  [/^\/billing\/subscription$/, ({ method, data }) => {
    if (method === 'post' || method === 'patch') return store.billing.subscription
    if (method === 'delete') return {}
    return store.billing.subscription
  }],
  [/^\/billing\/invoices\/(\d+)$/, ({ id }) => store.billing.invoices.find(i => i.id === +id!)],
  [/^\/billing\/invoices$/, () => store.billing.invoices],

  // Onboarding
  [/^\/onboarding\/state$/, () => store.onboarding],
  [/^\/onboarding\/step$/, ({ data }) => {
    if (data?.step) store.onboarding.onboarding_state.steps[data.step] = { completed: data.completed ?? false, skipped: data.skipped ?? false, entity_id: data.entity_id }
    return store.onboarding
  }],
  [/^\/onboarding\/complete$/, () => { store.onboarding.onboarding_completed = true; return store.onboarding }],
  [/^\/onboarding\/reset$/, () => { store.onboarding.onboarding_completed = false; return store.onboarding }],

  // Campaign templates
  [/^\/campaign-templates\/(\d+)$/, () => ({})],
  [/^\/campaign-templates$/, () => []],

  // AB tests
  [/^\/campaigns\/(\d+)\/ab-test$/, () => ({})],
  [/^\/campaigns\/(\d+)\/variants$/, () => []],
  [/^\/campaigns\/(\d+)\/ab-results$/, () => ({ campaign_id: 1, variants: [], winner_id: null })],
]

// ---------------------------------------------------------------------------
// Installer
// ---------------------------------------------------------------------------

function matchRoute(url: string, method: string, data?: any, params?: any) {
  const cleanUrl = url.replace(/\?.*$/, '')
  for (const [pattern, handler] of routes) {
    const match = cleanUrl.match(pattern)
    if (match) {
      return handler({ method, url: cleanUrl, id: match[1], data, query: params })
    }
  }
  console.warn(`[mock-api] Unmatched: ${method.toUpperCase()} ${url}`)
  return {}
}

export function installMockApi(instance: AxiosInstance) {
  instance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const method = (config.method ?? 'get').toLowerCase()
    const url = config.url ?? ''
    const data = config.data ? (typeof config.data === 'string' ? JSON.parse(config.data) : config.data) : undefined
    const result = matchRoute(url, method, data, config.params)

    // Create a fake response by using adapter override
    const fakeResponse = {
      data: result,
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
    }

    // Abort the real request and return mock data
    return Promise.reject({
      __MOCK__: true,
      response: fakeResponse,
    }) as any
  })

  // Intercept the "error" and extract mock response
  instance.interceptors.response.use(
    (response) => response,
    (error) => {
      if (error?.__MOCK__) {
        return Promise.resolve(error.response)
      }
      return Promise.reject(error)
    },
  )

  console.log('%c[mock-api] Mock API enabled — all requests return fake data', 'color: #EF3219; font-weight: bold')
}
