import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: {
          DEFAULT: 'rgb(var(--color-accent-rgb) / <alpha-value>)',
          hover: 'rgb(var(--color-accent-hover-rgb) / <alpha-value>)',
          light: 'var(--color-accent-light)',
        },
        sidebar: {
          DEFAULT: 'var(--color-sidebar-bg)',
          hover: 'var(--color-sidebar-hover)',
          active: 'var(--color-sidebar-active)',
          muted: 'var(--color-sidebar-muted)',
          text: 'var(--color-sidebar-text)',
        },
        surface: {
          DEFAULT: 'var(--color-surface)',
          card: 'var(--color-surface-card)',
          border: 'var(--color-surface-border)',
        },
        th: {
          primary: 'var(--color-text-primary)',
          secondary: 'var(--color-text-secondary)',
          muted: 'var(--color-text-muted)',
        },
      },
      width: {
        sidebar: '312px',
      },
      fontFamily: {
        // base body text — currently Inter
        sans: [
          'Inter',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'sans-serif',
        ],
        // legacy serif (kept for fallback/future use). Use `font-display` for headings instead.
        serif: [
          'Playfair Display',
          'Georgia',
          'Cambria',
          'serif',
        ],
        // numbers, code, monospace fragments
        mono: [
          'JetBrains Mono',
          'Fira Code',
          'Consolas',
          'monospace',
        ],
        // semantic heading family — driven by CSS variable so we can swap back to a serif globally
        display: ['var(--font-display)', 'sans-serif'],
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
        xl: 'var(--radius-xl)',
        '2xl': 'var(--radius-2xl)',
      },
      boxShadow: {
        card: 'var(--shadow-card)',
        'card-hover': 'var(--shadow-card-hover)',
      },
    },
  },
  plugins: [],
}

export default config
