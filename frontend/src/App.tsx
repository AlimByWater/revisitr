import { RouterProvider } from 'react-router-dom'
import { SWRConfig } from 'swr'
import { router } from './router'
import { ThemeProvider } from './contexts/ThemeContext'

export function App() {
  return (
    <ThemeProvider>
      <SWRConfig
        value={{
          revalidateOnFocus: false,
          dedupingInterval: 5 * 60 * 1000,
          errorRetryCount: 1,
        }}
      >
        <RouterProvider router={router} />
      </SWRConfig>
    </ThemeProvider>
  )
}
