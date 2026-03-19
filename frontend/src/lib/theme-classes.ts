/**
 * Theme-aware class sets for common UI patterns.
 * Use these with cn() to apply the right classes for the current theme.
 *
 * Instead of:   "bg-white border-surface-border"
 * Use:          themeClasses.card
 *
 * These classes work via CSS variables — they automatically adapt
 * when data-theme changes on the root element.
 */

/** Card background with border and shadow — adapts to glassmorphism in Aurora */
export const card = 'bg-[var(--color-surface-card)] border border-[var(--color-surface-border)] shadow-card'
export const cardHover = 'hover:shadow-card-hover hover:-translate-y-0.5'

/** Filter bar and controls wrapper */
export const controlBar = 'bg-[var(--color-surface-card)] border border-[var(--color-surface-border)]'

/** Input/select fields */
export const input = 'bg-[var(--color-surface-card)] border border-[var(--color-surface-border)] text-[var(--color-text-primary)] outline-none'

/** Active toggle button */
export const toggleActive = 'bg-[var(--color-text-primary)] text-[var(--color-surface)]'
export const toggleInactive = 'text-[var(--color-text-secondary)]'

/** Page heading */
export const heading = 'font-display text-3xl font-bold text-[var(--color-text-primary)] tracking-tight'

/** Subtitle text */
export const subtitle = 'text-sm text-[var(--color-text-muted)]'

/** Body text */
export const textPrimary = 'text-[var(--color-text-primary)]'
export const textSecondary = 'text-[var(--color-text-secondary)]'
export const textMuted = 'text-[var(--color-text-muted)]'
