/**
 * Centralized selectors for E2E tests.
 * Prefer data-testid, fall back to role/text selectors.
 */

export const selectors = {
  // ─── Layout ─────────────────────────────────────────────────────
  sidebar: 'aside, [data-testid="sidebar"], nav',
  header: 'header, [data-testid="header"]',
  mainContent: 'main, [data-testid="main-content"], [role="main"]',

  // ─── Auth ───────────────────────────────────────────────────────
  loginForm: 'form',
  emailInput: 'input[name="email"], input[type="email"]',
  passwordInput: 'input[name="password"], input[type="password"]',
  submitButton: 'button[type="submit"]',

  // ─── Navigation ─────────────────────────────────────────────────
  sidebarLink: (text: string) => `nav a:has-text("${text}"), aside a:has-text("${text}")`,

  // ─── Common UI ──────────────────────────────────────────────────
  loadingSpinner: '[data-testid="loading"], .animate-spin, [role="progressbar"]',
  errorMessage: '[data-testid="error"], [role="alert"]',
  emptyState: '[data-testid="empty-state"]',
  modal: '[role="dialog"], [data-testid="modal"]',
  modalClose: '[data-testid="modal-close"], [role="dialog"] button:has-text("Закрыть")',
  confirmButton: 'button:has-text("Подтвердить"), button:has-text("Да"), button:has-text("Удалить")',
  cancelButton: 'button:has-text("Отмена"), button:has-text("Нет")',

  // ─── Tables ─────────────────────────────────────────────────────
  table: 'table, [role="table"]',
  tableRow: 'tbody tr, [role="row"]',
  tableHeader: 'thead th, [role="columnheader"]',

  // ─── Forms ──────────────────────────────────────────────────────
  nameInput: 'input[name="name"]',
  saveButton: 'button:has-text("Сохранить"), button:has-text("Создать"), button[type="submit"]',
  deleteButton: 'button:has-text("Удалить")',

  // ─── User menu ─────────────────────────────────────────────────
  userAvatar: '[data-testid="user-avatar"], [data-testid="user-menu"]',
  logoutButton: 'button:has-text("Выйти"), a:has-text("Выйти")',
  settingsLink: 'a:has-text("Настройки")',
} as const;

/**
 * Error text patterns that indicate page loading failure.
 */
export const errorPatterns = [
  'Ошибка загрузки',
  'Не удалось загрузить',
  'Something went wrong',
  'Unexpected error',
] as const;
