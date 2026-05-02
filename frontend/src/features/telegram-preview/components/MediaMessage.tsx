import { useState } from 'react'
import { ImageIcon, Play } from 'lucide-react'
import { renderTextWithEmoji } from './renderEmoji'
import { PREVIEW_MESSAGE_TIME } from './previewConstants'
import tailIncoming from '../assets/tail-incoming.svg'

interface MediaMessageProps {
  type: 'photo' | 'video' | 'animation'
  mediaUrl?: string
  caption?: string
  showTail?: boolean
  containerClassName?: string
  mediaClassName?: string
}

export function MediaMessage({
  type,
  mediaUrl,
  caption,
  showTail = false,
  containerClassName,
  mediaClassName,
}: MediaMessageProps) {
  const [hasError, setHasError] = useState(false)

  const placeholder = (
    <div className="flex h-[180px] w-full items-center justify-center bg-gray-100">
      <ImageIcon className="h-10 w-10 text-gray-400" />
    </div>
  )

  return (
    <div className={`flex justify-start ${containerClassName ?? 'pl-[4px]'}`}>
      <div
        className={`relative max-w-[312px] overflow-hidden bg-white ${mediaClassName ?? ''}`}
        style={{
          borderRadius: caption ? '18px 18px 18px 4px' : '18px',
          padding: caption ? '1px' : '0',
        }}
      >
        {/* Media area */}
        <div
          className="overflow-hidden"
          style={{
            borderRadius: caption ? '17px 17px 0 0' : '18px',
          }}
        >
          {mediaUrl && !hasError ? (
            type === 'video' ? (
              <div className="relative">
                <img
                  src={mediaUrl}
                  alt=""
                  loading="lazy"
                  decoding="async"
                  className="max-h-[280px] w-full object-cover"
                  onError={() => setHasError(true)}
                />
                <div className="absolute inset-0 flex items-center justify-center">
                  <div className="flex h-12 w-12 items-center justify-center rounded-full bg-black/40">
                    <Play className="ml-0.5 h-6 w-6 fill-white text-white" />
                  </div>
                </div>
              </div>
            ) : (
              <img
                src={mediaUrl}
                alt=""
                loading="lazy"
                decoding="async"
                className="max-h-[280px] w-full object-cover"
                onError={() => setHasError(true)}
              />
            )
          ) : (
            placeholder
          )}
        </div>

        {/* Caption */}
        {caption && (
          <div className="relative px-[10px] pb-[18px] pt-[6px]">
            <span className="whitespace-pre-wrap text-[13px] font-normal leading-[15px] tracking-[-0.16px] text-black">
              {renderTextWithEmoji(caption)}
            </span>
            <span className="absolute bottom-[6px] right-[8px] text-[10px] leading-none text-black/40">
              {PREVIEW_MESSAGE_TIME}
            </span>
          </div>
        )}

        {/* Timestamp overlay on media (no caption) */}
        {!caption && (
          <span className="absolute bottom-[6px] right-[8px] rounded-full bg-black/40 px-[6px] py-[1px] text-[11px] text-white/80">
            {PREVIEW_MESSAGE_TIME}
          </span>
        )}

        {/* Tail */}
        {showTail && (
          <img
            src={tailIncoming}
            alt=""
            className="tg-tail-incoming"
            draggable={false}
          />
        )}
      </div>
    </div>
  )
}
