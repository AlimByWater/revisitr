import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'
import { StatusBar } from './StatusBar'
import bezelsUrl from '../assets/iphone-bezels.svg'

interface PhoneFrameProps {
  children: ReactNode
  className?: string
}

export function PhoneFrame({ children, className }: PhoneFrameProps) {
  return (
    <div
      className={cn('tg-root relative mx-auto', className)}
      style={{ width: 320, aspectRatio: '450 / 920' }}
    >
      {/* Screen — inset enough so gradient never bleeds past bezels */}
      <div
        className="tg-chat-background absolute flex flex-col overflow-hidden"
        style={{
          top: '16px',
          left: '17px',
          right: '17px',
          bottom: '16px',
          borderRadius: '38px',
        }}
      >
        <StatusBar />
        {children}
        <div className="absolute bottom-[6px] left-1/2 z-10 h-[3px] w-[80px] -translate-x-1/2 rounded-full bg-black/40" />
      </div>

      {/* Bezels overlay */}
      <img
        src={bezelsUrl}
        alt=""
        className="pointer-events-none absolute inset-0 z-20 h-full w-full select-none"
        draggable={false}
      />
    </div>
  )
}
