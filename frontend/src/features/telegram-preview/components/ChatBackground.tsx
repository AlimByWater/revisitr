import type { ReactNode } from 'react'

export function ChatBackground({ children }: { children: ReactNode }) {
  return (
    <div className="tg-chat-background relative flex-1 overflow-hidden">
      {children}
    </div>
  )
}
