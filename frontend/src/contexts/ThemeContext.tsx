import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'

export type Theme = 'default' | 'aurora'

interface ThemeContextType {
  theme: Theme
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
}

const ThemeContext = createContext<ThemeContextType | null>(null)

export function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(() => {
    const saved = localStorage.getItem('revisitr-theme')
    return (saved === 'aurora' ? 'aurora' : 'default') as Theme
  })

  const setTheme = (t: Theme) => {
    setThemeState(t)
    localStorage.setItem('revisitr-theme', t)
  }

  const toggleTheme = () => {
    setTheme(theme === 'default' ? 'aurora' : 'default')
  }

  useEffect(() => {
    const root = document.documentElement
    root.setAttribute('data-theme', theme)
    return () => root.removeAttribute('data-theme')
  }, [theme])

  return (
    <ThemeContext.Provider value={{ theme, setTheme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const ctx = useContext(ThemeContext)
  if (!ctx) throw new Error('useTheme must be used within ThemeProvider')
  return ctx
}
