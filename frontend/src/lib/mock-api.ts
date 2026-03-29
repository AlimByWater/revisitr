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

let idCounter = 100

function nextId() { return ++idCounter }

function randomBetween(min: number, max: number) {
  return Math.round(min + Math.random() * (max - min))
}

function chartPoints(days: number, key = 'date'): Record<string, string | number>[] {
  return Array.from({ length: days }, (_, i) => ({
    [key]: ago(days - 1 - i).slice(0, 10),
    value: randomBetween(5000, 50000),
  }))
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
    { id: 2, org_id: 1, name: 'Приведи друга', type: 'bonus', conditions: { min_amount: 0, segment_id: null, min_visits: 1 }, result: { discount_percent: 0, bonus_amount: 300, tag_add: 'referral', campaign_id: null }, recurrence: '', starts_at: ago(60), ends_at: ago(-90), usage_limit: 1, combinable: true, active: true, created_at: ago(60) },
  ] as any[],

  clients: Array.from({ length: 360 }, (_, i) => ({
    id: i + 1, bot_id: (i % 2) + 1, telegram_id: 100000 + i,
    username: `user_${i + 1}`, first_name: ['Алексей', 'Мария', 'Дмитрий', 'Анна', 'Сергей', 'Елена', 'Иван', 'Ольга', 'Павел', 'Наталья'][i % 10],
    last_name: ['Иванов', 'Петрова', 'Сидоров', 'Козлова', 'Смирнов', 'Новикова', 'Морозов', 'Волкова', 'Лебедев', 'Соколова'][i % 10],
    phone: `+7 9${10 + (i % 90)} ${100 + (i % 900)}-${10 + (i % 90)}-${10 + (i % 90)}`,
    gender: i % 2 === 0 ? 'male' : 'female', birth_date: `199${i % 10}-0${(i % 9) + 1}-${10 + (i % 20)}`,
    city: ['Москва', 'Санкт-Петербург', 'Казань'][i % 3], os: i % 3 === 0 ? 'iOS' : 'Android',
    tags: i % 4 === 0 ? ['vip'] : [], registered_at: ago(1 + (i % 180)),
    bot_name: i % 2 === 0 ? 'Кофейня на Маросейке' : 'Бар «Хмель»',
    loyalty_balance: 100 + (i * 13) % 5000, loyalty_level: ['Бронза', 'Серебро', 'Золото'][i % 3],
    total_purchases: 500 + (i * 277) % 100000, purchase_count: 1 + (i % 50),
    rfm_segment: ['vip', 'regular', 'promising', 'churn_risk', 'new', 'rare_valuable', 'lost'][i % 7],
  })),

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
      trends: chartPoints(30).map((p, i) => ({ ...p, segment: ['vip', 'regular', 'new'][i % 3] })),
      config: { id: 1, org_id: 1, period_days: 90, recalc_interval: 86400, last_calc_at: ago(0.5), clients_processed: 360, active_template_type: 'preset', active_template_key: 'default' },
    },
    config: { id: 1, org_id: 1, period_days: 90, recalc_interval: 86400, last_calc_at: ago(0.5), clients_processed: 360, active_template_type: 'preset', active_template_key: 'default' },
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
  [/^\/dashboard\/widgets$/, () => ({
    revenue: { value: 1112500, previous: 980000, trend: 13.5 },
    avg_check: { value: 1250, previous: 1180, trend: 5.9 },
    new_clients: { value: 47, previous: 38, trend: 23.7 },
    active_clients: { value: 245, previous: 220, trend: 11.4 },
  })],
  [/^\/dashboard\/charts$/, () => ({
    revenue: chartPoints(30),
    new_clients: chartPoints(30).map(p => ({ ...p, value: randomBetween(1, 15) })),
  })],
  [/^\/dashboard\/sales$/, () => ({
    revenue: 1112500, avg_check: 1250, tx_count: 890,
    loyalty_avg: 1450, non_loyalty_avg: 980,
  })],

  // Analytics
  [/^\/analytics\/sales$/, () => ({
    metrics: { transaction_count: 890, unique_clients: 245, total_amount: 1112500, avg_amount: 1250, buy_frequency: 3.6 },
    charts: { revenue: chartPoints(30, 'day'), transactions: chartPoints(30, 'day').map(p => ({ ...p, value: randomBetween(5, 30) })) },
    comparison: { participants_avg_amount: 1450, non_participants_avg_amount: 980 },
  })],
  [/^\/analytics\/loyalty$/, () => ({
    new_clients: 47, active_clients: 245, bonus_earned: 84500, bonus_spent: 51000,
    demographics: { by_gender: [{ label: 'Мужской', value: 55, percent: 55 }, { label: 'Женский', value: 45, percent: 45 }], by_age_group: [{ label: '18-25', value: 20, percent: 20 }, { label: '26-35', value: 40, percent: 40 }, { label: '36-45', value: 25, percent: 25 }, { label: '46+', value: 15, percent: 15 }], by_os: [{ label: 'iOS', value: 42, percent: 42 }, { label: 'Android', value: 58, percent: 58 }], loyalty_percent: 68 },
    bot_funnel: [{ step: 'Открыли бота', count: 520, percent: 100 }, { step: 'Зарегистрировались', count: 360, percent: 69.2 }, { step: 'Первая покупка', count: 245, percent: 47.1 }, { step: 'Повторная покупка', count: 156, percent: 30 }],
  })],
  [/^\/analytics\/campaigns$/, () => ({
    total_sent: 720, total_opened: 504, open_rate: 70.0, conversions: 144, conv_rate: 20.0,
    by_campaign: store.campaigns.map(c => ({ campaign_id: c.id, campaign_name: c.name, sent: c.stats.sent, open_rate: randomBetween(50, 85), conversions: randomBetween(10, 60) })),
  })],

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

  // Campaigns
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
  [/^\/campaigns\/preview-audience$/, () => ({ count: randomBetween(50, 300) })],
  [/^\/campaigns$/, ({ method, data }) => {
    if (method === 'post') {
      const c = { id: nextId(), org_id: 1, status: 'draft', stats: { total: 0, sent: 0, failed: 0 }, created_at: now, updated_at: now, ...data }
      store.campaigns.push(c)
      return c
    }
    return store.campaigns
  }],

  // Auto scenarios
  [/^\/scenarios\/(\d+)$/, ({ method }) => { if (method === 'delete') return {}; return {} }],
  [/^\/scenarios$/, () => []],

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
    { key: 'default', name: 'По умолчанию', description: 'Стандартная RFM-модель', r_thresholds: [30, 90], f_thresholds: [2, 5] },
    { key: 'horeca', name: 'HoReCa', description: 'Адаптирована для общепита', r_thresholds: [14, 60], f_thresholds: [3, 8] },
  ] })],
  [/^\/rfm\/template$/, () => ({ active_template_type: 'preset', active_template_key: 'default', template: { key: 'default', name: 'По умолчанию', description: 'Стандартная RFM-модель', r_thresholds: [30, 90], f_thresholds: [2, 5] } })],
  [/^\/rfm\/set-template$/, () => ({})],
  [/^\/rfm\/onboarding-questions$/, () => []],
  [/^\/rfm\/recommend-template$/, () => ({ recommended: 'horeca', alternative: 'default', all_scores: {} })],
  [/^\/rfm\/segments\/(.+)\/clients$/, () => ({ segment: 'vip', segment_name: 'VIP', total: 45, page: 1, per_page: 20, clients: [] })],

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
