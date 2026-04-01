import { useState } from 'react'
import { cn } from '@/lib/utils'
import { ImageIcon, Play } from 'lucide-react'

interface MediaMessageProps {
  type: 'photo' | 'video' | 'animation'
  mediaUrl?: string
  caption?: string
  theme?: 'light' | 'dark'
  showTail?: boolean
}

export function MediaMessage({
  type,
  mediaUrl,
  caption,
  theme = 'light',
  showTail = false,
}: MediaMessageProps) {
  const isDark = theme === 'dark'
  const [hasError, setHasError] = useState(false)
  const now = new Date()
  const timestamp = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`

  const placeholder = (
    <div
      className={cn(
        'w-full h-[180px] flex items-center justify-center',
        isDark ? 'bg-[#1E2C3A]' : 'bg-gray-100'
      )}
    >
      <ImageIcon className="w-10 h-10 text-gray-400" />
    </div>
  )

  return (
    <div className="flex justify-start">
      <div
        className={cn(
          'tg-bubble',
          isDark && 'tg-bubble-dark',
          showTail && 'tg-bubble-tail',
          !caption && 'p-0'
        )}
      >
        <div className={cn('tg-bubble-media', !caption && 'rounded-[14px] !m-0')}>
          {mediaUrl && !hasError ? (
            type === 'video' ? (
              <div className="relative">
                <img
                  src={mediaUrl}
                  alt=""
                  className="w-full max-h-[280px] object-cover"
                  onError={() => setHasError(true)}
                />
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="w-12 h-12 rounded-full bg-black/40 flex items-center justify-center">
                    <Play className="w-6 h-6 text-white fill-white ml-0.5" />
                  </div>
                </div>
              </div>
            ) : (
              <img
                src={mediaUrl}
                alt=""
                className="w-full max-h-[280px] object-cover"
                onError={() => setHasError(true)}
              />
            )
          ) : (
            placeholder
          )}
        </div>
        {caption && (
          <div className="tg-bubble-caption">
            <span className="whitespace-pre-wrap">{caption}</span>
            <span className={cn('tg-timestamp', isDark && 'tg-timestamp-dark')}>
              {timestamp}
            </span>
          </div>
        )}
        {!caption && (
          <span
            className={cn(
              'tg-timestamp absolute bottom-2 right-2 bg-black/40 rounded-full px-1.5 text-white/80',
              'text-[10px]'
            )}
          >
            {timestamp}
          </span>
        )}
      </div>
    </div>
  )
}
