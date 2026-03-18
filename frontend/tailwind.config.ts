import type { Config } from 'tailwindcss'

const config: Config = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: {
          DEFAULT: '#E85D3A',
          hover: '#D4502F',
          light: '#FFF0EC',
        },
        sidebar: {
          DEFAULT: '#1A1A1A',
          hover: '#2A2A2A',
          active: '#333333',
          muted: '#888888',
        },
        surface: {
          DEFAULT: '#FAFAFA',
          card: '#FFFFFF',
          border: '#E5E5E5',
        },
      },
      width: {
        sidebar: '312px',
      },
      fontFamily: {
        sans: [
          'Inter',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'sans-serif',
        ],
      },
    },
  },
  plugins: [],
}

export default config
