import { cn } from '@/lib/utils'
import type { ReactNode } from 'react'

interface PhoneFrameProps {
  children: ReactNode
  className?: string
}

export function PhoneFrame({ children, className }: PhoneFrameProps) {
  return (
    <div className={cn('tg-phone-frame w-full max-w-[360px] bg-black', className)}>
      {/* Status bar */}
      <div className="flex items-center justify-between px-6 py-2 bg-[#17212B]">
        <span className="text-white text-xs font-medium">9:41</span>
        <div className="flex items-center gap-1">
          <div className="w-4 h-2.5 border border-white/60 rounded-sm relative">
            <div className="absolute inset-[1px] right-[2px] bg-white/60 rounded-[1px]" />
          </div>
        </div>
      </div>
      {children}
    </div>
  )
}
